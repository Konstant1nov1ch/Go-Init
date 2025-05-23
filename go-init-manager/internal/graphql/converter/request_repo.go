package converter

import (
	"context"

	"gitlab.com/go-init/go-init-common/default/logger"

	dbModel "go-init/internal/database/request_repo/models"

	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
)

func FromInputToDbServiceTemplate(
	ctx context.Context,
	input model.CreateTemplateInput,
	log *logger.Logger,
) (*dbModel.ServiceTemplate, error) {
	// Инициализируем пустой URL для архива
	emptyURL := "" // Будет установлен позже, когда архив будет сгенерирован и загружен

	// Create a new UUID for the template
	templateUUID := uuid.New()

	// TODO: In a production environment, you should extract the user ID from
	// the authentication context. This is a temporary solution.
	defaultUserID := uuid.New()
	status := "pending"

	// Name is required by schema, so we can directly use it
	name := input.Name

	// API версия - если поддерживается в GraphQL модели
	version := "1.0" // Default version

	// Create template with required fields
	template := &dbModel.ServiceTemplate{
		ServiceTemplateUuid: &templateUUID,
		ServiceTemplateName: &name,
		ZipURL:              &emptyURL,
		UserId:              &defaultUserID,
		Status:              &status,
		Version:             &version,
	}

	// Conditionally add optional fields if they are provided

	// Add endpoints if provided (optional in schema)
	if input.Endpoints != nil && len(input.Endpoints) > 0 {
		template.Endpoints = convertEndpoints(input.Endpoints)
	}

	// Add database config if provided (optional in schema)
	if input.Database != nil {
		template.DatabaseConfigs = convertDatabase(input.Database)
	}

	// Add docker config if provided (optional in schema)
	if input.Docker != nil {
		template.DockerConfigs = convertDocker(input.Docker)
	}

	// Add advanced config if provided (optional in schema)
	if input.Advanced != nil {
		template.AdvancedConfigs = convertAdvanced(input.Advanced)
	}

	return template, nil
}

// Helper functions to convert input types to dbModel types
func convertEndpoints(inputs []*model.EndpointInput) []*dbModel.Endpoint {
	var endpoints []*dbModel.Endpoint
	for _, input := range inputs {
		if input != nil {
			protocol := input.Protocol.String()
			role := input.Role.String()
			endpoints = append(endpoints, &dbModel.Endpoint{
				Protocol: &protocol,
				Role:     &role,
			})
		}
	}
	return endpoints
}

func convertDatabase(input *model.DatabaseInput) []*dbModel.DatabaseConfig {
	if input == nil {
		return nil
	}
	typeStr := input.Type.String()
	return []*dbModel.DatabaseConfig{
		{
			Type: &typeStr,
			DDL:  input.Ddl,
		},
	}
}

func convertDocker(input *model.DockerInput) []*dbModel.DockerConfig {
	if input == nil {
		return nil
	}

	imageName := ""
	if input.ImageName != "" {
		imageName = input.ImageName
	}

	return []*dbModel.DockerConfig{
		{
			Registry:  input.Registry,
			ImageName: &imageName,
		},
	}
}

func convertAdvanced(input *model.AdvancedInput) []*dbModel.AdvancedConfig {
	if input == nil {
		return nil
	}
	return []*dbModel.AdvancedConfig{
		{
			EnableAuthentication: input.EnableAuthentication,
			GenerateSwaggerDocs:  input.GenerateSwaggerDocs,
		},
	}
}

// ConvertToEndpointConfig converts EndpointInput to EndpointConfig
func ConvertToEndpointConfig(inputs []*model.EndpointInput) []*model.EndpointConfig {
	var configs []*model.EndpointConfig
	for _, input := range inputs {
		configs = append(configs, &model.EndpointConfig{
			Protocol: input.Protocol,
			Role:     input.Role,
		})
	}
	return configs
}

// ConvertToDatabaseConfig converts DatabaseInput to DatabaseConfig
func ConvertToDatabaseConfig(input *model.DatabaseInput) *model.DatabaseConfig {
	if input == nil {
		return nil
	}
	return &model.DatabaseConfig{
		Type: input.Type,
		Ddl:  input.Ddl,
	}
}

// ConvertToDockerConfig converts DockerInput to DockerConfig
func ConvertToDockerConfig(input *model.DockerInput) *model.DockerConfig {
	if input == nil {
		return nil
	}
	return &model.DockerConfig{
		Registry:  input.Registry,
		ImageName: input.ImageName,
	}
}

// ConvertToAdvancedConfig converts AdvancedInput to AdvancedConfig
func ConvertToAdvancedConfig(input *model.AdvancedInput) *model.AdvancedConfig {
	if input == nil {
		return nil
	}
	return &model.AdvancedConfig{
		EnableAuthentication: input.EnableAuthentication,
		GenerateSwaggerDocs:  input.GenerateSwaggerDocs,
	}
}
