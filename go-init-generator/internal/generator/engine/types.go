package engine

import (
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/features"
)

// NewTemplateFile represents a template file with metadata
type NewTemplateFile struct {
	// Name is the file name (with path)
	Name string
	// Content is the file content (template or source)
	Content string
	// CodeGeneration defines the strategy to use for this file
	CodeGeneration CodeGenStrategy
	// UseAST indicates whether to use AST-based code generation (deprecated, use CodeGeneration)
	UseAST bool
	// TargetPath is the path where the file should be written
	TargetPath string
	// Metadata holds additional file-specific data
	Metadata map[string]interface{}
}

// TemplateFile represents a template file to be generated
type TemplateFile struct {
	Name           string                 // Имя файла (с путем)
	Content        string                 // Содержимое файла (шаблон или исходный код)
	UseAST         bool                   // Указывает, использовать ли AST-генерацию кода
	CodeGeneration CodeGenStrategy        // Стратегия генерации кода
	TargetPath     string                 // Путь, куда должен быть записан файл
	Metadata       map[string]interface{} // Дополнительные метаданные файла
}

// FileClassification represents a file classification for determining the generator
type FileClassification struct {
	Pattern     string          // Шаблон файла (например, "*.go")
	Strategy    CodeGenStrategy // Стратегия генерации кода
	Description string          // Описание классификации
}

// Standard file classifications
var StandardFileClassifications = []FileClassification{
	// Simple templates
	{Pattern: "main.go", Strategy: StrategyTextTemplate, Description: "Main application entry point"},
	{Pattern: "app.go", Strategy: StrategyTextTemplate, Description: "Application setup"},
	{Pattern: "config.go", Strategy: StrategyTextTemplate, Description: "Configuration management"},
	{Pattern: "README.md", Strategy: StrategyTextTemplate, Description: "Project documentation"},
	{Pattern: "Makefile", Strategy: StrategyTextTemplate, Description: "Build automation"},
	{Pattern: "Dockerfile", Strategy: StrategyTextTemplate, Description: "Docker configuration"},
	{Pattern: ".gitignore", Strategy: StrategyTextTemplate, Description: "Git ignore rules"},

	// AST-based generation
	{Pattern: "repository.go", Strategy: StrategyASTGeneration, Description: "Data access layer"},
	{Pattern: "model.go", Strategy: StrategyASTGeneration, Description: "Data models"},
	{Pattern: "handler.go", Strategy: StrategyASTGeneration, Description: "HTTP handlers"},
	{Pattern: "service.go", Strategy: StrategyASTGeneration, Description: "Business logic"},
	{Pattern: "entity.go", Strategy: StrategyASTGeneration, Description: "Domain entities"},

	// Hybrid approach
	{Pattern: "middleware.go", Strategy: StrategyHybrid, Description: "HTTP middleware"},
}

// DetermineStrategy determines the appropriate generation strategy based on file type
func DetermineStrategy(fileName string) CodeGenStrategy {
	// Delegate to the GetFileStrategy function
	return GetFileStrategy(fileName)
}

// Customizer processes input data and customizes variables for templates
type Customizer struct {
	variables    map[string]interface{}
	featureFlags map[string]bool
}

// NewCustomizer creates a new customizer instance
func NewCustomizer() *Customizer {
	return &Customizer{
		variables:    make(map[string]interface{}),
		featureFlags: make(map[string]bool),
	}
}

// ProcessInput processes the input data and prepares variables
func (c *Customizer) ProcessInput(input *eventdata.TemplateEventData) {
	// Extract service name
	name := input.Name
	c.variables["serviceName"] = name
	c.variables["serviceNameCamel"] = ToCamelCase(name)
	c.variables["serviceNameSnake"] = ToSnakeCase(name)

	// Add the Name variable directly which is used by templates
	c.variables["Name"] = name

	// Use the features package to detect features
	fs := features.DetectFeatures(input)

	// Set the feature flags based on the feature set
	c.featureFlags["hasGRPC"] = fs.HasGRPC
	c.featureFlags["hasGraphQL"] = fs.HasGraphQL
	c.featureFlags["hasREST"] = fs.HasREST
	c.featureFlags["hasHTTP"] = fs.HasHTTP
	c.featureFlags["hasDatabase"] = fs.HasDatabase

	// Handle database-specific flags
	if fs.HasDatabase {
		c.featureFlags["hasDatabase"] = true

		if fs.HasPostgres() {
			c.featureFlags["hasPostgres"] = true
		} else if fs.HasMySQL() {
			c.featureFlags["hasMySQL"] = true
		}

		// Add database type to variables for templates
		c.variables["databaseType"] = fs.DatabaseType
	}

	// Handle additional features (Kafka is not in the features package yet)
	for _, endpoint := range input.Endpoints {
		protocol := strings.ToUpper(endpoint.Protocol)
		if protocol == "KAFKA" {
			c.featureFlags["hasKafka"] = true
		}
	}

	// Docker settings
	if input.Docker.ImageName != "" {
		c.featureFlags["hasDocker"] = true
		c.variables["dockerImage"] = input.Docker.ImageName
		if input.Docker.Registry != "" {
			c.variables["dockerRegistry"] = input.Docker.Registry
		}
	}

	// Advanced features
	if input.Advanced != nil {
		if input.Advanced.EnableAuthentication {
			c.featureFlags["hasAuth"] = true
		}
		if input.Advanced.GenerateSwaggerDocs {
			c.featureFlags["hasSwagger"] = true
		}
	}
}

// GetVariables returns the customized variables
func (c *Customizer) GetVariables() map[string]interface{} {
	return c.variables
}

// GetFeatureFlags returns feature flags
func (c *Customizer) GetFeatureFlags() map[string]bool {
	return c.featureFlags
}

// IsFeatureEnabled checks if a specific feature is enabled
func (c *Customizer) IsFeatureEnabled(feature string) bool {
	enabled, exists := c.featureFlags[feature]
	return exists && enabled
}

// Helper functions for string manipulation

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	// Replace non-alphanumeric with space
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return ' '
	}, s)

	// Split into words
	words := strings.Fields(s)

	// Convert to camel case
	for i, word := range words {
		if i == 0 {
			words[i] = strings.ToLower(word)
		} else {
			words[i] = strings.Title(strings.ToLower(word))
		}
	}

	return strings.Join(words, "")
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	// Replace non-alphanumeric with underscore
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '_'
	}, s)

	return strings.ToLower(s)
}

// ToKebabCase converts a string to kebab-case
func ToKebabCase(s string) string {
	// Replace non-alphanumeric with hyphen
	s = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			return r
		}
		return '-'
	}, s)

	return strings.ToLower(s)
}
