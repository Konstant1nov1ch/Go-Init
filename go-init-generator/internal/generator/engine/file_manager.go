package engine

import (
	"go-init-gen/internal/eventdata"
)

// FileFilter defines a filter for determining whether a file should be generated
type FileFilter interface {
	// ShouldGenerate returns true if the file should be generated based on the provided configuration
	ShouldGenerate(filePath string, data *eventdata.TemplateEventData) bool
}

// FileManager handles decisions about which files to generate
type FileManager struct {
	fileFilters []FileFilter
}

// NewFileManager creates a new file manager instance
func NewFileManager() *FileManager {
	manager := &FileManager{
		fileFilters: make([]FileFilter, 0),
	}
	return manager
}

// RegisterFilter adds a file filter to the manager
func (m *FileManager) RegisterFilter(filter FileFilter) {
	m.fileFilters = append(m.fileFilters, filter)
}

// ShouldGenerate determines if the file should be generated based on the provided configuration
func (m *FileManager) ShouldGenerate(filePath string, data *eventdata.TemplateEventData) bool {
	// If no filters are registered, generate all files
	if len(m.fileFilters) == 0 {
		return true
	}

	// Apply filters
	for _, filter := range m.fileFilters {
		if !filter.ShouldGenerate(filePath, data) {
			return false
		}
	}

	// If we passed all filters, generate the file
	return true
}
