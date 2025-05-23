package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"gitlab.com/go-init/go-init-common/default/logger"
)

// FileStorage provides methods for storing and retrieving files locally
type FileStorage struct {
	log       *logger.Logger
	baseDir   string
	createDir bool
}

// FileStorageOption is a function that configures a FileStorage
type FileStorageOption func(*FileStorage)

// WithCreateDir returns an option that configures whether the storage directory should be created
func WithCreateDir(create bool) FileStorageOption {
	return func(s *FileStorage) {
		s.createDir = create
	}
}

// NewFileStorage creates a new file storage service
func NewFileStorage(log *logger.Logger, baseDir string, opts ...FileStorageOption) (*FileStorage, error) {
	storage := &FileStorage{
		log:       log,
		baseDir:   baseDir,
		createDir: true, // By default, create the directory
	}

	// Apply options
	for _, opt := range opts {
		opt(storage)
	}

	// Create base directory if it doesn't exist
	if storage.createDir {
		if err := os.MkdirAll(baseDir, 0o755); err != nil {
			log.Error("Ошибка создания директории хранилища",
				logger.String("dir", baseDir),
				logger.Error(err))
			return nil, fmt.Errorf("failed to create storage directory: %w", err)
		}
	}

	log.Info("Хранилище файлов инициализировано",
		logger.String("base_dir", baseDir))
	return storage, nil
}

// SaveFile stores a file with the given content
func (s *FileStorage) SaveFile(ctx context.Context, fileName string, content []byte) (string, error) {
	// Create a directory based on the current date for organization
	dateDir := time.Now().Format("2006-01-02")
	dirPath := filepath.Join(s.baseDir, dateDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		s.log.ErrorContext(ctx, "Ошибка создания дневной директории",
			logger.String("dir", dirPath),
			logger.Error(err))
		return "", fmt.Errorf("failed to create date directory: %w", err)
	}

	// Full path to the file
	filePath := filepath.Join(dirPath, fileName)

	// Save the file
	err := os.WriteFile(filePath, content, 0o644)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка записи файла",
			logger.String("path", filePath),
			logger.Error(err))
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	s.log.InfoContext(ctx, "Файл успешно сохранен",
		logger.String("path", filePath))
	return filePath, nil
}

// SaveFileWriter returns a writer for storing a file
func (s *FileStorage) SaveFileWriter(ctx context.Context, fileName string) (io.WriteCloser, string, error) {
	// Create a directory based on the current date for organization
	dateDir := time.Now().Format("2006-01-02")
	dirPath := filepath.Join(s.baseDir, dateDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(dirPath, 0o755); err != nil {
		s.log.ErrorContext(ctx, "Ошибка создания дневной директории",
			logger.String("dir", dirPath),
			logger.Error(err))
		return nil, "", fmt.Errorf("failed to create date directory: %w", err)
	}

	// Full path to the file
	filePath := filepath.Join(dirPath, fileName)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка создания файла для записи",
			logger.String("path", filePath),
			logger.Error(err))
		return nil, "", fmt.Errorf("failed to create file: %w", err)
	}

	s.log.InfoContext(ctx, "Создан файл для записи",
		logger.String("path", filePath))
	return file, filePath, nil
}

// GetFile retrieves a file's content
func (s *FileStorage) GetFile(ctx context.Context, filePath string) ([]byte, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		s.log.ErrorContext(ctx, "Ошибка чтения файла",
			logger.String("path", filePath),
			logger.Error(err))
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	s.log.InfoContext(ctx, "Файл успешно прочитан",
		logger.String("path", filePath),
		logger.Int("size", len(content)))
	return content, nil
}

// GetFilePath returns the full path for a file
func (s *FileStorage) GetFilePath(fileName string, dateStr string) string {
	return filepath.Join(s.baseDir, dateStr, fileName)
}
