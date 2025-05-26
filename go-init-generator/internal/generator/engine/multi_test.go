package engine

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"go-init-gen/internal/eventdata"
)

// TestMultiVariantGeneration тестирует генерацию трех различных вариантов сервиса
func TestMultiVariantGeneration(t *testing.T) {
	// Создаём директорию debug_archives, если её нет
	debugDir := filepath.Join("..", "debug_archives")
	if err := os.MkdirAll(debugDir, 0o755); err != nil {
		t.Fatalf("Failed to create debug directory: %v", err)
	}

	// Устанавливаем переменную окружения для сохранения архивов
	os.Setenv("GENERATOR_SAVE_ARCHIVE_LOCALLY", "true")
	defer os.Unsetenv("GENERATOR_SAVE_ARCHIVE_LOCALLY")

	// Настраиваем путь к шаблонам
	originalTemplateDir := os.Getenv("TEMPLATE_DIR")
	templatesPath := filepath.Join("..", "templates", "microservices")
	absPath, err := filepath.Abs(templatesPath)
	if err != nil {
		t.Fatalf("Failed to get absolute path for templates: %v", err)
	}
	os.Setenv("TEMPLATE_DIR", absPath)
	defer func() {
		if originalTemplateDir != "" {
			os.Setenv("TEMPLATE_DIR", originalTemplateDir)
		} else {
			os.Unsetenv("TEMPLATE_DIR")
		}
	}()

	// Создаём контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Тестовые случаи
	testCases := []struct {
		name       string
		template   eventdata.ProcessTemplate
		outputPath string
	}{
		{
			name:       "Вариант 1: gRPC + GraphQL + PostgreSQL",
			template:   createTestVariant1(),
			outputPath: filepath.Join(debugDir, "template_variant1.zip"),
		},
		{
			name:       "Вариант 2: GraphQL + PostgreSQL",
			template:   createTestVariant2(),
			outputPath: filepath.Join(debugDir, "template_variant2.zip"),
		},
		{
			name:       "Вариант 3: Минимальный сервис",
			template:   createTestVariant3(),
			outputPath: filepath.Join(debugDir, "template_variant3.zip"),
		},
	}

	// Создаём генератор
	gen := New()
	gen.SetDebugDir(debugDir)

	// Запускаем тесты для каждого варианта
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Генерируем архив
			archive, err := gen.Generate(ctx, &tc.template)
			if err != nil {
				t.Fatalf("Failed to generate template: %v", err)
			}

			// Проверяем, что архив не пустой
			if len(archive) == 0 {
				t.Error("Generated archive is empty")
			}

			// Сохраняем архив локально
			if err := os.WriteFile(tc.outputPath, archive, 0o644); err != nil {
				t.Fatalf("Failed to write archive to disk: %v", err)
			}

			// Проверяем, что файл создан и не пустой
			fileInfo, err := os.Stat(tc.outputPath)
			if err != nil {
				t.Fatalf("Failed to stat archive file: %v", err)
			}

			if fileInfo.Size() == 0 {
				t.Error("Archive file exists but is empty")
			} else {
				t.Logf("Archive successfully generated at %s (%d bytes)", tc.outputPath, fileInfo.Size())
			}
		})
	}
}

// createTestVariant1 создаёт шаблон для варианта 1 (gRPC + GraphQL + PostgreSQL)
func createTestVariant1() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant1",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name: "test-service1",
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

// createTestVariant2 создаёт шаблон для варианта 2 (GraphQL + PostgreSQL)
func createTestVariant2() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant2",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name: "test-service2",
			Endpoints: []*eventdata.EndpointEventData{
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
				ServiceDescription:   "A test microservice with GraphQL",
				EnableGraphQL:        true,
				EnableGRPC:           false,
			},
		},
	}
}

// createTestVariant3 создаёт шаблон для варианта 3 (минимальный сервис)
func createTestVariant3() eventdata.ProcessTemplate {
	return eventdata.ProcessTemplate{
		ID:     "variant3",
		Status: "PROCESSING",
		Data: eventdata.TemplateEventData{
			Name:      "test-service3",
			Endpoints: []*eventdata.EndpointEventData{},
			Database:  eventdata.DatabaseEventData{},
			Docker: eventdata.DockerEventData{
				Registry:  "docker.io",
				ImageName: "test-service",
			},
			Advanced: &eventdata.AdvancedEventData{
				EnableAuthentication: false,
				GenerateSwaggerDocs:  false,
				ModulePath:           "github.com/example/test-service",
				ServiceDescription:   "A minimal microservice",
				EnableGraphQL:        false,
				EnableGRPC:           false,
			},
		},
	}
}
