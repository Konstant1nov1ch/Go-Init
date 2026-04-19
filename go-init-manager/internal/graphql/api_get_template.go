package graphql

import (
	"context"
	"fmt"
	"strconv"

	dbModel "go-init/internal/database/request_repo/models"
	"go-init/internal/graphql/converter"
	"go-init/internal/tracing"
	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

const statusDone = "Done"

func (s *Service) GetTemplate(ctx context.Context, id string) (*model.TemplateResponse, error) {
	s.logger.Info("Getting template by ID: " + id)

	ctx, opSpan := tracing.StartGraphQLOperationSpan(ctx, "GetTemplate",
		attribute.String("graphql.template.id", id),
	)
	defer opSpan.End()

	// Сначала пробуем как UUID
	if templateUUID, err := uuid.Parse(id); err == nil {
		ctx, dbSpan := tracing.StartDBSpan(ctx, "GetTemplateByUUID")
		template, err := s.dbManagerRepo.GetTemplateByUUID(ctx, templateUUID)
		if err != nil {
			tracing.EndError(dbSpan, err)
			tracing.EndError(opSpan, err)
			dbSpan.End()
			return &model.TemplateResponse{
				Success: false,
				Message: strPtr(fmt.Sprintf("Template not found: %v", err)),
			}, nil
		}
		dbSpan.End()
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

	ctx, dbSpan := tracing.StartDBSpan(ctx, "GetTemplateByID")
	template, err := s.dbManagerRepo.GetTemplateByID(ctx, templateID)
	if err != nil {
		tracing.EndError(dbSpan, err)
		tracing.EndError(opSpan, err)
		dbSpan.End()
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr(fmt.Sprintf("Template not found: %v", err)),
		}, nil
	}
	dbSpan.End()

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
