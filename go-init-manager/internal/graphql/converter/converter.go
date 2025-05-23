package converter

import (
	"strconv"

	"go-init/pkg/api/graphql/model"

	dbModels "go-init/internal/database/request_repo/models"
)

// DbTemplateToGraphqlTemplate converts a database template model to a GraphQL template model
func DbTemplateToGraphqlTemplate(dbTemplate *dbModels.ServiceTemplate) *model.ServiceTemplate {
	if dbTemplate == nil {
		return nil
	}

	// Convert ID to string
	id := strconv.Itoa(*dbTemplate.ServiceTemplateId)

	// The zipUrl might be optional now in the GraphQL schema
	var zipURL *string
	if dbTemplate.ZipURL != nil && *dbTemplate.ZipURL != "" {
		strVal := *dbTemplate.ZipURL
		zipURL = &strVal
	}

	// Name is still required
	name := ""
	if dbTemplate.ServiceTemplateName != nil {
		name = *dbTemplate.ServiceTemplateName
	}

	// Version is optional
	var version *string
	if dbTemplate.Version != nil {
		version = dbTemplate.Version
	}

	// Initialize minimum required fields
	template := &model.ServiceTemplate{
		ID:      id,
		Name:    name,
		ZipURL:  zipURL,
		Version: version,
	}

	// Convert status field if present
	if dbTemplate.Status != nil {
		switch *dbTemplate.Status {
		case "PENDING", "pending":
			pending := model.TemplateStatusPending
			template.Status = &pending
		case "PROCESSING", "processing":
			processing := model.TemplateStatusProcessing
			template.Status = &processing
		case "COMPLETED", "completed":
			completed := model.TemplateStatusCompleted
			template.Status = &completed
		case "FAILED", "failed":
			failed := model.TemplateStatusFailed
			template.Status = &failed
		}
	} else {
		// Default status is PENDING if none specified
		pending := model.TemplateStatusPending
		template.Status = &pending
	}

	// Set error field if present
	if dbTemplate.Error != nil && *dbTemplate.Error != "" {
		template.Error = dbTemplate.Error
	}

	// Convert CreatedAt to string if it exists
	if dbTemplate.CreatedAt != nil {
		template.CreatedAt = dbTemplate.CreatedAt.Format("2006-01-02T15:04:05Z")
	}

	// Convert UpdatedAt to string if it exists
	if dbTemplate.UpdatedAt != nil {
		template.UpdatedAt = &[]string{dbTemplate.UpdatedAt.Format("2006-01-02T15:04:05Z")}[0]
	}

	// Convert Endpoint configs - now optional
	if len(dbTemplate.Endpoints) > 0 {
		template.Endpoints = make([]*model.EndpointConfig, 0, len(dbTemplate.Endpoints))

		for _, endpoint := range dbTemplate.Endpoints {
			if endpoint != nil {
				endpointConfig := &model.EndpointConfig{}

				// Convert Protocol
				if endpoint.Protocol != nil {
					switch *endpoint.Protocol {
					case "GRPC":
						endpointConfig.Protocol = model.ServiceProtocolGrpc
					case "REST":
						endpointConfig.Protocol = model.ServiceProtocolRest
					case "GRAPHQL":
						endpointConfig.Protocol = model.ServiceProtocolGraphql
					}
				}

				// Convert Role
				if endpoint.Role != nil {
					switch *endpoint.Role {
					case "CLIENT":
						endpointConfig.Role = model.ServiceRoleClient
					case "SERVER":
						endpointConfig.Role = model.ServiceRoleServer
					}
				}

				template.Endpoints = append(template.Endpoints, endpointConfig)
			}
		}
	}

	// Convert Database config - now optional
	if len(dbTemplate.DatabaseConfigs) > 0 && dbTemplate.DatabaseConfigs[0] != nil {
		dbConfig := dbTemplate.DatabaseConfigs[0]
		databaseConfig := &model.DatabaseConfig{}

		// Convert Type
		if dbConfig.Type != nil {
			switch *dbConfig.Type {
			case "POSTGRESQL":
				databaseConfig.Type = model.DatabaseTypePostgresql
			case "MYSQL":
				databaseConfig.Type = model.DatabaseTypeMysql
			case "NONE":
				databaseConfig.Type = model.DatabaseTypeNone
			}
		}

		// Convert DDL
		if dbConfig.DDL != nil {
			databaseConfig.Ddl = dbConfig.DDL
		}

		template.Database = databaseConfig
	}

	// Convert Docker config - now optional
	if len(dbTemplate.DockerConfigs) > 0 && dbTemplate.DockerConfigs[0] != nil {
		dockerConfig := &model.DockerConfig{}
		dbDockerConfig := dbTemplate.DockerConfigs[0]

		// Convert Registry
		if dbDockerConfig.Registry != nil {
			dockerConfig.Registry = dbDockerConfig.Registry
		}

		// Convert ImageName - might still be required in Docker config
		if dbDockerConfig.ImageName != nil {
			dockerConfig.ImageName = *dbDockerConfig.ImageName
		}

		template.Docker = dockerConfig
	}

	// Convert Advanced config - already optional
	if len(dbTemplate.AdvancedConfigs) > 0 && dbTemplate.AdvancedConfigs[0] != nil {
		advancedConfig := &model.AdvancedConfig{}
		dbAdvancedConfig := dbTemplate.AdvancedConfigs[0]

		// Convert EnableAuthentication
		if dbAdvancedConfig.EnableAuthentication != nil {
			advancedConfig.EnableAuthentication = dbAdvancedConfig.EnableAuthentication
		}

		// Convert GenerateSwaggerDocs
		if dbAdvancedConfig.GenerateSwaggerDocs != nil {
			advancedConfig.GenerateSwaggerDocs = dbAdvancedConfig.GenerateSwaggerDocs
		}

		template.Advanced = advancedConfig
	}

	return template
}
