package engine

import (
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"

	"go-init-gen/internal/eventdata"
	"go-init-gen/internal/generator/engine/generators/features"
)

var strategyLogger = log.New(log.Writer(), "[FileStrategy] ", log.LstdFlags)

// CodeGenStrategy represents a strategy for generating code
type CodeGenStrategy string

const (
	// StrategyTextTemplate uses text/template for code generation
	StrategyTextTemplate CodeGenStrategy = "text_template"

	// StrategyASTGeneration uses Go AST for code generation
	StrategyASTGeneration CodeGenStrategy = "ast_generation"

	// StrategyHybrid combines text templates and AST manipulation
	StrategyHybrid CodeGenStrategy = "hybrid"

	// StrategyRaw uses raw file content without processing
	StrategyRaw CodeGenStrategy = "raw"
)

// FileStrategy defines how files should be handled during generation
type FileStrategy struct {
	featureSet *features.FeatureSet
}

// NewFileStrategy creates a new file strategy based on the template data
func NewFileStrategy(data *eventdata.TemplateEventData) *FileStrategy {
	return &FileStrategy{
		featureSet: features.DetectFeatures(data),
	}
}

// ShouldIncludeFile determines if a file should be included in the generated output
func (s *FileStrategy) ShouldIncludeFile(file fs.DirEntry) bool {
	filePath := file.Name()

	// Skip directories in this phase
	if file.IsDir() {
		return false
	}

	// Always include base files
	if s.isBaseFile(filePath) {
		return true
	}

	// Skip files that are not relevant to our features
	if !s.featureSet.HasGRPC && s.isGRPCFile(filePath) {
		return false
	}

	if !s.featureSet.HasGraphQL && s.isGraphQLFile(filePath) {
		return false
	}

	// Skip REST files if REST is not enabled
	if !s.featureSet.HasREST && isRESTFile(filePath) {
		return false
	}

	// Skip Kafka files if Kafka is not enabled
	if !s.featureSet.HasKafka && isKafkaFile(filePath) {
		return false
	}

	// Skip database files if database is not enabled
	if !s.featureSet.HasDatabase && s.isDatabaseFile(filePath) {
		return false
	}

	// Skip database-specific files for database types not in use
	if s.featureSet.HasDatabase {
		// Skip PostgreSQL files if not using PostgreSQL
		if !s.featureSet.HasPostgres() && isPostgresFile(filePath) {
			return false
		}

		// Skip MySQL files if not using MySQL
		if !s.featureSet.HasMySQL() && isMySQLFile(filePath) {
			return false
		}

		// Skip MongoDB files if not using MongoDB
		if !s.featureSet.HasMongoDB() && isMongoDBFile(filePath) {
			return false
		}

		// Skip Redis files if not using Redis
		if !s.featureSet.HasRedis() && isRedisFile(filePath) {
			return false
		}
	}

	// Skip files for unsupported technologies
	if isUnsupportedFile(filePath) {
		return false
	}

	// Include the file by default
	return true
}

// RequiresTransformation determines if a file requires AST transformation
func (s *FileStrategy) RequiresTransformation(filePath string) bool {
	baseName := filepath.Base(filePath)

	// Identify files that need AST transformation
	switch baseName {
	case "app.go", "config.go":
		return true
	default:
		return false
	}
}

// GetTransformationType returns the type of transformation needed
func (s *FileStrategy) GetTransformationType(filePath string) string {
	baseName := filepath.Base(filePath)
	return baseName
}

// GetFeatureFlags returns a map of enabled features
func (s *FileStrategy) GetFeatureFlags() map[string]bool {
	return map[string]bool{
		"hasGRPC":       s.featureSet.HasGRPC,
		"hasGraphQL":    s.featureSet.HasGraphQL,
		"hasREST":       s.featureSet.HasREST,
		"hasHTTP":       s.featureSet.HasHTTP,
		"hasKafka":      s.featureSet.HasKafka,
		"hasDatabase":   s.featureSet.HasDatabase,
		"hasPostgreSQL": s.featureSet.HasPostgres(),
		"hasMySQL":      s.featureSet.HasMySQL(),
		"hasMongoDB":    s.featureSet.HasMongoDB(),
		"hasRedis":      s.featureSet.HasRedis(),
	}
}

// GetDatabaseType returns the database type
func (s *FileStrategy) GetDatabaseType() string {
	return s.featureSet.DatabaseType
}

// Debug returns a string representation of the strategy for debugging
func (s *FileStrategy) Debug() string {
	return fmt.Sprintf("FileStrategy{features: %+v}", s.featureSet)
}

// isBaseFile checks if a file is a base file that should always be included
func (s *FileStrategy) isBaseFile(filePath string) bool {
	baseFiles := []string{
		"main.go", "app.go", "config.go", "Makefile", "Dockerfile",
		".gitignore", "go.mod", "go.sum", "README.md", "LICENSE",
	}

	for _, baseFile := range baseFiles {
		if filepath.Base(filePath) == baseFile {
			return true
		}
	}

	return false
}

// isGRPCFile checks if a file is related to gRPC
func (s *FileStrategy) isGRPCFile(filePath string) bool {
	if strings.Contains(filePath, "grpc") ||
		strings.Contains(filePath, "proto") ||
		strings.HasSuffix(filePath, ".proto") ||
		strings.Contains(filePath, "server/grpc") {
		return true
	}
	return false
}

// isGraphQLFile checks if a file is related to GraphQL
func (s *FileStrategy) isGraphQLFile(filePath string) bool {
	if strings.Contains(filePath, "graphql") ||
		strings.Contains(filePath, "gql") ||
		strings.HasSuffix(filePath, ".graphqls") ||
		strings.Contains(filePath, "server/graphql") {
		return true
	}
	return false
}

// isDatabaseFile checks if a file is related to database operations
func (s *FileStrategy) isDatabaseFile(filePath string) bool {
	// Common database files
	if strings.Contains(filePath, "repository") ||
		strings.Contains(filePath, "storage") ||
		strings.Contains(filePath, "database") ||
		strings.Contains(filePath, "model") ||
		strings.Contains(filePath, "migration") ||
		strings.Contains(filePath, "sql") {
		return true
	}

	// Database-specific files
	if s.featureSet.HasPostgres() && (strings.Contains(filePath, "postgres") ||
		strings.Contains(filePath, "postgresql") ||
		strings.HasSuffix(filePath, ".sql")) {
		return true
	}

	if s.featureSet.HasMySQL() && (strings.Contains(filePath, "mysql") ||
		strings.HasSuffix(filePath, ".sql")) {
		return true
	}

	if s.featureSet.HasMongoDB() && (strings.Contains(filePath, "mongo") ||
		strings.Contains(filePath, "mongodb")) {
		return true
	}

	if s.featureSet.HasRedis() && strings.Contains(filePath, "redis") {
		return true
	}

	return false
}

// UnsupportedPatterns is now deprecated, kept for backward compatibility
var UnsupportedPatterns = []string{}

// Helper function to check if a file is related to unsupported technologies
func isUnsupportedFile(filePath string) bool {
	// We don't consider any file pattern as completely unsupported anymore
	// Instead, we use feature detection to determine if a file should be included

	// But we might want to keep certain patterns that should never be included
	// regardless of the enabled features
	excludedPatterns := []string{
		".git/",
		".github/",
		".idea/",
		".vscode/",
		"node_modules/",
		"vendor/",
		"dist/",
		// Excluding build/ directory but preserving build/config/ directory
		"build/docker/",
		"build/migrations/",
		"build/tmp/",
		"build/logs/",
		"build/scripts/",
		// Preserve build/config directory, which contains config.yml.tmpl
		".DS_Store",
		"__pycache__/",
	}

	for _, pattern := range excludedPatterns {
		if strings.Contains(filePath, pattern) {
			return true
		}
	}

	// Special check to avoid excluding build/config directory
	if strings.Contains(filePath, "build/") && !strings.Contains(filePath, "build/config/") {
		// Check if this is some other build subdirectory that should be excluded
		parts := strings.Split(filePath, "build/")
		if len(parts) > 1 {
			secondPart := parts[1]
			if !strings.HasPrefix(secondPart, "config/") && strings.Contains(secondPart, "/") {
				return true
			}
		}
	}

	return false
}

// StrategyMap defines mappings between file patterns and their generation strategies
var StrategyMap = map[string]CodeGenStrategy{
	// Base Go files
	"app.go":    StrategyHybrid,
	"config.go": StrategyHybrid,
	"main.go":   StrategyTextTemplate,

	// File types
	"repository": StrategyTextTemplate,
	"entity":     StrategyTextTemplate,
	"model":      StrategyTextTemplate,
	"service":    StrategyHybrid,
	// "handler":    StrategyHybrid,
	// "middleware": StrategyHybrid,
	// "controller": StrategyHybrid,
	"resolver": StrategyTextTemplate,

	// Database related
	"database/models": StrategyTextTemplate,
	"models.go":       StrategyTextTemplate,
	"database":        StrategyTextTemplate,
	"migrations":      StrategyTextTemplate,
	".sql":            StrategyTextTemplate,

	// API related
	".proto":   StrategyTextTemplate,
	".graphql": StrategyTextTemplate,
	"schema":   StrategyTextTemplate,
	"gql":      StrategyTextTemplate,

	// Configuration
	".yml":        StrategyHybrid,
	".yaml":       StrategyTextTemplate,
	"Makefile":    StrategyTextTemplate,
	".gitignore":  StrategyTextTemplate,
	"Dockerfile":  StrategyTextTemplate,
	".Dockerfile": StrategyTextTemplate,
	"VERSION":     StrategyRaw,
}

// DirectoryStrategyMap defines mappings between directory patterns and their generation strategies
var DirectoryStrategyMap = map[string]CodeGenStrategy{
	"api":     StrategyHybrid,
	"grpc":    StrategyHybrid,
	"graphql": StrategyHybrid,
	"tools":   StrategyTextTemplate,
}

// GetFileStrategy determines the appropriate generation strategy for a file
func GetFileStrategy(filePath string) CodeGenStrategy {
	ext := filepath.Ext(filePath)
	fileName := filepath.Base(filePath)
	dirName := filepath.Dir(filePath)

	// Log the file being processed
	strategyLogger.Printf("Determining strategy for file: %s", filePath)

	// Check if file is unsupported
	if isUnsupportedFile(filePath) {
		strategyLogger.Printf("'%s' relates to unsupported technology, using raw approach", filePath)
		return StrategyRaw
	}

	// Special debug for config.yml files to help troubleshoot filtering issues
	if strings.Contains(fileName, "config.yml") {
		strategyLogger.Printf("Processing config.yml file: '%s', in directory: '%s'", fileName, dirName)
	}

	// First check exact filename matches
	if strategy, exists := StrategyMap[fileName]; exists {
		strategyLogger.Printf("'%s' matched exact filename, using %s", fileName, strategy)
		return strategy
	}

	// Check file extension matches
	if strategy, exists := StrategyMap[ext]; exists {
		strategyLogger.Printf("'%s' matched extension %s, using %s", fileName, ext, strategy)
		return strategy
	}

	// Check for substring matches in filename
	for pattern, strategy := range StrategyMap {
		if strings.Contains(fileName, pattern) {
			strategyLogger.Printf("'%s' contains pattern '%s', using %s", fileName, pattern, strategy)
			return strategy
		}
	}

	// Check for substring matches in filepath
	for pattern, strategy := range StrategyMap {
		if strings.Contains(filePath, pattern) {
			strategyLogger.Printf("'%s' contains pattern '%s', using %s", filePath, pattern, strategy)
			return strategy
		}
	}

	// Check directory patterns
	for pattern, strategy := range DirectoryStrategyMap {
		if strings.Contains(dirName, pattern) {
			strategyLogger.Printf("'%s' is in directory matching '%s', using %s", filePath, pattern, strategy)
			return strategy
		}
	}

	// Default to text template for all other files
	strategyLogger.Printf("'%s' uses default template approach", filePath)
	return StrategyTextTemplate
}

// Helper functions for file type detection

// isRESTFile checks if a file is related to REST API
func isRESTFile(filePath string) bool {
	return strings.Contains(filePath, "rest") ||
		strings.Contains(filePath, "http") ||
		strings.Contains(filePath, "handler") ||
		strings.Contains(filePath, "controller") ||
		strings.Contains(filePath, "router") ||
		strings.Contains(filePath, "swagger") ||
		strings.Contains(filePath, "openapi")
}

// isKafkaFile checks if a file is related to Kafka
func isKafkaFile(filePath string) bool {
	return strings.Contains(filePath, "kafka") ||
		strings.Contains(filePath, "event") ||
		strings.Contains(filePath, "consumer") ||
		strings.Contains(filePath, "producer") ||
		strings.Contains(filePath, "subscriber") ||
		strings.Contains(filePath, "publisher")
}

// isPostgresFile checks if a file is specifically for PostgreSQL
func isPostgresFile(filePath string) bool {
	return strings.Contains(strings.ToLower(filePath), "postgres") ||
		strings.Contains(strings.ToLower(filePath), "postgresql") ||
		strings.Contains(filePath, "pg/") ||
		strings.Contains(filePath, "/pg/")
}

// isMySQLFile checks if a file is specifically for MySQL
func isMySQLFile(filePath string) bool {
	return strings.Contains(strings.ToLower(filePath), "mysql")
}

// isMongoDBFile checks if a file is specifically for MongoDB
func isMongoDBFile(filePath string) bool {
	return strings.Contains(strings.ToLower(filePath), "mongo") ||
		strings.Contains(strings.ToLower(filePath), "mongodb") ||
		strings.Contains(filePath, "bson")
}

// isRedisFile checks if a file is specifically for Redis
func isRedisFile(filePath string) bool {
	return strings.Contains(strings.ToLower(filePath), "redis")
}

// GetFeatureSet returns the detected feature set
func (s *FileStrategy) GetFeatureSet() *features.FeatureSet {
	return s.featureSet
}
