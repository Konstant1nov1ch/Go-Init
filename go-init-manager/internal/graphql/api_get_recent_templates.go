package graphql

import (
	"context"

	"go-init/internal/graphql/converter"
	"go-init/pkg/api/graphql/model"
)

// GetRecentTemplates retrieves a list of recently created templates
func (s *Service) GetRecentTemplates(ctx context.Context, limit *int) (*model.TemplatesResponse, error) {
	// Set default limit if not provided
	limitValue := 5
	if limit != nil {
		limitValue = *limit
	}

	// Get templates from repository
	templates, err := s.dbManagerRepo.GetRecentTemplates(ctx, limitValue)
	if err != nil {
		return &model.TemplatesResponse{
			Success: false,
			Message: strPtr(err.Error()),
		}, nil
	}

	// Convert database models to GraphQL models
	graphqlTemplates := make([]*model.ServiceTemplate, 0, len(templates))
	for _, template := range templates {
		graphqlTemplate := converter.DbTemplateToGraphqlTemplate(template)
		if graphqlTemplate != nil {
			graphqlTemplates = append(graphqlTemplates, graphqlTemplate)
		}
	}

	return &model.TemplatesResponse{
		Success:   true,
		Message:   strPtr("Templates retrieved successfully"),
		Templates: graphqlTemplates,
	}, nil
}
