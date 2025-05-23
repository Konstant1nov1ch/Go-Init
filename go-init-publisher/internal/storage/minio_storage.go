package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	common "gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/logger"
	minio_client "gitlab.com/go-init/go-init-common/default/s3/minio"
)

const (
	// ContentTypeZip - MIME тип для ZIP-архивов
	ContentTypeZip = "application/zip"

	// FileExtensionZip - расширение файла ZIP-архива
	FileExtensionZip = ".zip"

	// EventTypeArchiveReady - тип события готовности архива
	EventTypeArchiveReady = "archive-ready"

	// EventSchemaArchive - схема события архива
	EventSchemaArchive = "go-init-archive-schema"

	// EventSourcePublisher - источник события
	EventSourcePublisher = "go-init-publisher"

	// PresignedURLExpiry - срок действия пресигнированной ссылки (24 часа)
	PresignedURLExpiry = 24 * time.Hour
)

// MinIOStorage реализует операции хранения с использованием MinIO
type MinIOStorage struct {
	minioClient   *minio_client.Client
	kafkaProducer *common.ClientConfig
	log           *logger.Logger
	config        *minio_client.Config
	topic         string
}

// NewMinIOStorage создает новый экземпляр хранилища MinIO
func NewMinIOStorage(minioClient *minio_client.Client, kafkaProducer *common.ClientConfig, log *logger.Logger, config *minio_client.Config, topic string) *MinIOStorage {
	return &MinIOStorage{
		minioClient:   minioClient,
		kafkaProducer: kafkaProducer,
		log:           log,
		config:        config,
		topic:         topic,
	}
}

// SaveArchive сохраняет архив в MinIO и публикует метаданные в Kafka
func (s *MinIOStorage) SaveArchive(ctx context.Context, archiveID string, data []byte) (string, error) {
	// Проверка идентификатора архива
	if archiveID == "" {
		archiveID = uuid.New().String()
		s.log.WarnContext(ctx, "Предоставлен пустой ID архива, сгенерирован новый",
			logger.String("archive_id", archiveID))
	}

	// Формирование имени объекта
	objectName := fmt.Sprintf("%s%s", archiveID, FileExtensionZip)

	// Загрузка архива в MinIO
	s.log.InfoContext(ctx, "Загрузка архива в MinIO",
		logger.String("object_name", objectName),
		logger.String("bucket", s.config.DefaultBucket))

	info, err := s.minioClient.UploadObject(ctx, s.config.DefaultBucket, objectName, data, ContentTypeZip)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка загрузки архива в MinIO",
			logger.String("object_name", objectName),
			logger.Error(err))
		return "", fmt.Errorf("ошибка загрузки архива в MinIO: %w", err)
	}

	// Создание метаданных
	archiveMetadata := s.createArchiveMetadata(ctx, info, archiveID, objectName)

	// Публикация метаданных в Kafka
	s.publishMetadataToKafka(ctx, archiveMetadata, objectName)

	return objectName, nil
}

// SaveArchiveFromReader сохраняет архив из потока в MinIO и публикует метаданные в Kafka
func (s *MinIOStorage) SaveArchiveFromReader(ctx context.Context, reader io.Reader, size int64) (string, error) {
	// Генерация уникального ID для архива
	archiveID := uuid.New().String()
	objectName := fmt.Sprintf("%s%s", archiveID, FileExtensionZip)

	// Загрузка архива в MinIO из потока
	s.log.InfoContext(ctx, "Загрузка архива в MinIO из потока",
		logger.String("object_name", objectName),
		logger.String("bucket", s.config.DefaultBucket),
		logger.Int64("size", size))

	info, err := s.minioClient.UploadObjectFromReader(ctx, s.config.DefaultBucket, objectName, reader, size, ContentTypeZip)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка загрузки архива в MinIO из потока",
			logger.String("object_name", objectName),
			logger.Error(err))
		return "", fmt.Errorf("ошибка загрузки архива в MinIO из потока: %w", err)
	}

	// Создание метаданных
	archiveMetadata := s.createArchiveMetadata(ctx, info, archiveID, objectName)

	// Публикация метаданных в Kafka
	s.publishMetadataToKafka(ctx, archiveMetadata, objectName)

	return objectName, nil
}

// createArchiveMetadata создает и настраивает метаданные архива
func (s *MinIOStorage) createArchiveMetadata(ctx context.Context, info minio.UploadInfo, archiveID string, objectName string) *minio_client.ArchiveMetadata {
	// Создание метаданных объекта
	objectMetadata := minio_client.NewObjectMetadata(
		info.Bucket,
		objectName,
		info.Size,
		ContentTypeZip,
		info.ETag,
	)

	// Создание метаданных архива
	archiveMetadata := minio_client.NewArchiveMetadata(objectMetadata, "zip")
	archiveMetadata.ID = archiveID

	// Генерация пресигнированной ссылки
	presignedURL, err := s.GeneratePresignedURL(ctx, objectName, PresignedURLExpiry)
	if err != nil {
		s.log.WarnContext(ctx, "Не удалось сгенерировать пресигнированную ссылку",
			logger.String("object_name", objectName),
			logger.Error(err))
	} else {
		archiveMetadata.SetPresignedURL(presignedURL, PresignedURLExpiry)
	}

	return &archiveMetadata
}

// publishMetadataToKafka публикует метаданные архива в Kafka
func (s *MinIOStorage) publishMetadataToKafka(ctx context.Context, metadata *minio_client.ArchiveMetadata, objectName string) {
	// Публикация метаданных в Kafka, если продюсер включен
	if s.kafkaProducer.IsEnabled() {
		s.log.InfoContext(ctx, "Публикация метаданных архива в Kafka",
			logger.String("topic", s.topic),
			logger.String("archive_id", metadata.ID))

		event := &common.ProduceEvent{
			Type:    EventTypeArchiveReady,
			Schema:  EventSchemaArchive,
			Source:  EventSourcePublisher,
			TopicID: s.topic,
		}

		event.SetCorrelationID(uuid.New().String())
		event.SetData(metadata)

		s.kafkaProducer.Produce(ctx, event)
		s.log.InfoContext(ctx, "Метаданные архива опубликованы в Kafka",
			logger.String("object_name", objectName))
	} else {
		s.log.InfoContext(ctx, "Kafka продюсер отключен, пропуск публикации метаданных")
	}
}

// DownloadArchive загружает архив из MinIO
func (s *MinIOStorage) DownloadArchive(ctx context.Context, objectName string) ([]byte, error) {
	s.log.InfoContext(ctx, "Загрузка архива из MinIO",
		logger.String("object_name", objectName),
		logger.String("bucket", s.config.DefaultBucket))

	data, err := s.minioClient.DownloadObject(ctx, s.config.DefaultBucket, objectName)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка загрузки архива из MinIO",
			logger.String("object_name", objectName),
			logger.Error(err))
		return nil, fmt.Errorf("ошибка загрузки архива из MinIO: %w", err)
	}

	return data, nil
}

// DeleteArchive удаляет архив из MinIO
func (s *MinIOStorage) DeleteArchive(ctx context.Context, objectName string) error {
	s.log.InfoContext(ctx, "Удаление архива из MinIO",
		logger.String("object_name", objectName),
		logger.String("bucket", s.config.DefaultBucket))

	err := s.minioClient.DeleteObject(ctx, s.config.DefaultBucket, objectName)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка удаления архива из MinIO",
			logger.String("object_name", objectName),
			logger.Error(err))
		return fmt.Errorf("ошибка удаления архива из MinIO: %w", err)
	}

	return nil
}

func (s *MinIOStorage) GeneratePresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	s.log.InfoContext(ctx, "Генерация пресигнированной ссылки для архива",
		logger.String("object_name", objectName),
		logger.String("bucket", s.config.DefaultBucket),
		logger.Duration("expiry", expiry))

	origURL, err := s.minioClient.GeneratePresignedURL(ctx, s.config.DefaultBucket, objectName, expiry)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка генерации пресигнированной ссылки",
			logger.String("object_name", objectName),
			logger.Error(err))
		return "", fmt.Errorf("ошибка генерации пресигнированной ссылки: %w", err)
	}

	if s.config.PublicAccessURL != "" {
		publicBase, err := url.Parse(s.config.PublicAccessURL)
		if err == nil {
			parsed := *origURL
			parsed.Scheme = publicBase.Scheme
			parsed.Host = publicBase.Host
			parsed.Path = publicBase.Path + parsed.Path // 💥 именно так

			s.log.InfoContext(ctx, "Генерация пресигнированной ссылки для архива",
				logger.String("link", parsed.String()))

			return parsed.String(), nil
		}
	}

	return origURL.String(), nil
}
