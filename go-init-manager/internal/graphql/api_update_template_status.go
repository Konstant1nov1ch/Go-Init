package graphql

import (
	"context"
	"fmt"

	"go-init/internal/graphql/converter"
	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
	"gitlab.com/go-init/go-init-common/default/logger"
)

// Template status constants
const (
	StatusPending    = "PENDING"
	StatusProcessing = "PROCESSING"
	StatusCompleted  = "COMPLETED"
	StatusFailed     = "FAILED"
)

// UpdateTemplateStatus updates the status of a template by UUID and returns a response
func (s *Service) UpdateTemplateStatus(ctx context.Context, templateUUID uuid.UUID, newStatus string) (*model.TemplateResponse, error) {
	// Validate status
	if !isValidStatus(newStatus) {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr(fmt.Sprintf("Invalid status: %s", newStatus)),
		}, nil
	}

	// Update status in the database
	err := s.dbManagerRepo.UpdateTemplateStatusByUUID(ctx, templateUUID, newStatus)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr(fmt.Sprintf("Failed to update template status: %v", err)),
		}, nil
	}

	// Retrieve the updated template
	template, err := s.dbManagerRepo.GetTemplateByUUID(ctx, templateUUID)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr(fmt.Sprintf("Template status updated but failed to retrieve template: %v", err)),
		}, nil
	}

	// Convert to GraphQL model
	graphqlTemplate := converter.DbTemplateToGraphqlTemplate(template)

	return &model.TemplateResponse{
		Success:  true,
		Message:  strPtr(fmt.Sprintf("Template status updated to %s", newStatus)),
		Template: graphqlTemplate,
	}, nil
}

// isValidStatus checks if the provided status is valid according to the TemplateStatus enum
func isValidStatus(status string) bool {
	switch status {
	case StatusPending, StatusProcessing, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

// Helper methods for common status updates

// MarkTemplateAsProcessing sets a template's status to PROCESSING
func (s *Service) MarkTemplateAsProcessing(ctx context.Context, templateUUID uuid.UUID) (*model.TemplateResponse, error) {
	return s.UpdateTemplateStatus(ctx, templateUUID, StatusProcessing)
}

// MarkTemplateAsCompleted sets a template's status to COMPLETED
func (s *Service) MarkTemplateAsCompleted(ctx context.Context, templateUUID uuid.UUID) (*model.TemplateResponse, error) {
	return s.UpdateTemplateStatus(ctx, templateUUID, StatusCompleted)
}

// MarkTemplateAsFailed sets a template's status to FAILED with an optional error message
func (s *Service) MarkTemplateAsFailed(ctx context.Context, templateUUID uuid.UUID, errorMsg string) (*model.TemplateResponse, error) {
	// First update the error message if provided
	if errorMsg != "" {
		if err := s.dbManagerRepo.UpdateTemplateErrorByUUID(ctx, templateUUID, errorMsg); err != nil {
			s.logger.WarnContext(ctx, "Failed to update template error message, continuing with status update",
				logger.String("template_uuid", templateUUID.String()),
				logger.Error(err))
		}
	}

	// Then update the status
	return s.UpdateTemplateStatus(ctx, templateUUID, StatusFailed)
}
