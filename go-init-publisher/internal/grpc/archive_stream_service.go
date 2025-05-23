package grpc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go-init-publisher/internal/storage"
	"io"
	"os"
	"strconv"
	"time"

	pb "go-init-publisher/pkg/api/grpc"

	"github.com/google/uuid"
	"gitlab.com/go-init/go-init-common/default/logger"
)

const (
	// OptimalChunkSize размер чанка для стриминга архива (примерно 16 000 байт)
	OptimalChunkSize = 16 * 1024
	// DefaultDebugDir директория для сохранения отладочных архивов
	DefaultDebugDir = "debug_archives"
	// FeatureToggleSaveLocallyEnv переменная окружения для включения локального сохранения
	FeatureToggleSaveLocallyEnv = "PUBLISHER_SAVE_ARCHIVE_LOCALLY"
	// DefaultArchiveExtension расширение файла архива по умолчанию
	DefaultArchiveExtension = ".zip"
)

// ArchiveStreamService provides an implementation of the streaming service
type ArchiveStreamService struct {
	pb.UnimplementedArchivePublisherServer
	log          *logger.Logger
	fileStorage  *storage.FileStorage
	minioStorage *storage.MinIOStorage
}

// NewArchiveStreamService creates a new archive streaming service
func NewArchiveStreamService(
	log *logger.Logger,
	fileStorage *storage.FileStorage,
	minioStorage *storage.MinIOStorage,
) *ArchiveStreamService {
	return &ArchiveStreamService{
		log:          log,
		fileStorage:  fileStorage,
		minioStorage: minioStorage,
	}
}

// StreamArchive handles streaming archive chunks from the client
func (s *ArchiveStreamService) StreamArchive(stream pb.ArchivePublisher_StreamArchiveServer) error {
	ctx := stream.Context()
	s.log.InfoContext(ctx, "Получен новый запрос на стриминг архива")

	// Проверка включена ли функция локального сохранения
	saveLocally := s.isLocalSavingEnabled(ctx)

	// Подготовка к приему архива
	var chunks [][]byte
	var totalBytes int64
	startTime := time.Now()
	var archiveID string
	var expectedHash string

	// Получаем все чанки
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// Конец потока
			break
		}
		if err != nil {
			s.log.ErrorContext(ctx, "Ошибка при получении чанка",
				logger.Error(err))
			return fmt.Errorf("error receiving chunk: %w", err)
		}

		// Сохраняем ID архива из первого чанка
		if archiveID == "" && chunk.ArchiveId != "" {
			archiveID = chunk.ArchiveId
			s.log.InfoContext(ctx, "Получен ID архива из запроса",
				logger.String("archive_id", archiveID))
		}

		// Сохраняем ожидаемый хеш, если он передан
		if expectedHash == "" && chunk.ExpectedHash != "" {
			expectedHash = chunk.ExpectedHash
			s.log.InfoContext(ctx, "Получен ожидаемый хеш архива",
				logger.String("expected_hash", expectedHash))
		}

		chunks = append(chunks, chunk.Data)
		totalBytes += int64(len(chunk.Data))

		// Логирование прогресса для больших файлов
		if totalBytes%(1024*1024) < OptimalChunkSize {
			s.log.InfoContext(ctx, "Прогресс приема архива",
				logger.Int64("received_mb", totalBytes/(1024*1024)))
		}
	}

	// Объединяем все чанки в один массив байтов
	archiveData := make([]byte, 0, totalBytes)
	for _, chunk := range chunks {
		archiveData = append(archiveData, chunk...)
	}

	// Если ID архива не был получен из запроса, генерируем новый
	if archiveID == "" {
		archiveID = uuid.New().String()
		s.log.WarnContext(ctx, "ID архива не был получен из запроса, сгенерирован новый",
			logger.String("archive_id", archiveID))
	}

	// Проверка целостности данных с помощью SHA-256, если был передан ожидаемый хеш
	if expectedHash != "" {
		actualHash := sha256.Sum256(archiveData)
		actualHashHex := hex.EncodeToString(actualHash[:])

		if actualHashHex != expectedHash {
			errMsg := "Ошибка проверки целостности: хеш полученного архива не совпадает с ожидаемым"
			s.log.ErrorContext(ctx, errMsg,
				logger.String("archive_id", archiveID),
				logger.String("expected_hash", expectedHash),
				logger.String("actual_hash", actualHashHex))

			return stream.SendMsg(&pb.StreamResponse{
				Success:     false,
				Message:     errMsg,
				ArchivePath: "",
			})
		}

		s.log.InfoContext(ctx, "Проверка целостности данных успешно пройдена",
			logger.String("archive_id", archiveID),
			logger.String("hash", actualHashHex))
	} else {
		s.log.WarnContext(ctx, "Ожидаемый хеш не был передан, проверка целостности не выполнена",
			logger.String("archive_id", archiveID))
	}

	// Загрузка в MinIO и публикация метаданных в Kafka
	s.log.InfoContext(ctx, "Загрузка архива в MinIO",
		logger.String("archive_id", archiveID),
		logger.Int64("size", totalBytes))

	objectName, err := s.minioStorage.SaveArchive(ctx, archiveID, archiveData)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка при загрузке архива в MinIO",
			logger.String("archive_id", archiveID),
			logger.Error(err))
		return fmt.Errorf("failed to upload archive to MinIO: %w", err)
	}

	// Также сохраняем локально, если функция включена
	if saveLocally {
		s.saveArchiveLocally(ctx, archiveID, archiveData)
	}

	// Отправляем ответ
	duration := time.Since(startTime)
	s.log.InfoContext(ctx, "Стриминг архива завершен",
		logger.String("archive_id", archiveID),
		logger.String("object_name", objectName),
		logger.Int64("size", totalBytes),
		logger.Duration("duration", duration))

	return stream.SendMsg(&pb.StreamResponse{
		Success:     true,
		Message:     fmt.Sprintf("Архив успешно сохранен (%d байт) и отправлен в MinIO", totalBytes),
		ArchivePath: objectName,
	})
}

// isLocalSavingEnabled проверяет, включена ли функция локального сохранения
func (s *ArchiveStreamService) isLocalSavingEnabled(ctx context.Context) bool {
	saveLocallyStr := os.Getenv(FeatureToggleSaveLocallyEnv)
	saveLocally, _ := strconv.ParseBool(saveLocallyStr) // Ignore error, defaults to false

	if saveLocally {
		s.log.InfoContext(ctx, "Локальное сохранение архива включено")
	} else {
		s.log.InfoContext(ctx, "Локальное сохранение архива отключено")
	}

	return saveLocally
}

// saveArchiveLocally сохраняет архив локально
func (s *ArchiveStreamService) saveArchiveLocally(ctx context.Context, archiveID string, data []byte) {
	// Генерируем имя файла
	fileName := fmt.Sprintf("%s%s", archiveID, DefaultArchiveExtension)

	// Сохраняем архив с использованием сервиса файлового хранилища
	filePath, err := s.fileStorage.SaveFile(ctx, fileName, data)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка при локальном сохранении архива",
			logger.String("archive_id", archiveID),
			logger.Error(err))
	} else {
		s.log.InfoContext(ctx, "Архив также сохранен локально",
			logger.String("archive_id", archiveID),
			logger.String("path", filePath))
	}
}
