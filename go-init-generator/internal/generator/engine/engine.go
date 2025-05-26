package engine

import (
	"context"
	"os"
	"path/filepath"

	"go-init-gen/internal/eventdata"
)

const (
	tmpSuffix = ".tmpl"
)

// Generator handles the code generation process using a pipeline architecture
type Generator struct {
	pipeline *GenerationPipeline
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
		pipeline: NewGenerationPipeline(templateDir, debugArchives, debugDir),
	}
}

// Generate creates a template based on input data
func (g *Generator) Generate(ctx context.Context, template *eventdata.ProcessTemplate) ([]byte, error) {
	return g.pipeline.Execute(ctx, template)
}

// SetDebugArchives enables or disables debug archive saving (for testing)
func (g *Generator) SetDebugArchives(enabled bool) {
	g.pipeline.archiver.debugArchives = enabled
}

// SetDebugDir sets the debug directory path (for testing)
func (g *Generator) SetDebugDir(dir string) {
	g.pipeline.archiver.debugDir = dir
}

// GetDebugArchives returns whether debug archives are enabled (for testing)
func (g *Generator) GetDebugArchives() bool {
	return g.pipeline.archiver.debugArchives
}

// GetDebugDir returns the debug directory path (for testing)
func (g *Generator) GetDebugDir() string {
	return g.pipeline.archiver.debugDir
}
