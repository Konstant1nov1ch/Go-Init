package converter

import (
	"go-init/internal/eventdata"
	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
)

// ConvertToEndpointEventData converts EndpointInput to EndpointEventData
func ConvertToEndpointEventData(inputs []*model.EndpointInput) []*eventdata.EndpointEventData {
	var eventData []*eventdata.EndpointEventData
	for _, input := range inputs {
		eventData = append(eventData, &eventdata.EndpointEventData{
			Protocol: input.Protocol.String(),
			Role:     input.Role.String(),
		})
	}
	return eventData
}

// FromInputToEvent converts CreateTemplateInput to ProcessTemplate event data
func FromInputToEvent(input model.CreateTemplateInput, requestUUID uuid.UUID) eventdata.ProcessTemplate {
	// Log the request UUID for debugging
	event := eventdata.ProcessTemplate{
		ID:     requestUUID.String(),
		Status: "created",
		Data: eventdata.TemplateEventData{
			Name: input.Name,
		},
	}

	// Each of these fields is optional in the GraphQL schema

	// Add endpoints if available
	if input.Endpoints != nil {
		event.Data.Endpoints = ConvertToEndpointEventData(input.Endpoints)
	}

	// Add database config if available
	if input.Database != nil {
		dbType := input.Database.Type.String()
		event.Data.Database = eventdata.DatabaseEventData{
			Type: dbType,
			DDL:  StringValue(input.Database.Ddl, ""),
		}
	} else {
		// Default database type if not provided
		event.Data.Database = eventdata.DatabaseEventData{
			Type: "NONE",
			DDL:  "",
		}
	}

	// Add docker config if available
	if input.Docker != nil {
		event.Data.Docker = eventdata.DockerEventData{
			Registry:  StringValue(input.Docker.Registry, ""),
			ImageName: input.Docker.ImageName,
		}
	} else {
		// Default docker config if not provided
		event.Data.Docker = eventdata.DockerEventData{
			Registry:  "",
			ImageName: "",
		}
	}

	// Add advanced config if available
	if input.Advanced != nil {
		event.Data.Advanced = &eventdata.AdvancedEventData{
			EnableAuthentication: BoolValue(input.Advanced.EnableAuthentication, false),
			GenerateSwaggerDocs:  BoolValue(input.Advanced.GenerateSwaggerDocs, false),
		}
	} else {
		// Default advanced config if not provided
		event.Data.Advanced = &eventdata.AdvancedEventData{
			EnableAuthentication: false,
			GenerateSwaggerDocs:  false,
		}
	}

	return event
}

// StringValue returns the value of the string pointer or a default value if nil.
func StringValue(ptr *string, defaultValue string) string {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}

// BoolValue returns the value of the bool pointer or a default value if nil.
func BoolValue(ptr *bool, defaultValue bool) bool {
	if ptr != nil {
		return *ptr
	}
	return defaultValue
}
