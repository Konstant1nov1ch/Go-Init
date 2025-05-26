package engine

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/app"
	"go-init-gen/internal/generator/engine/generators/config"
	"go-init-gen/internal/generator/engine/generators/entity"
	"go-init-gen/internal/generator/engine/generators/features"
	"go-init-gen/internal/generator/engine/generators/model"
	"go-init-gen/internal/generator/engine/generators/repository"
	"go-init-gen/internal/generator/engine/generators/service"
	"go-init-gen/internal/generator/engine/generators/yaml"
)

// ContentGenerator handles generation of file content based on different strategies
type ContentGenerator struct {
	renderer    Renderer
	templateDir string
}

// NewContentGenerator creates a new content generator
func NewContentGenerator(templateDir string) *ContentGenerator {
	return &ContentGenerator{
		renderer:    NewRenderer(templateDir),
		templateDir: templateDir,
	}
}

// GenerateFiles generates the content for each file
func (cg *ContentGenerator) GenerateFiles(files []TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (map[string][]byte, error) {
	generatedFiles := make(map[string][]byte)

	for _, file := range files {
		content, err := cg.generateFileContent(file, data, variables)
		if err != nil {
			return nil, fmt.Errorf("failed to generate content for %s: %w", file.Name, err)
		}

		generatedFiles[file.TargetPath] = []byte(content)
	}

	return generatedFiles, nil
}

// generateFileContent generates content for a single file based on its strategy
func (cg *ContentGenerator) generateFileContent(file TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (string, error) {
	switch file.CodeGeneration {
	case StrategyTextTemplate:
		return cg.renderer.RenderTemplateWithData(file.Name, file.Content, variables)

	case StrategyASTGeneration:
		return cg.generateWithAST(file, data)

	case StrategyHybrid:
		return cg.processHybridFile(file, data, variables)

	case StrategyRaw:
		return file.Content, nil

	default:
		return file.Content, nil
	}
}

// processHybridFile handles files with hybrid generation strategy
func (cg *ContentGenerator) processHybridFile(file TemplateFile, data *eventdata.TemplateEventData, variables map[string]interface{}) (string, error) {
	// Special handling for YAML files and Makefile
	if cg.isConfigYaml(file.Name) || cg.isMakefile(file.Name) {
		// First render the template
		content, err := cg.renderer.RenderTemplateWithData(file.Name, file.Content, variables)
		if err != nil {
			return "", fmt.Errorf("failed to render template %s: %w", file.Name, err)
		}

		// Then apply YAML processing
		if cg.isConfigYaml(file.Name) {
			return cg.processConfigYml(content, data)
		} else if cg.isMakefile(file.Name) {
			return cg.processMakefile(content, data)
		}
	}

	// Standard hybrid approach for other files
	content, err := cg.renderer.RenderTemplateWithData(file.Name, file.Content, variables)
	if err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", file.Name, err)
	}

	// Apply AST transformations
	return cg.applyASTTransformations(content, file.Name, data)
}

// generateWithAST generates code using AST manipulation
func (cg *ContentGenerator) generateWithAST(file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	// Parse file to get AST
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file.Name, file.Content, parser.ParseComments)
	if err != nil {
		// If parsing fails, fall back to template rendering
		fmt.Printf("Warning: Failed to parse %s as Go code: %v\n", file.Name, err)
		fmt.Printf("Falling back to template rendering for %s\n", file.Name)

		return cg.renderFallback(file, data)
	}

	// Apply appropriate generator based on file type
	if err := cg.applyAstGenerator(f, file.Name, data); err != nil {
		return "", fmt.Errorf("failed to generate %s: %w", file.Name, err)
	}

	// Format generated AST back to source code
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		// If formatting fails, fall back to template rendering
		fmt.Printf("Warning: Failed to format generated code for %s: %v\n", file.Name, err)
		fmt.Printf("Falling back to template rendering for %s\n", file.Name)

		return cg.renderFallback(file, data)
	}

	return buf.String(), nil
}

// applyAstGenerator applies the appropriate AST generator based on file type
func (cg *ContentGenerator) applyAstGenerator(f *ast.File, fileName string, data *eventdata.TemplateEventData) error {
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
func (cg *ContentGenerator) renderFallback(file TemplateFile, data *eventdata.TemplateEventData) (string, error) {
	return cg.renderer.RenderTemplateWithData(file.Name, file.Content, map[string]interface{}{
		"Name":      data.Name,
		"Database":  data.Database,
		"Endpoints": data.Endpoints,
		"Docker":    data.Docker,
		"Advanced":  data.Advanced,
	})
}

// applyASTTransformations applies AST transformations to the rendered template
func (cg *ContentGenerator) applyASTTransformations(content, fileName string, data *eventdata.TemplateEventData) (string, error) {
	// Skip empty files
	if len(content) == 0 {
		return content, nil
	}

	// For YAML files and other non-Go files, just return the content
	if !strings.HasSuffix(fileName, ".go"+tmpSuffix) {
		// Skip config.yml and Makefile, as they're already processed in generateFiles
		if cg.isConfigYaml(fileName) || cg.isMakefile(fileName) {
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
	appliedTransformations := cg.applySpecificTransformations(f, fileName, data, fs)

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
func (cg *ContentGenerator) applySpecificTransformations(f *ast.File, fileName string, data *eventdata.TemplateEventData, fs *features.FeatureSet) bool {
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

	if fs.HasDatabase && cg.isDatabaseFile(fileName) {
		appliedTransformations = true
	}

	return appliedTransformations
}

// isConfigYaml checks if a file is config.yml
func (cg *ContentGenerator) isConfigYaml(fileName string) bool {
	// Check for config.yml file regardless of path
	if strings.HasSuffix(fileName, "config.yml"+tmpSuffix) {
		return true
	}

	// Traditional check for backward compatibility
	return strings.Contains(fileName, "config") && strings.HasSuffix(fileName, ".yml"+tmpSuffix)
}

// isMakefile checks if a file is a Makefile
func (cg *ContentGenerator) isMakefile(fileName string) bool {
	return strings.Contains(fileName, "Makefile")
}

// isDatabaseFile checks if a file is related to a database
func (cg *ContentGenerator) isDatabaseFile(fileName string) bool {
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

// processConfigYml processes config.yml based on input parameters
func (cg *ContentGenerator) processConfigYml(content string, data *eventdata.TemplateEventData) (string, error) {
	yamlGen := yaml.NewGenerator()
	return yamlGen.ProcessConfigYAML(content, data)
}

// processMakefile processes Makefile based on input parameters
func (cg *ContentGenerator) processMakefile(content string, data *eventdata.TemplateEventData) (string, error) {
	yamlGen := yaml.NewGenerator()
	return yamlGen.ProcessMakefile(content, data)
}
