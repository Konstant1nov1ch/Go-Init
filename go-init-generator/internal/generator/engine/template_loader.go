package engine

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// TemplateLoader loads template files from disk
type TemplateLoader struct {
	rootDir   string
	funcMap   template.FuncMap
	templates map[string]*template.Template
	fileCache map[string]string
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(rootDir string) *TemplateLoader {
	return &TemplateLoader{
		rootDir:   rootDir,
		funcMap:   make(template.FuncMap),
		templates: make(map[string]*template.Template),
		fileCache: make(map[string]string),
	}
}

// AddFunc adds a function to the template function map
func (l *TemplateLoader) AddFunc(name string, fn interface{}) {
	l.funcMap[name] = fn
}

// LoadTemplates loads all templates from the root directory
func (l *TemplateLoader) LoadTemplates() error {
	// Validate root directory
	info, err := os.Stat(l.rootDir)
	if err != nil {
		return fmt.Errorf("failed to access template root directory: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", l.rootDir)
	}

	// Walk through all files in the directory
	return filepath.Walk(l.rootDir, l.walkFunc)
}

// walkFunc processes each file during directory walk
func (l *TemplateLoader) walkFunc(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	// Skip directories
	if info.IsDir() {
		return nil
	}

	// Skip non-template files
	if !strings.HasSuffix(path, ".tmpl") {
		return nil
	}

	// Read template file content
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", path, err)
	}

	// Calculate template ID relative to root
	relPath, err := filepath.Rel(l.rootDir, path)
	if err != nil {
		return fmt.Errorf("failed to get relative path: %w", err)
	}

	// Normalize path separators
	relPath = filepath.ToSlash(relPath)

	// Store template content in cache
	l.fileCache[relPath] = string(content)

	// Parse the template
	tmpl, err := template.New(relPath).Funcs(l.funcMap).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", relPath, err)
	}

	// Store the parsed template
	l.templates[relPath] = tmpl

	return nil
}

// GetTemplate retrieves a parsed template by its path
func (l *TemplateLoader) GetTemplate(path string) (*template.Template, error) {
	tmpl, exists := l.templates[path]
	if !exists {
		return nil, fmt.Errorf("template not found: %s", path)
	}
	return tmpl, nil
}

// RenderTemplate renders a template with the provided data
func (l *TemplateLoader) RenderTemplate(path string, data interface{}) (string, error) {
	// Get the template
	tmpl, err := l.GetTemplate(path)
	if err != nil {
		return "", err
	}

	// Render the template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to render template %s: %w", path, err)
	}

	return buf.String(), nil
}

// GetTemplateContent returns the raw template content
func (l *TemplateLoader) GetTemplateContent(path string) (string, bool) {
	content, exists := l.fileCache[path]
	return content, exists
}

// ListTemplates returns a list of all loaded templates
func (l *TemplateLoader) ListTemplates() []string {
	result := make([]string, 0, len(l.templates))
	for path := range l.templates {
		result = append(result, path)
	}
	return result
}

// Reload clears the cache and reloads all templates
func (l *TemplateLoader) Reload() error {
	l.templates = make(map[string]*template.Template)
	l.fileCache = make(map[string]string)
	return l.LoadTemplates()
}
