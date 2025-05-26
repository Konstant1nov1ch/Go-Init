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

// TestGenerateServiceVariants tests generation of different service variants as described in examples.md
func TestGenerateServiceVariants(t *testing.T) {
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

	// Set template directory to the actual templates path
	templatesPath := filepath.Join("..", "templates", "microservices")
	absPath, err := filepath.Abs(templatesPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for templates: %v", err)
	}
	os.Setenv("TEMPLATE_DIR", absPath)
	t.Logf("Using templates directory: %s", absPath)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Create debug variants directory - this will be preserved after the test
	variantsDir := filepath.Join("..", "debug_archives", "variants")
	if err := os.MkdirAll(variantsDir, 0o755); err != nil {
		t.Fatalf("Failed to create variants directory: %v", err)
	}

	// Define the three variants
	variants := []struct {
		name     string
		template eventdata.ProcessTemplate
	}{
		{
			name:     "variant1",
			template: createVariant1Template(),
		},
		{
			name:     "variant2",
			template: createVariant2Template(),
		},
		{
			name:     "variant3",
			template: createVariant3Template(),
		},
	}

	// Generate each variant
	for _, variant := range variants {
		t.Run(variant.name, func(t *testing.T) {
			// Create unique ID
			templateID := fmt.Sprintf("%s_%d", variant.name, time.Now().UnixNano())
			variant.template.ID = templateID

			// Set custom debug directory for this variant
			variantDir := filepath.Join(variantsDir, variant.name)
			if err := os.MkdirAll(variantDir, 0o755); err != nil {
				t.Fatalf("Failed to create variant directory: %v", err)
			}

			// Create a new generator
			gen := New()

			// Manually set debug directory for test using the new methods
			gen.SetDebugArchives(true)
			gen.SetDebugDir(variantDir)

			// Generate the archive
			archive, err := gen.Generate(ctx, &variant.template)
			if err != nil {
				t.Fatalf("Failed to generate template: %v", err)
			}

			// Check that we got a non-empty archive
			if len(archive) == 0 {
				t.Error("Generated archive is empty")
			}

			// Save the archive to the variant directory for verification
			archivePath := filepath.Join(variantDir, fmt.Sprintf("template_%s.zip", templateID))
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
		})
	}

	t.Log("Service variants generation test completed successfully")
}

// createVariant1Template creates variant 1 - gRPC + GraphQL + PostgreSQL
func createVariant1Template() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant1",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name: "test-service1",
			Endpoints: []*eventdata.EndpointEventData{
				{
					Protocol: "GRPC",
					Role:     "SERVER",
					Config: map[string]string{
						"service": "TestService",
					},
				},
				{
					Protocol: "GRAPHQL",
					Role:     "SERVER",
					Config: map[string]string{
						"schema": "type Query { test: String }",
					},
				},
			},
			Database: eventdata.DatabaseEventData{
				Type:       "POSTGRESQL",
				DDL:        "CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT);",
				Migrations: true,
				Models:     true,
			},
			Docker: eventdata.DockerEventData{
				Registry:  "docker.io",
				ImageName: "test-service1",
			},
			Advanced: &eventdata.AdvancedEventData{
				EnableAuthentication: true,
				GenerateSwaggerDocs:  false,
				ModulePath:           "github.com/example/test-service1",
				ServiceDescription:   "A test microservice with gRPC and GraphQL",
				EnableGraphQL:        true,
				EnableGRPC:           true,
			},
		},
	}
}

// createVariant2Template creates variant 2 - GraphQL + PostgreSQL only
func createVariant2Template() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant2",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name: "test-service2",
			Endpoints: []*eventdata.EndpointEventData{
				{
					Protocol: "GRAPHQL",
					Role:     "SERVER",
					Config: map[string]string{
						"schema": "type Query { test: String }",
					},
				},
			},
			Database: eventdata.DatabaseEventData{
				Type:       "POSTGRESQL",
				DDL:        "CREATE TABLE test (id SERIAL PRIMARY KEY, name TEXT);",
				Migrations: true,
				Models:     true,
			},
			Docker: eventdata.DockerEventData{
				Registry:  "docker.io",
				ImageName: "test-service2",
			},
			Advanced: &eventdata.AdvancedEventData{
				EnableAuthentication: true,
				GenerateSwaggerDocs:  false,
				ModulePath:           "github.com/example/test-service2",
				ServiceDescription:   "A test microservice with GraphQL",
				EnableGraphQL:        true,
				EnableGRPC:           false,
			},
		},
	}
}

// createVariant3Template creates variant 3 - minimal service with no endpoints or database
func createVariant3Template() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant3",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name:      "test-service3",
			Endpoints: []*eventdata.EndpointEventData{},
			Database:  eventdata.DatabaseEventData{},
			Docker: eventdata.DockerEventData{
				Registry:  "docker.io",
				ImageName: "test-service3",
			},
			Advanced: &eventdata.AdvancedEventData{
				EnableAuthentication: false,
				GenerateSwaggerDocs:  false,
				ModulePath:           "github.com/example/test-service3",
				ServiceDescription:   "A minimal microservice",
				EnableGraphQL:        false,
				EnableGRPC:           false,
			},
		},
	}
}
