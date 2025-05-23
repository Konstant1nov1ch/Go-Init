package engine

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-init-gen/internal/eventdata"
)

func TestGenerator(t *testing.T) {
	// Set custom template directory for test
	originalTemplateDir := os.Getenv("TEMPLATE_DIR")
	defer func() {
		// Restore original template dir after test
		if originalTemplateDir != "" {
			os.Setenv("TEMPLATE_DIR", originalTemplateDir)
		} else {
			os.Unsetenv("TEMPLATE_DIR")
		}
	}()

	// Force debug archive saving for testing
	originalDebugArchive := os.Getenv("GENERATOR_SAVE_ARCHIVE_LOCALLY")
	os.Setenv("GENERATOR_SAVE_ARCHIVE_LOCALLY", "true")
	defer func() {
		// Restore original setting after test
		if originalDebugArchive != "" {
			os.Setenv("GENERATOR_SAVE_ARCHIVE_LOCALLY", originalDebugArchive)
		} else {
			os.Unsetenv("GENERATOR_SAVE_ARCHIVE_LOCALLY")
		}
	}()

	// Create debug directory - this will be preserved after the test
	debugDir := filepath.Join("..", "debug_archives")
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		t.Fatalf("Failed to create debug directory: %v", err)
	}

	// Set template directory to the actual templates path
	templatesPath := filepath.Join("..", "templates", "microservices")
	absPath, err := filepath.Abs(templatesPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for templates: %v", err)
	}
	os.Setenv("TEMPLATE_DIR", absPath)
	t.Logf("Using templates directory: %s", absPath)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create a test template with a unique ID
	templateID := fmt.Sprintf("test_%d", time.Now().UnixNano())
	template := createTestTemplate()
	template.ID = templateID

	// Create a new generator
	gen := New()

	// Manually set debug directory for test and pass it the correct path
	gen.debugArchives = true
	gen.debugDir = debugDir

	// Generate the archive
	archive, err := gen.Generate(ctx, &template)
	if err != nil {
		t.Fatalf("Failed to generate template: %v", err)
	}

	// Check that we got a non-empty archive
	if len(archive) == 0 {
		t.Error("Generated archive is empty")
	}

	// Save the archive to the debug directory for verification
	archivePath := filepath.Join(debugDir, fmt.Sprintf("template_%s.zip", templateID))
	if err := os.WriteFile(archivePath, archive, 0o644); err != nil {
		t.Fatalf("Failed to write archive to debug directory: %v", err)
	}

	// Verify the archive exists and has content
	fileInfo, err := os.Stat(archivePath)
	if err != nil {
		t.Fatalf("Failed to stat archive file: %v", err)
	}

	if fileInfo.Size() == 0 {
		t.Error("Archive file exists but is empty")
	} else {
		t.Logf("Archive successfully generated at %s (%d bytes)", archivePath, fileInfo.Size())
	}

	t.Log("Generator test completed successfully")
}

// createTestTemplate creates a test template with sample data
func createTestTemplate() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "test123",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name: "test-service",
			Endpoints: []*eventdata.EndpointEventData{
				{
					Protocol: "GRPC",
					Role:     "server",
					Config: map[string]string{
						"service": "TestService",
					},
				},
				{
					Protocol: "GRAPHQL",
					Role:     "server",
					Config: map[string]string{
						"schema": "type Query { test: String }",
					},
				},
			},
			Database: eventdata.DatabaseEventData{
				Type:       "postgresql",
				DDL:        "CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT);",
				Migrations: true,
				Models:     true,
			},
			Docker: eventdata.DockerEventData{
				Registry:  "docker.io",
				ImageName: "test-service",
			},
			Advanced: &eventdata.AdvancedEventData{
				EnableAuthentication: true,
				GenerateSwaggerDocs:  false,
				ModulePath:           "github.com/example/test-service",
				ServiceDescription:   "A test microservice with gRPC and GraphQL",
				EnableGraphQL:        true,
				EnableGRPC:           true,
			},
		},
	}
}
