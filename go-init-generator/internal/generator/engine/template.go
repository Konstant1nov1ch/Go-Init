package engine

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-init-gen/internal/eventdata"
)

// Template defines the interface for all code templates
type Template interface {
	// Name returns the unique identifier for this template
	Name() string

	// Description returns a human-readable description
	Description() string

	// Generate creates a template based on input data
	Generate(ctx context.Context, input *eventdata.TemplateEventData) ([]byte, error)

	// SupportedFeatures returns a list of features this template supports
	SupportedFeatures() []string

	// Files returns the list of files in this template
	Files() []TemplateFile
}

// TemplateRegistry manages available templates
type TemplateRegistry struct {
	templates map[string]Template
	fm        *FileManager
	mu        sync.RWMutex
}

// NewTemplateRegistry creates a new registry
func NewTemplateRegistry() *TemplateRegistry {
	return &TemplateRegistry{
		templates: make(map[string]Template),
	}
}

// SetFileManager sets the file manager for the registry
func (r *TemplateRegistry) SetFileManager(fm *FileManager) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fm = fm
}

// Register adds a template to the registry
func (r *TemplateRegistry) Register(template Template) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.templates[template.Name()] = template
}

// Get returns a template by name
func (r *TemplateRegistry) Get(name string) (Template, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	t, exists := r.templates[name]
	if !exists {
		return nil, errors.New("template not found: " + name)
	}

	return t, nil
}

// List returns all available templates
func (r *TemplateRegistry) List() []Template {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Template, 0, len(r.templates))
	for _, t := range r.templates {
		result = append(result, t)
	}

	return result
}

// GetFilesToGenerate returns the list of files to generate based on the provided data
func (r *TemplateRegistry) GetFilesToGenerate(ctx context.Context, templateName string, data *eventdata.TemplateEventData) ([]TemplateFile, error) {
	template, err := r.Get(templateName)
	if err != nil {
		return nil, err
	}

	// Get all possible files from the template
	allFiles := template.Files()

	// If no file manager is set, return all files
	if r.fm == nil {
		return allFiles, nil
	}

	// Filter files based on generation rules
	filesToGenerate := make([]TemplateFile, 0)
	for _, file := range allFiles {
		if r.fm.ShouldGenerate(file.Name, data) {
			filesToGenerate = append(filesToGenerate, file)
		}
	}

	return filesToGenerate, nil
}

// LoadTemplateFiles loads template files from a directory
func (r *TemplateRegistry) LoadTemplateFiles(basePath, templateName string) ([]TemplateFile, error) {
	files := make([]TemplateFile, 0)

	// Function to process each file
	walkFunc := func(path string, info fs.FileInfo, err error) error {
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

		// Calculate relative path from basePath
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path for %s: %w", path, err)
		}

		// Normalize path separators to forward slashes
		relPath = filepath.ToSlash(relPath)

		// Determine if this is a template file
		isTemplate := strings.HasSuffix(relPath, ".tmpl")

		// If it's a template, remove the .tmpl extension
		filePath := relPath
		if isTemplate {
			filePath = strings.TrimSuffix(relPath, ".tmpl")
		}

		// Create template file
		file := TemplateFile{
			Name:           filePath,
			Content:        string(content),
			CodeGeneration: GetFileStrategy(filePath),
			TargetPath:     filePath,
		}

		files = append(files, file)
		return nil
	}

	// Start the walk
	if err := filepath.Walk(basePath, walkFunc); err != nil {
		return nil, fmt.Errorf("failed to walk template directory: %w", err)
	}

	return files, nil
}
