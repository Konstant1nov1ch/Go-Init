package graphql

import (
	"context"
	"go-init/internal/graphql/converter"
	"go-init/pkg/api/graphql/model"

	"github.com/google/uuid"
)

const PendingStatus = "pending"

func (s *Service) CreateTemplate(ctx context.Context, input model.CreateTemplateInput) (*model.TemplateResponse, error) {
	tx, err := s.agent.BeginTx(ctx)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr("Failed to begin transaction: " + err.Error()),
		}, nil
	}
	defer tx.Enfold(ctx, &err)

	template, err := converter.FromInputToDbServiceTemplate(ctx, input, s.logger)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr("Failed to convert input: " + err.Error()),
		}, nil
	}

	err = s.dbManagerRepo.CreateNewTemplate(ctx, template, tx)
	if err != nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr("Failed to create template: " + err.Error()),
		}, nil
	}

	var templateUUID uuid.UUID
	if template == nil || template.ServiceTemplateUuid == nil {
		templateUUID = uuid.New()
		s.logger.Warn("Template UUID is nil, using generated UUID")
	} else {
		templateUUID = *template.ServiceTemplateUuid
	}

	// Публикуем событие, используя UUID шаблона
	ev := converter.FromInputToEvent(input, templateUUID)
	s.ProduceEvent(ctx, &ev)

	graphqlTemplate := converter.DbTemplateToGraphqlTemplate(template)
	if graphqlTemplate == nil {
		return &model.TemplateResponse{
			Success: false,
			Message: strPtr("Failed to convert template to GraphQL model"),
		}, nil
	}

	return &model.TemplateResponse{
		Success:  true,
		Message:  strPtr("Template created successfully"),
		Template: graphqlTemplate,
	}, nil
}
