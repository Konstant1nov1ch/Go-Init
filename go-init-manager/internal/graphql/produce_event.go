package graphql

import (
	"context"

	"go-init/internal/eventdata"

	"github.com/google/uuid"
	common "gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/logger"
)

// ProduceEvent публикует событие обработки шаблона в Kafka
func (s *Service) ProduceEvent(ctx context.Context, data *eventdata.ProcessTemplate) error {
	if !s.KafkaProducer.IsEnabled() {
		s.logger.InfoContext(ctx, "Kafka продюсер отключен, событие не будет отправлено")
		return nil
	}

	// Создаем событие для публикации
	event := common.ProduceEvent{
		Type:    eventdata.ProcessTemplateEventType,
		Schema:  eventdata.JsonSchema,
		Source:  s.serviceName,
		TopicID: eventdata.ProcessingTopicID,
	}

	// Генерируем уникальный ID корреляции
	correlationID := uuid.New().String()
	event.SetCorrelationID(correlationID)
	event.SetData(data)

	// Публикуем событие в Kafka
	s.logger.InfoContext(ctx, "Публикация события обработки шаблона",
		logger.String("template_id", data.ID),
		logger.String("correlation_id", correlationID),
		logger.String("topic", eventdata.ProcessingTopicID))

	s.KafkaProducer.Produce(ctx, &event)

	return nil
}
