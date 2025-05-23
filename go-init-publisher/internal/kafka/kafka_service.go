package kafka

import (
	"context"
	"encoding/json"

	"gitlab.com/go-init/go-init-common/default/logger"
)

// ArchiveMessage сообщение с данными архива
type ArchiveMessage struct {
	ArchiveID string `json:"archive_id"`
	Data      []byte `json:"data"`
}

// KafkaService сервис для обработки сообщений из Kafka
type KafkaService struct {
	log *logger.Logger
}

// NewKafkaService создает новый экземпляр сервиса
func NewKafkaService(log *logger.Logger) *KafkaService {
	return &KafkaService{
		log: log,
	}
}

// ProcessArchiveMessage обрабатывает сообщение с архивом
func (s *KafkaService) ProcessArchiveMessage(ctx context.Context, value []byte) error {
	s.log.Info("Получено сообщение с архивом")

	var archiveMsg ArchiveMessage
	if err := json.Unmarshal(value, &archiveMsg); err != nil {
		s.log.Error("Ошибка при декодировании сообщения: " + err.Error())
		return err
	}

	s.log.Info("Сохранение архива с ID: " + archiveMsg.ArchiveID)

	s.log.Info("Архив успешно сохранен")
	return nil
}

// Work метод для реализации интерфейса kafka.ConsumerWorker
func (s *KafkaService) Work(ctx context.Context, value []byte) error {
	return s.ProcessArchiveMessage(ctx, value)
}
