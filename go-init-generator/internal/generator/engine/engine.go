package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/app"
	"go-init-gen/internal/generator/engine/generators/config"
	"go-init-gen/internal/generator/engine/generators/entity"
	"go-init-gen/internal/generator/engine/generators/features"
	"go-init-gen/internal/generator/engine/generators/model"
	"go-init-gen/internal/generator/engine/generators/repository"
	"go-init-gen/internal/generator/engine/generators/service"
	"go-init-gen/internal/generator/engine/generators/yaml"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

const (
	tmpSuffix = ".tmpl"
)

// Generator handles the code generation process
type Generator struct {
	renderer      Renderer
	templateDir   string
	debugArchives bool
	debugDir      string // Directory to save debug archives
}

// New creates a new generator instance
func New() *Generator {
	// Get template directory from environment or use default
	templateDir := os.Getenv("TEMPLATE_DIR")
	if templateDir == "" {
		templateDir = "../internal/generator/templates/microservices"
	}

	// Check if we should save debug archives
	debugArchives := os.Getenv("GENERATOR_SAVE_ARCHIVE_LOCALLY") == "true"

	// Set default debug directory
	debugDir := filepath.Join("internal", "generator", "debug_archives")

	return &Generator{
		renderer:      NewRenderer(templateDir),
		templateDir:   templateDir,
		debugArchives: debugArchives,
		debugDir:      debugDir,
	}
}

// Generate creates a template based on input data
func (g *Generator) Generate(ctx context.Context, template *eventdata.ProcessTemplate) ([]byte, error) {
	// Prepare template variables
	variables, _ := g.prepareTemplateVariables(&template.Data)

	// Load and filter template files
	files, err := g.loadTemplateFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to load template files: %w", err)
	}

	filesToGenerate := g.filterFiles(files, &template.Data)

	// Generate files
	generatedFiles, err := g.generateFiles(filesToGenerate, &template.Data, variables)
	if err != nil {
		return nil, fmt.Errorf("failed to generate files: %w", err)
	}

	// Create archive
	archiveBytes, err := g.createArchive(generatedFiles, template.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	return archiveBytes, nil
}

// prepareTemplateVariables prepares variables for template rendering
func (g *Generator) prepareTemplateVariables(data *eventdata.TemplateEventData) (map[string]interface{}, map[string]bool) {
	customizer := NewCustomizer()
	customizer.ProcessInput(data)

	variables := customizer.GetVariables()
	features := customizer.GetFeatureFlags()

	// Add features to variables for template usage
	variables["features"] = features

	return variables, features
}

// loadTemplateFiles loads all template files from the template directory
func (g *Generator) loadTemplateFiles() ([]TemplateFile, error) {
	var files []TemplateFile

	// Walk the template directory
	err := filepath.Walk(g.templateDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Calculate relative path
		relPath, err := filepath.Rel(g.templateDir, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %v", path, err)
		}

		// Normalize path separators
		relPath = filepath.ToSlash(relPath)

		// Determine if this is a template file
		isTemplate := strings.HasSuffix(relPath, tmpSuffix)

		// Remove .tmpl extension for target path
		targetPath := relPath
		if isTemplate {
			targetPath = strings.TrimSuffix(relPath, tmpSuffix)
		}

		// Determine code generation strategy
		strategy := GetFileStrategy(targetPath)

		// Create template file
		file := TemplateFile{
			Name:           relPath,
			Content:        string(content),
			CodeGeneration: strategy,
			UseAST:         strategy == StrategyASTGeneration || strategy == StrategyHybrid,
			TargetPath:     targetPath,
		}

		files = append(files, file)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

// FeatureFilterRules defines rules for including/excluding files based on features
var FeatureFilterRules = map[string]struct {
	// RequiredFeatures lists features that must be present for the file to be included
	RequiredFeatures []string
	// ExcludedFeatures lists features that, if present, cause the file to be excluded
	ExcludedFeatures []string
}{
	// gRPC-specific files
	"grpc":  {RequiredFeatures: []string{"hasGRPC"}, ExcludedFeatures: nil},
	"proto": {RequiredFeatures: []string{"hasGRPC"}, ExcludedFeatures: nil},

	// GraphQL-specific files
	"graphql": {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil},
	"gql":     {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil},
	"tools/":  {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil}, // tools directory is GraphQL specific (for gqlgen)

	// Database-specific files
	"repository": {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},
	"storage":    {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},
	"database":   {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},
	"model":      {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},
	"migration":  {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},
	"sql":        {RequiredFeatures: []string{"hasDatabase"}, ExcludedFeatures: nil},

	// HTTP/REST-specific files (excluded)
	"rest":       {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"http":       {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"handler":    {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"router":     {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"middleware": {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},

	// Unsupported database files (excluded)
	"mysql":   {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"mongodb": {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"sqlite":  {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
	"redis":   {RequiredFeatures: nil, ExcludedFeatures: []string{"all"}},
}

// filterFiles filters template files based on input features
func (g *Generator) filterFiles(files []TemplateFile, data *eventdata.TemplateEventData) []TemplateFile {
	// Use the centralized features detector
	fs := features.DetectFeatures(data)

	// Create a feature map for easy checking
	featuresMap := map[string]bool{
		"hasGRPC":     fs.HasGRPC,
		"hasGraphQL":  fs.HasGraphQL,
		"hasDatabase": fs.HasDatabase,
		"hasPostgres": fs.HasPostgres(),
		"hasMySQL":    fs.HasMySQL(),
		"hasMongoDB":  fs.HasMongoDB(),
		"hasRedis":    fs.HasRedis(),
		"hasREST":     fs.HasREST,
		"hasHTTP":     fs.HasHTTP,
		"hasKafka":    fs.HasKafka,
	}

	// Find no-db service template for special handling
	serviceNoDBFile := g.findNoDatabaseServiceFile(files, fs.HasDatabase)

	// Filter the files
	filteredFiles := make([]TemplateFile, 0)

	for _, file := range files {
		// Skip the no-db service template - we'll handle it separately
		if strings.HasSuffix(file.Name, "service/service_no_db.go.tmpl") {
			continue
		}

		// Special handling for no-database case: replace regular service with no-db service
		if !fs.HasDatabase && strings.HasSuffix(file.Name, "service/service.go.tmpl") && serviceNoDBFile != nil {
			filteredFiles = append(filteredFiles, *serviceNoDBFile)
			continue
		}

		// Always include base files
		if isBaseFile(file.Name) {
			filteredFiles = append(filteredFiles, file)
			continue
		}

		// Apply filter rules
		shouldInclude := g.shouldIncludeFile(file.Name, featuresMap, fs, data)

		// If the file passed all checks, include it
		if shouldInclude {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

// findNoDatabaseServiceFile finds the no-database service file for special handling
func (g *Generator) findNoDatabaseServiceFile(files []TemplateFile, hasDatabase bool) *TemplateFile {
	if hasDatabase {
		return nil
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name, "service/service_no_db.go.tmpl") {
			// Create a copy of the file with the correct target path
			noDBFile := file
			noDBFile.TargetPath = strings.Replace(noDBFile.TargetPath, "service_no_db.go", "service.go", 1)
			return &noDBFile
		}
	}
	return nil
}

// shouldIncludeFile determines if a file should be included based on feature rules
func (g *Generator) shouldIncludeFile(fileName string, featuresMap map[string]bool, fs *features.FeatureSet, data *eventdata.TemplateEventData) bool {
	// Check against each pattern in our filter rules
	for pattern, rule := range FeatureFilterRules {
		if strings.Contains(fileName, pattern) {
			// Check required features
			for _, requiredFeature := range rule.RequiredFeatures {
				if !featuresMap[requiredFeature] {
					return false
				}
			}

			// Check excluded features
			for _, excludedFeature := range rule.ExcludedFeatures {
				if excludedFeature == "all" || featuresMap[excludedFeature] {
					return false
				}
			}
		}
	}

	// Special handling for database type
	if fs.HasDatabase && isDatabaseFile(fileName) && !fs.HasPostgres() && strings.Contains(fileName, strings.ToLower(data.Database.Type)) {
		return false
	}

	return true
}

// isBaseFile checks if a file is a base file for any project
func isBaseFile(fileName string) bool {
	basePatterns := []string{
		"go.mod",
		"go.sum",
		"README.md",
		"Makefile",
		"Dockerfile",
		".gitignore",
		"cmd/main.go",
	}

	for _, pattern := range basePatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	// Note that config/config.go and internal/app/app.go are dynamic files
	// and will be processed specially during AST transformations

	return false
}

// isDatabaseFile checks if a file is related to a database
func isDatabaseFile(fileName string) bool {
	dbPatterns := []string{
		"database",
		"repository",
		"models",
		"migrations",
		"entity",
	}

	for _, pattern := range dbPatterns {
		if strings.Contains(fileName, pattern) {
			return true
		}
	}

	return false
}

// generateFiles generates the content for each file
func (g *Generator) generateFiles(files []TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (map[string][]byte, error) {
	generatedFiles := make(map[string][]byte)

	for _, file := range files {
		content, err := g.generateFileContent(file, data, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to generate content for %s: %w", file.Name, err)
		}

		generatedFiles[file.TargetPath] = []byte(content)
	}

	return generatedFiles, nil
}

// generateFileContent generates content for a single file based on its strategy
func (g *Generator) generateFileContent(file TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (string, error) {
	switch file.CodeGeneration {
	case StrategyTextTemplate:
		return g.renderer.RenderTemplateWithData(file.Name, file.Content, variables)

	case StrategyASTGeneration:
		return g.generateWithAST(file, data)

	case StrategyHybrid:
		return g.processHybridFile(file, data, variables)

	case StrategyRaw:
		return file.Content, nil

	default:
		return file.Content, nil
	}
}

// processHybridFile handles files with hybrid generation strategy
func (g *Generator) processHybridFile(file TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (string, error) {
	// Special handling for YAML files and Makefile
	if isConfigYaml(file.Name) || isMakefile(file.Name) {
		// First render the template
		content, err := g.renderer.RenderTemplateWithData(file.Name, file.Content, variables)
		if err != nil {
			return "", fmt.Errorf("failed to render template %s: %w", file.Name, err)
		}

		// Then apply YAML processing
		if isConfigYaml(file.Name) {
			return g.processConfigYml(content, data)
		} else if isMakefile(file.Name) {
			return g.processMakefile(content, data)
		}
	}

	// Standard hybrid approach for other files
	content, err := g.renderer.RenderTemplateWithData(file.Name, file.Content, variables)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", file.Name, err)
	}

	// Apply AST transformations
	return g.applyASTTransformations(content, file.Name, data)
}

// isConfigYaml checks if a file is config.yml
func isConfigYaml(fileName string) bool {
	// Check for config.yml file regardless of path
	if strings.HasSuffix(fileName, "config.yml"+tmpSuffix) {
		return true
	}

	// Traditional check for backward compatibility
	return strings.Contains(fileName, "config") && strings.HasSuffix(fileName, ".yml"+tmpSuffix)
}

// isMakefile checks if a file is a Makefile
func isMakefile(fileName string) bool {
	return strings.Contains(fileName, "Makefile")
}

// generateWithAST generates code using AST manipulation
func (g *Generator) generateWithAST(file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	// Parse file to get AST
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file.Name, file.Content, parser.ParseComments)
	if err != nil {
		// If parsing fails, fall back to template rendering
		fmt.Printf("Warning: Failed to parse %s as Go code: %v\n", file.Name, err)
		fmt.Printf("Falling back to template rendering for %s\n", file.Name)

		return g.renderFallback(file, data)
	}

	// Apply appropriate generator based on file type
	if err := g.applyAstGenerator(f, file.Name, data); err != nil {
		return "", fmt.Errorf("failed to generate %s: %w", file.Name, err)
	}

	// Format generated AST back to source code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		// If formatting fails, fall back to template rendering
		fmt.Printf("Warning: Failed to format generated code for %s: %v\n", file.Name, err)
		fmt.Printf("Falling back to template rendering for %s\n", file.Name)

		return g.renderFallback(file, data)
	}

	return buf.String(), nil
}

// applyAstGenerator applies the appropriate AST generator based on file type
func (g *Generator) applyAstGenerator(f *ast.File, fileName string, data *eventdata.TemplateEventData) error {
	switch {
	case strings.Contains(fileName, "repository"):
		return repository.NewGenerator().Generate(f, data)
	case strings.Contains(fileName, "model"):
		return model.NewGenerator().Generate(f, data)
	case strings.Contains(fileName, "entity"):
		return entity.NewGenerator().Generate(f, data)
	case strings.Contains(fileName, "service"):
		return service.NewGenerator().Generate(f, data)
	}
	return nil
}

// renderFallback renders a template as fallback when AST processing fails
func (g *Generator) renderFallback(file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	return g.renderer.RenderTemplateWithData(file.Name, file.Content, map[string]interface{}{
		"Name":      data.Name,
		"Database":  data.Database,
		"Endpoints": data.Endpoints,
		"Docker":    data.Docker,
		"Advanced":  data.Advanced,
	})
}

// applyASTTransformations applies AST transformations to the rendered template
func (g *Generator) applyASTTransformations(content, fileName string, data *eventdata.TemplateEventData) (string, error) {
	// Skip empty files
	if len(content) == 0 {
		return content, nil
	}

	// For YAML files and other non-Go files, just return the content
	if !strings.HasSuffix(fileName, ".go"+tmpSuffix) {
		// Skip config.yml and Makefile, as they're already processed in generateFiles
		if isConfigYaml(fileName) || isMakefile(fileName) {
			return content, nil
		}
		return content, nil
	}

	// Parse the content
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fileName, content, parser.ParseComments)
	if err != nil {
		// If parsing fails, log and return original content
		fmt.Printf("Warning: Failed to parse content for %s: %v\n", fileName, err)
		return content, nil
	}

	// Use centralized feature detector
	fs := features.DetectFeatures(data)

	fmt.Printf("AST transformations for %s: hasGRPC=%v, hasGraphQL=%v, hasDatabase=%v\n",
		fileName, fs.HasGRPC, fs.HasGraphQL, fs.HasDatabase)

	// Apply transformations based on file type
	appliedTransformations := g.applySpecificTransformations(f, fileName, data, fs)

	if err != nil {
		fmt.Printf("Warning: Failed to apply AST transformations to %s: %v\n", fileName, err)
		return content, nil
	}

	// If no transformations were applied, return original content
	if !appliedTransformations {
		return content, nil
	}

	// Format the modified AST
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		fmt.Printf("Warning: Failed to format AST for %s: %v\n", fileName, err)
		return content, nil
	}

	return buf.String(), nil
}

// applySpecificTransformations applies specific transformations based on file type
func (g *Generator) applySpecificTransformations(f *ast.File, fileName string, data *eventdata.TemplateEventData, fs *features.FeatureSet) bool {
	var appliedTransformations bool

	// Apply transformations based on file type
	switch {
	case strings.Contains(fileName, "app.go"+tmpSuffix):
		if err := app.NewGenerator().Generate(f, data); err != nil {
			fmt.Printf("Warning: Failed to apply app transformation: %v\n", err)
		}
		appliedTransformations = true

	case strings.Contains(fileName, "config.go"+tmpSuffix):
		if err := config.NewGenerator().Generate(f, data); err != nil {
			fmt.Printf("Warning: Failed to apply config transformation: %v\n", err)
		}
		appliedTransformations = true

	case strings.Contains(fileName, "service.go"+tmpSuffix):
		// Skip service generation for specific service files
		if !strings.Contains(fileName, "graphql/service.go"+tmpSuffix) &&
			!strings.Contains(fileName, "grpc/service.go"+tmpSuffix) &&
			!strings.Contains(fileName, "templates/") {
			if err := service.NewGenerator().Generate(f, data); err != nil {
				fmt.Printf("Warning: Failed to apply service transformation: %v\n", err)
			}
			appliedTransformations = true
		}
	}

	// Protocol-specific transformations
	if fs.HasGRPC && strings.Contains(fileName, "grpc") {
		appliedTransformations = true
	}

	if fs.HasGraphQL && strings.Contains(fileName, "graphql") {
		appliedTransformations = true
	}

	if fs.HasDatabase && isDatabaseFile(fileName) {
		appliedTransformations = true
	}

	return appliedTransformations
}

// processConfigYml processes config.yml based on input parameters
func (g *Generator) processConfigYml(content string, data *eventdata.TemplateEventData) (string, error) {
	yamlGen := yaml.NewGenerator()
	return yamlGen.ProcessConfigYAML(content, data)
}

// processMakefile processes Makefile based on input parameters
func (g *Generator) processMakefile(content string, data *eventdata.TemplateEventData) (string, error) {
	yamlGen := yaml.NewGenerator()
	return yamlGen.ProcessMakefile(content, data)
}

// createArchive creates a ZIP archive with the generated files
func (g *Generator) createArchive(files map[string][]byte, id string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add files to the archive
	for path, content := range files {
		file, err := zipWriter.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in archive: %w", err)
		}

		if _, err := file.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write content to archive: %w", err)
		}
	}

	// Close the writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close archive writer: %w", err)
	}

	// Save debug archive if enabled
	g.saveDebugArchive(buf.Bytes(), id)

	return buf.Bytes(), nil
}

// saveDebugArchive saves the archive locally if debug mode is enabled
func (g *Generator) saveDebugArchive(archive []byte, id string) {
	if !g.debugArchives {
		return
	}

	// Use the specified debug directory or fallback to default
	debugDir := g.debugDir
	if debugDir == "" {
		debugDir = "debug_archives"
	}

	// Ensure directory exists
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		fmt.Printf("Warning: Failed to create debug directory: %v\n", err)
		return
	}

	// Write archive to file
	archivePath := filepath.Join(debugDir, fmt.Sprintf("template_%s.zip", id))
	if err := os.WriteFile(archivePath, archive, 0o644); err != nil {
		fmt.Printf("Warning: Failed to save debug archive: %v\n", err)
		return
	}

	fmt.Printf("Debug archive saved to: %s\n", archivePath)
}
