package graphql

import (
	"context"
	"fmt"
	"strconv"

	dbModel "go-init/internal/database/request_repo/models"
	"go-init/internal/graphql/converter"
	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
)

const statusDone = "Done"

func (s *Service) GetTemplate(ctx context.Context, id string) (*model.TemplateResponse, error) {
	s.logger.Info("Getting template by ID: " + id)

	// Сначала пробуем как UUID
	if templateUUID, err := uuid.Parse(id); err == nil {
		template, err := s.dbManagerRepo.GetTemplateByUUID(ctx, templateUUID)
		if err != nil {
			return &model.TemplateResponse{
				Success: false,
				Message: strPtr(fmt.Sprintf("Template not found: %v", err)),
			}, nil
		}
		return createSuccessResponse(template), nil
	}

	// Иначе — пытаемся считать как int
	templateID, err := strconv.Atoi(id)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr("Invalid template ID format"),
		}, nil
	}

	template, err := s.dbManagerRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr(fmt.Sprintf("Template not found: %v", err)),
		}, nil
	}

	return createSuccessResponse(template), nil
}

// Helper function to create a pointer to a string
func strPtr(s string) *string {
	return &s
}

// Helper function to create a success response with template
func createSuccessResponse(template *dbModel.ServiceTemplate) *model.TemplateResponse {
	return &model.TemplateResponse{
		Success:  true,
		Message:  strPtr("Template retrieved successfully"),
		Template: converter.DbTemplateToGraphqlTemplate(template),
	}
}
