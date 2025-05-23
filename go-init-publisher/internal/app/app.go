package app

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"go-init-publisher/config"
	"go-init-publisher/internal/grpc"
	"go-init-publisher/internal/storage"
	pb "go-init-publisher/pkg/api/grpc"

	"gitlab.com/go-init/go-init-common/default/kafka"
	"gitlab.com/go-init/go-init-common/default/s3/minio"

	"gitlab.com/go-init/go-init-common/default/closer"
	"gitlab.com/go-init/go-init-common/default/logger"
)

type App struct {
	cfg           *config.AppConfig
	log           *logger.Logger
	KafkaProducer *kafka.ClientConfig
	grpcServer    *pb.Server
	fileStorage   *storage.FileStorage
	minioClient   *minio.Client
	minioStorage  *storage.MinIOStorage
}

const (
	serviceName        = "go-init-publisher"
	shutDownTimeOut    = time.Second * 5
	defaultStoragePath = "./storage" // Путь по умолчанию для хранения файлов
)

func New(ctx context.Context) (*App, error) {
	a := &App{}
	err := a.initDeps(ctx)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (a *App) initDeps(ctx context.Context) error {
	inits := []func(context.Context) error{
		a.initConfig,
		a.initLogger,
		a.initCloser,
		a.initKafka,
		a.initFileStorage,
		a.initMinioClient,
		a.initMinioStorage,
		a.initGrpcServer,
		a.initServices,
	}
	for _, f := range inits {
		err := f(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *App) initConfig(_ context.Context) error {
	a.cfg = config.GetConfig()
	return nil
}

func (a *App) initKafka(_ context.Context) error {
	k, err := kafka.NewClientConfig(&a.cfg.Kafka, a.log)
	if err != nil {
		return fmt.Errorf("failed to initialize kafka client: %w", err)
	}
	a.KafkaProducer = k
	return nil
}

func (a *App) initLogger(_ context.Context) error {
	a.log = logger.New(&a.cfg.Logger, a.cfg.HttpServ.Name, a.cfg.HttpServ.Env)
	return nil
}

func (a *App) initCloser(_ context.Context) error {
	closer.InitCloser(a.log)
	return nil
}

func (a *App) initFileStorage(ctx context.Context) error {
	if a.log == nil || a.cfg == nil {
		return fmt.Errorf("logger or config not initialized")
	}

	// Используем путь из переменной окружения или путь по умолчанию
	storagePath := defaultStoragePath
	if os.Getenv("STORAGE_PATH") != "" {
		storagePath = os.Getenv("STORAGE_PATH")
	}

	// Создаем хранилище файлов с настроенным путем
	fileStorage, err := storage.NewFileStorage(a.log, storagePath)
	if err != nil {
		a.log.ErrorContext(ctx, "Ошибка инициализации хранилища файлов",
			logger.String("path", storagePath),
			logger.Error(err))
		return fmt.Errorf("failed to initialize file storage: %w", err)
	}

	a.fileStorage = fileStorage
	a.log.InfoContext(ctx, "Хранилище файлов инициализировано",
		logger.String("path", storagePath))
	return nil
}

func (a *App) initMinioClient(ctx context.Context) error {
	if a.log == nil || a.cfg == nil {
		return fmt.Errorf("logger or config not initialized")
	}

	// Создаем клиент MinIO
	minioClient, err := minio.New(&a.cfg.MinIO, a.log)
	if err != nil {
		a.log.ErrorContext(ctx, "Ошибка инициализации клиента MinIO",
			logger.String("endpoint", a.cfg.MinIO.Endpoint),
			logger.Error(err))
		return fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	// Проверяем, что бакет существует
	err = minioClient.EnsureBucketExists(ctx, a.cfg.MinIO.DefaultBucket)
	if err != nil {
		a.log.ErrorContext(ctx, "Ошибка проверки существования бакета MinIO",
			logger.String("bucket", a.cfg.MinIO.DefaultBucket),
			logger.Error(err))
		return fmt.Errorf("failed to ensure MinIO bucket exists: %w", err)
	}

	a.minioClient = minioClient
	a.log.InfoContext(ctx, "Клиент MinIO инициализирован",
		logger.String("endpoint", a.cfg.MinIO.Endpoint),
		logger.String("bucket", a.cfg.MinIO.DefaultBucket))
	return nil
}

func (a *App) initMinioStorage(ctx context.Context) error {
	if a.log == nil || a.cfg == nil || a.minioClient == nil || a.KafkaProducer == nil {
		return fmt.Errorf("dependencies for MinIO storage not initialized")
	}

	// Определение топика для публикации событий
	topic := a.getKafkaTopic()

	// Создание сервиса хранения MinIO с использованием существующего Kafka продюсера
	minioStorage := storage.NewMinIOStorage(a.minioClient, a.KafkaProducer, a.log, &a.cfg.MinIO, topic)
	a.minioStorage = minioStorage

	a.log.InfoContext(ctx, "MinIO storage service initialized",
		logger.String("topic", topic))
	return nil
}

// getKafkaTopic возвращает ID топика Kafka для публикации событий о готовности архива
func (a *App) getKafkaTopic() string {
	const defaultTopic = "go-init-done"
	const archiveDoneTopicID = "go-init-done"

	// Сначала пытаемся получить топик из конфигурации Kafka продюсера
	if a.cfg.Kafka.ProducerConfig.Enabled && len(a.cfg.Kafka.ProducerConfig.Topic) > 0 {
		for _, t := range a.cfg.Kafka.ProducerConfig.Topic {
			if t.Id == archiveDoneTopicID && t.IsEnabled {
				a.log.Info("Using Kafka topic from config",
					logger.String("topic", t.Id))
				return t.Id
			}
		}
	}

	// Если не нашли в конфигурации, пробуем переменную окружения
	envTopic := os.Getenv("KAFKA_TOPIC_ARCHIVE_DONE")
	if envTopic != "" {
		a.log.Info("Using Kafka topic from environment",
			logger.String("topic", envTopic))
		return envTopic
	}

	// Используем значение по умолчанию
	a.log.Info("Using default Kafka topic",
		logger.String("topic", defaultTopic))
	return defaultTopic
}

func (a *App) initGrpcServer(ctx context.Context) error {
	if a.log == nil || a.cfg == nil {
		return fmt.Errorf("logger or config not initialized")
	}

	// Создаем конфигурацию для нашего gRPC сервера
	config := pb.ServerConfig{
		Port: a.cfg.GrpcServ.Port,
	}

	// Создаем экземпляр нашего gRPC сервера
	grpcServer, err := pb.NewServer(config)
	if err != nil {
		a.log.ErrorContext(ctx, "Ошибка инициализации gRPC сервера",
			logger.String("port", a.cfg.GrpcServ.Port),
			logger.Error(err))
		return fmt.Errorf("failed to initialize gRPC server: %w", err)
	}

	a.grpcServer = grpcServer
	a.log.InfoContext(ctx, "gRPC сервер инициализирован",
		logger.String("port", a.cfg.GrpcServ.Port))
	return nil
}

func (a *App) runGrpcServer() error {
	a.log.Info("Запуск gRPC сервера",
		logger.String("port", a.cfg.GrpcServ.Port))

	if err := a.grpcServer.Start(); err != nil {
		a.log.Error("Ошибка gRPC сервера",
			logger.Error(err))
		return err
	}
	return nil
}

func (a *App) initServices(ctx context.Context) error {
	if a.log == nil || a.cfg == nil || a.minioClient == nil || a.KafkaProducer == nil || a.fileStorage == nil || a.minioStorage == nil {
		a.log.ErrorContext(ctx, "Ошибка инициализации сервисов: зависимости не инициализированы")
		return fmt.Errorf("dependencies not initialized")
	}

	// Создаем сервис для стриминга архивов с поддержкой MinIO
	archiveStreamService := grpc.NewArchiveStreamService(a.log, a.fileStorage, a.minioStorage)

	// Регистрируем сервис в gRPC сервере
	pb.RegisterArchivePublisherServer(a.grpcServer.GetGRPCServer(), archiveStreamService)
	a.log.InfoContext(ctx, "Зарегистрирован сервис для стриминга архивов")

	// Запускаем Kafka продюсер если он включен
	if a.KafkaProducer.ProducerIsEnabled() {
		go a.KafkaProducer.Start(ctx)
		a.log.InfoContext(ctx, "Kafka продюсер запущен")
	} else {
		a.log.InfoContext(ctx, "Kafka продюсер отключен")
	}

	return nil
}

func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		err := a.runGrpcServer()
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to start grpc server: %v", err))
		}
	}()

	wg.Wait()
	return nil
}
