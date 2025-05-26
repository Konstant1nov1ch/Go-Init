package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// TemplateLoader handles loading template files from the filesystem
type TemplateLoader struct {
	templateDir string
}

// NewTemplateLoader creates a new template loader
func NewTemplateLoader(templateDir string) *TemplateLoader {
	return &TemplateLoader{
		templateDir: templateDir,
	}
}

// LoadTemplateFiles loads all template files from the template directory
func (tl *TemplateLoader) LoadTemplateFiles() ([]TemplateFile, error) {
	var files []TemplateFile

	// Walk the template directory
	err := filepath.Walk(tl.templateDir, func(path string, info os.FileInfo, err error) error {
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
		relPath, err := filepath.Rel(tl.templateDir, path)
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
