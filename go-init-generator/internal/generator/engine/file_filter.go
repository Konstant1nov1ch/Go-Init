package engine

import (
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/features"
)

// FeatureBasedFileFilter handles filtering of template files based on features
type FeatureBasedFileFilter struct{}

// NewFileFilter creates a new feature-based file filter
func NewFileFilter() *FeatureBasedFileFilter {
	return &FeatureBasedFileFilter{}
}

// FilterFiles filters template files based on input features
func (f *FeatureBasedFileFilter) FilterFiles(files []TemplateFile, data *eventdata.TemplateEventData) []TemplateFile {
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
	serviceNoDBFile := f.findNoDatabaseServiceFile(files, fs.HasDatabase)

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
		if f.isBaseFile(file.Name) {
			filteredFiles = append(filteredFiles, file)
			continue
		}

		// Apply filter rules
		shouldInclude := f.shouldIncludeFile(file.Name, featuresMap, fs, data)

		// If the file passed all checks, include it
		if shouldInclude {
			filteredFiles = append(filteredFiles, file)
		}
	}

	return filteredFiles
}

// findNoDatabaseServiceFile finds the no-database service file for special handling
func (f *FeatureBasedFileFilter) findNoDatabaseServiceFile(files []TemplateFile, hasDatabase bool) *TemplateFile {
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
func (f *FeatureBasedFileFilter) shouldIncludeFile(fileName string, featuresMap map[string]bool, fs *features.FeatureSet, data *eventdata.TemplateEventData) bool {
	// Check against each pattern in our filter rules
	for pattern, rule := range featureFilterRules {
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
	if fs.HasDatabase && f.isDatabaseFile(fileName) && !fs.HasPostgres() && strings.Contains(fileName, strings.ToLower(data.Database.Type)) {
		return false
	}

	return true
}

// isBaseFile checks if a file is a base file for any project
func (f *FeatureBasedFileFilter) isBaseFile(fileName string) bool {
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
func (f *FeatureBasedFileFilter) isDatabaseFile(fileName string) bool {
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

// featureFilterRules defines rules for including/excluding files based on features
var featureFilterRules = map[string]struct {
	// RequiredFeatures lists features that must be present for the file to be included
	RequiredFeatures []string
	// ExcludedFeatures lists features that, if present, cause the file to be excluded
	ExcludedFeatures []string
}{
	// gRPC-specific files
	"grpc":                   {RequiredFeatures: []string{"hasGRPC"}, ExcludedFeatures: nil},
	"proto":                  {RequiredFeatures: []string{"hasGRPC"}, ExcludedFeatures: nil},
	"pkg/api/grpc/README.md": {RequiredFeatures: []string{"hasGRPC"}, ExcludedFeatures: nil},

	// GraphQL-specific files
	"graphql":                   {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil},
	"gql":                       {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil},
	"tools/":                    {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil}, // tools directory is GraphQL specific (for gqlgen)
	"pkg/api/graphql/README.md": {RequiredFeatures: []string{"hasGraphQL"}, ExcludedFeatures: nil},

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
