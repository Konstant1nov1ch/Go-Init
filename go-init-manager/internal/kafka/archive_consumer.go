package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	dbRepo "go-init/internal/database"
	"go-init/internal/eventdata"
	"go-init/internal/graphql"

	"github.com/google/uuid"
	"gitlab.com/go-init/go-init-common/default/logger"
)

// ArchiveConsumerService обрабатывает сообщения Kafka для событий archive-ready
type ArchiveConsumerService struct {
	log        *logger.Logger
	repository dbRepo.GoInitManagerRepository
}

// NewArchiveConsumerService создает новый сервис потребителя архивов
func NewArchiveConsumerService(log *logger.Logger, repository dbRepo.GoInitManagerRepository) *ArchiveConsumerService {
	return &ArchiveConsumerService{
		log:        log,
		repository: repository,
	}
}

// Work реализует интерфейс ConsumerWorker для обработки сообщений из Kafka
func (s *ArchiveConsumerService) Work(ctx context.Context, value []byte) error {
	s.log.InfoContext(ctx, "Processing Kafka message",
		logger.String("value", string(value)))

	// Парсим CloudEvent
	var cloudEvent eventdata.CloudEvent
	if err := json.Unmarshal(value, &cloudEvent); err != nil {
		s.log.ErrorContext(ctx, "Ошибка парсинга CloudEvent",
			logger.Error(err),
			logger.String("raw_message", string(value)))
		return fmt.Errorf("failed to parse CloudEvent: %w", err)
	}

	// Парсим поле data, которое содержит метаданные архива
	var metadata eventdata.ArchiveMetadata
	if err := json.Unmarshal(cloudEvent.Data, &metadata); err != nil {
		s.log.ErrorContext(ctx, "Ошибка парсинга метаданных архива",
			logger.Error(err),
			logger.String("data_field", string(cloudEvent.Data)))
		return fmt.Errorf("failed to parse archive metadata: %w", err)
	}

	s.log.InfoContext(ctx, "Обработка метаданных архива",
		logger.String("object_name", metadata.ObjectName),
		logger.String("archive_id", metadata.ID))

	// Извлекаем UUID запроса из ID или имени объекта
	var requestUUID uuid.UUID
	var err error

	// Сначала пробуем использовать ID из метаданных
	if metadata.ID != "" {
		requestUUID, err = uuid.Parse(metadata.ID)
		if err != nil {
			s.log.WarnContext(ctx, "Ошибка парсинга UUID из metadata.ID, пробуем ObjectName",
				logger.Error(err),
				logger.String("id", metadata.ID))
			// Продолжаем и попробуем использовать ObjectName
		} else {
			s.log.InfoContext(ctx, "Успешно использован UUID из metadata.ID",
				logger.String("uuid", requestUUID.String()))
		}
	}

	// Если ID не работает, используем ObjectName
	if requestUUID == uuid.Nil && metadata.ObjectName != "" {
		// Удаляем расширение .zip, если оно есть
		objectID := strings.TrimSuffix(metadata.ObjectName, ".zip")
		requestUUID, err = uuid.Parse(objectID)
		if err != nil {
			s.log.ErrorContext(ctx, "Недействительный UUID в имени объекта",
				logger.String("object_name", metadata.ObjectName),
				logger.Error(err))
			return fmt.Errorf("invalid UUID in object name: %w", err)
		}
		s.log.InfoContext(ctx, "Успешно использован UUID из ObjectName",
			logger.String("uuid", requestUUID.String()))
	}

	// Проверяем, что у нас есть действительный UUID
	if requestUUID == uuid.Nil {
		s.log.ErrorContext(ctx, "Не удалось получить UUID запроса ни из ID, ни из ObjectName")
		return fmt.Errorf("failed to extract request UUID from message")
	}

	// Обновление статуса на COMPLETED
	if err := s.repository.UpdateTemplateStatusByUUID(ctx, requestUUID, graphql.StatusCompleted); err != nil {
		s.log.ErrorContext(ctx, "Не удалось обновить статус шаблона на COMPLETED",
			logger.Error(err),
			logger.String("template_uuid", requestUUID.String()))
		return err
	}
	s.log.InfoContext(ctx, "Статус шаблона обновлен на COMPLETED",
		logger.String("template_uuid", requestUUID.String()))

	// Обновление URL архива
	if metadata.PresignedURL != "" {
		if err := s.repository.UpdateZipUrl(ctx, requestUUID, metadata.PresignedURL); err != nil {
			s.log.ErrorContext(ctx, "Не удалось обновить URL архива шаблона",
				logger.Error(err),
				logger.String("template_uuid", requestUUID.String()),
				logger.String("presigned_url", metadata.PresignedURL))
			return err
		}
		s.log.InfoContext(ctx, "URL архива шаблона обновлен",
			logger.String("template_uuid", requestUUID.String()),
			logger.String("presigned_url", metadata.PresignedURL))
	} else {
		s.log.WarnContext(ctx, "Отсутствует presignedURL в метаданных архива",
			logger.String("template_uuid", requestUUID.String()))
	}

	return nil
}
