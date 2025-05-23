package postprocessors

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

// ZipArchiver creates a ZIP archive from generated files
type ZipArchiver struct{}

// Process implements the Postprocessor interface
func (a *ZipArchiver) Process(ctx context.Context, data []byte) ([]byte, error) {
	// Check if the input is already a ZIP file (starts with PK\x03\x04)
	if isZipFile(data) {
		return data, nil
	}

	// Try to parse as JSON map of file paths to contents
	var fileMap map[string]string
	if err := json.Unmarshal(data, &fileMap); err != nil {
		// If not JSON, it's likely the content of a single file or another format
		// In this case, create a simple archive with a single file
		return createSingleFileArchive("content.txt", data)
	}

	// Create a proper zip archive with multiple files
	return createMultiFileArchive(fileMap)
}

// CreateArchiver creates a new archiver postprocessor
func CreateArchiver() *ZipArchiver {
	return &ZipArchiver{}
}

// Archiver creates a ZIP archive from generated files
type Archiver struct {
	timestamp time.Time
}

// NewArchiver creates a new archiver
func NewArchiver() *Archiver {
	return &Archiver{
		timestamp: time.Now(),
	}
}

// Process implements the Postprocessor interface
func (a *Archiver) Process(ctx context.Context, input []byte) ([]byte, error) {
	// Check if the input is already a ZIP file
	if isZipFile(input) {
		return input, nil
	}

	// Try to decode JSON (from old template format or other sources)
	var fileMap map[string]string
	if err := json.Unmarshal(input, &fileMap); err != nil {
		return nil, fmt.Errorf("failed to parse file map: %w", err)
	}

	// Convert string content to bytes
	fileBytes := make(map[string][]byte)
	for path, content := range fileMap {
		fileBytes[path] = []byte(content)
	}

	// Create a ZIP archive
	return a.CreateArchive(fileBytes)
}

// CreateArchive creates a ZIP archive from a map of filenames to contents
func (a *Archiver) CreateArchive(files map[string][]byte) ([]byte, error) {
	// Create an in-memory buffer for the ZIP file
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Add each file to the ZIP archive
	for filename, content := range files {
		// Normalize file path (use forward slashes and trim leading slashes)
		normalizedPath := filepath.ToSlash(filename)
		normalizedPath = strings.TrimPrefix(normalizedPath, "/")

		// Create a file in the archive
		fileHeader := &zip.FileHeader{
			Name:     normalizedPath,
			Method:   zip.Deflate,
			Modified: a.timestamp,
		}
		fileWriter, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in archive: %w", err)
		}

		// Write the file content
		if _, err := fileWriter.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write file content: %w", err)
		}
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Helper functions

// isZipFile checks if the data is already a ZIP file
func isZipFile(data []byte) bool {
	// ZIP files start with PK\x03\x04
	return len(data) > 4 && data[0] == 0x50 && data[1] == 0x4B && data[2] == 0x03 && data[3] == 0x04
}

// createSingleFileArchive creates a ZIP archive with a single file
func createSingleFileArchive(filename string, content []byte) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Create a file in the archive
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create file in archive: %w", err)
	}

	// Write the file content
	if _, err := writer.Write(content); err != nil {
		return nil, fmt.Errorf("failed to write file content: %w", err)
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}

// createMultiFileArchive creates a ZIP archive from a map of filenames to contents
func createMultiFileArchive(fileMap map[string]string) ([]byte, error) {
	// Create an in-memory buffer for the ZIP file
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Current timestamp for all files
	now := time.Now()

	// Add each file to the ZIP archive
	for filename, content := range fileMap {
		if filename == "" {
			continue
		}

		// Normalize file path (use forward slashes and trim leading slashes)
		normalizedPath := filepath.ToSlash(filename)
		normalizedPath = strings.TrimPrefix(normalizedPath, "/")

		// Create a file in the archive
		fileHeader := &zip.FileHeader{
			Name:     normalizedPath,
			Method:   zip.Deflate,
			Modified: now,
		}

		// Set file mode to regular file with read/write permissions
		fileHeader.SetMode(0o644)

		fileWriter, err := zipWriter.CreateHeader(fileHeader)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in archive: %w", err)
		}

		// Write the file content
		if _, err := fileWriter.Write([]byte(content)); err != nil {
			return nil, fmt.Errorf("failed to write file content: %w", err)
		}
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	return buf.Bytes(), nil
}
