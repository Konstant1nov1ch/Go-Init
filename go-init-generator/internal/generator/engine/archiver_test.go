package engine

import (
	"archive/zip"
	"bytes"
	"testing"
)

func TestArchiver_CreateDirectories(t *testing.T) {
	archiver := NewArchiver(false, "")

	// Test files that would create directories
	files := map[string][]byte{
		"pkg/api/grpc/README.md":      []byte("# gRPC directory"),
		"pkg/api/graphql/README.md":   []byte("# GraphQL directory"),
		"internal/service/service.go": []byte("package service"),
		"cmd/main.go":                 []byte("package main"),
	}

	// Create archive
	archiveBytes, err := archiver.CreateArchive(files, "test")
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	// Read the archive and check for directories
	reader, err := zip.NewReader(bytes.NewReader(archiveBytes), int64(len(archiveBytes)))
	if err != nil {
		t.Fatalf("Failed to read archive: %v", err)
	}

	// Track directories found
	directoriesFound := make(map[string]bool)
	filesFound := make(map[string]bool)

	for _, file := range reader.File {
		if file.Name[len(file.Name)-1] == '/' {
			// This is a directory
			directoriesFound[file.Name[:len(file.Name)-1]] = true
			t.Logf("Found directory: %s", file.Name)
		} else {
			// This is a file
			filesFound[file.Name] = true
			t.Logf("Found file: %s", file.Name)
		}
	}

	// Check that expected directories were created
	expectedDirs := []string{
		"pkg",
		"pkg/api",
		"pkg/api/grpc",
		"pkg/api/graphql",
		"internal",
		"internal/service",
		"cmd",
	}

	for _, expectedDir := range expectedDirs {
		if !directoriesFound[expectedDir] {
			t.Errorf("Expected directory %s was not found in archive", expectedDir)
		}
	}

	// Check that all files are present
	expectedFiles := []string{
		"pkg/api/grpc/README.md",
		"pkg/api/graphql/README.md",
		"internal/service/service.go",
		"cmd/main.go",
	}

	for _, expectedFile := range expectedFiles {
		if !filesFound[expectedFile] {
			t.Errorf("Expected file %s was not found in archive", expectedFile)
		}
	}
}
