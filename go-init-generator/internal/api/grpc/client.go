package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	externalgrpc "go-init-gen/pkg/api/grpc/external"

	"github.com/google/uuid"
	"gitlab.com/go-init/go-init-common/default/grpcpkg"
	"gitlab.com/go-init/go-init-common/default/logger"
)

const (
	// OptimalChunkSize размер чанка для стриминга архива (примерно 16 KB)
	OptimalChunkSize = 16 * 1024
)

// PublisherClient клиент для работы с publisher сервисом
type PublisherClient struct {
	client           *grpcpkg.GRPCClient
	archivePublisher externalgrpc.ArchivePublisherClient
	logger           *logger.Logger
}

// StreamArchive отправляет архив в publisher сервис через стриминг
func (c *PublisherClient) StreamArchive(ctx context.Context, archiveID string, data []byte) error {
	if archiveID == "" {
		// Генерируем уникальный ID для архива только если не был предоставлен
		archiveID = uuid.New().String()
		c.logger.Info(fmt.Sprintf("Generated new archive ID: %s", archiveID))
	} else {
		c.logger.Info(fmt.Sprintf("Using provided archive ID: %s", archiveID))
	}

	// Вычисляем хеш для проверки целостности
	hash := sha256.Sum256(data)
	hashString := hex.EncodeToString(hash[:])
	c.logger.Info(fmt.Sprintf("Calculated hash for archive %s: %s", archiveID, hashString))

	// Создаем клиентский стрим
	stream, err := c.archivePublisher.StreamArchive(ctx)
	if err != nil {
		return fmt.Errorf("ошибка при создании стрима: %w", err)
	}

	// Разбиваем архив на чанки и отправляем каждый чанк
	totalBytes := len(data)
	sentBytes := 0

	for i := 0; i < totalBytes; i += OptimalChunkSize {
		// Проверяем контекст на отмену
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Определяем размер чанка
			chunkSize := OptimalChunkSize
			if i+chunkSize > totalBytes {
				chunkSize = totalBytes - i
			}

			// Создаем чанк с данными
			chunk := &externalgrpc.ArchiveChunk{
				ArchiveId:    archiveID,
				Data:         data[i : i+chunkSize],
				IsLast:       (i + chunkSize) >= totalBytes,
				ExpectedHash: hashString,
			}

			// Отправляем чанк
			if err := stream.Send(chunk); err != nil {
				return fmt.Errorf("ошибка при отправке чанка: %w", err)
			}

			sentBytes += chunkSize
		}
	}

	// Закрываем стрим и получаем ответ
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return fmt.Errorf("ошибка при закрытии стрима: %w", err)
	}

	if !resp.Success {
		return fmt.Errorf("ошибка от сервера: %s", resp.Message)
	}

	return nil
}

// Close closes the gRPC client connection
func (c *PublisherClient) Close() error {
	return c.client.GetConnection().Close()
}

// NewPublisherClient initializes a new PublisherClient
func NewPublisherClient(config grpcpkg.ClientConfig, logger *logger.Logger) (*PublisherClient, error) {
	client, err := grpcpkg.NewGRPCClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC client: %w", err)
	}

	archivePublisher := externalgrpc.NewArchivePublisherClient(client.GetConnection())

	return &PublisherClient{
		client:           client,
		archivePublisher: archivePublisher,
		logger:           logger,
	}, nil
}
