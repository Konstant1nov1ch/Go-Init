package engine

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Archiver handles creation of ZIP archives from generated files
type Archiver struct {
	debugArchives bool
	debugDir      string
}

// NewArchiver creates a new archiver
func NewArchiver(debugArchives bool, debugDir string) *Archiver {
	return &Archiver{
		debugArchives: debugArchives,
		debugDir:      debugDir,
	}
}

// CreateArchive creates a ZIP archive with the generated files
func (a *Archiver) CreateArchive(files map[string][]byte, id string) ([]byte, error) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)

	// Collect all directories that need to be created
	directories := a.collectDirectories(files)

	// Create directories in the archive first
	for dir := range directories {
		// Add trailing slash to indicate directory
		dirPath := dir + "/"
		_, err := zipWriter.Create(dirPath)
		if err != nil {
			return nil, fmt.Errorf("failed to create directory in archive: %w", err)
		}
	}

	// Add files to the archive
	for path, content := range files {
		file, err := zipWriter.Create(path)
		if err != nil {
			return nil, fmt.Errorf("failed to create file in archive: %w", err)
		}

		if _, err := file.Write(content); err != nil {
			return nil, fmt.Errorf("failed to write content to archive: %w", err)
		}
	}

	// Close the writer
	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close archive writer: %w", err)
	}

	// Save debug archive if enabled
	a.saveDebugArchive(buf.Bytes(), id)

	return buf.Bytes(), nil
}

// collectDirectories extracts all unique directory paths from file paths
func (a *Archiver) collectDirectories(files map[string][]byte) map[string]bool {
	directories := make(map[string]bool)

	for filePath := range files {
		// Normalize path separators to forward slashes for ZIP archives
		normalizedPath := strings.ReplaceAll(filePath, "\\", "/")
		dir := filepath.Dir(normalizedPath)

		// Convert back to forward slashes in case filepath.Dir changed them
		dir = strings.ReplaceAll(dir, "\\", "/")

		// Add all parent directories
		for dir != "." && dir != "/" && dir != "" {
			directories[dir] = true
			parentDir := filepath.Dir(dir)
			// Ensure parent dir also uses forward slashes
			parentDir = strings.ReplaceAll(parentDir, "\\", "/")
			dir = parentDir
		}
	}

	return directories
}

// saveDebugArchive saves the archive locally if debug mode is enabled
func (a *Archiver) saveDebugArchive(archive []byte, id string) {
	if !a.debugArchives {
		return
	}

	// Use the specified debug directory or fallback to default
	debugDir := a.debugDir
	if debugDir == "" {
		debugDir = "debug_archives"
	}

	// Ensure directory exists
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		fmt.Printf("Warning: Failed to create debug directory: %v\n", err)
		return
	}

	// Write archive to file
	archivePath := filepath.Join(debugDir, fmt.Sprintf("template_%s.zip", id))
	if err := os.WriteFile(archivePath, archive, 0o644); err != nil {
		fmt.Printf("Warning: Failed to save debug archive: %v\n", err)
		return
	}

	fmt.Printf("Debug archive saved to: %s\n", archivePath)
}
