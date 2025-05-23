package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"go-init-gen/config"
	"go-init-gen/internal/api/grpc"
	"go-init-gen/internal/work"

	database "gitlab.com/go-init/go-init-common/default/db/pg/orm"
	"gitlab.com/go-init/go-init-common/default/grpcpkg"
	"gitlab.com/go-init/go-init-common/default/kafka"

	"gitlab.com/go-init/go-init-common/default/closer"
	"gitlab.com/go-init/go-init-common/default/logger"
)

type App struct {
	cfg             *config.AppConfig
	log             *logger.Logger
	db              *database.AgentImpl
	KafkaConsumer   *kafka.ClientConfig
	worker          *work.Worker
	publisherClient *grpc.PublisherClient
	cancelFunc      context.CancelFunc
}

const (
	serviceName     = "go-init-generator"
	shutDownTimeOut = time.Second * 5
)

func New(ctx context.Context) (*App, error) {
	ctx, cancel := context.WithCancel(ctx)

	a := &App{
		cancelFunc: cancel,
	}
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
		a.initDb,
		a.initKafka,
		a.initPublisherClient,
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
	if k.ConsumerIsEnabled() {
		a.KafkaConsumer = k
	}

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

func (a *App) initPublisherClient(_ context.Context) error {
	if a.log == nil || a.cfg == nil {
		return fmt.Errorf("logger or config not initialized")
	}

	// Get address from environment variable or fallback to config
	address := os.Getenv("GRPC_CLIENT_ADDRESS")
	if address == "" {
		address = a.cfg.GrpcClients.Address
	}

	config := grpcpkg.ClientConfig{
		Address: address,
		UseTLS:  a.cfg.GrpcClients.UseTLS, // Use app config for TLS
	}

	a.log.Info(fmt.Sprintf("Connecting to publisher service at: %s", address))

	client, err := grpc.NewPublisherClient(config, a.log)
	if err != nil {
		return fmt.Errorf("failed to initialize publisher client: %w", err)
	}

	a.publisherClient = client

	closer.Add(func() error {
		a.log.Info("Closing publisher client connection...")
		return a.publisherClient.Close()
	})

	a.log.Info("Publisher client initialized successfully")
	return nil
}

func (a *App) initServices(ctx context.Context) error {
	if a.log == nil || a.cfg == nil {
		return fmt.Errorf("logger or config not initialized")
	}

	// Create the worker with publisher client
	a.worker = work.NewWorker(ctx, a.log, a.KafkaConsumer, a.publisherClient)

	// Then register it with Kafka if consumer is enabled
	if a.KafkaConsumer != nil && a.KafkaConsumer.ConsumerIsEnabled() {
		if len(a.cfg.Kafka.ConsumerConfig.Topic) > 0 {
			topicID := a.cfg.Kafka.ConsumerConfig.Topic[0].Id
			err := a.KafkaConsumer.RegisterConsumerWorkersByTopic(topicID, a.worker)
			if err != nil {
				return fmt.Errorf("failed to register kafka consumer worker: %w", err)
			}
			a.log.Info(fmt.Sprintf("Registered worker for topic ID: %s", topicID))
		} else {
			return fmt.Errorf("no topics available for registration")
		}
	}

	closer.Add(func() error {
		return a.worker.Stop()
	})

	return nil
}

func (a *App) initDb(ctx context.Context) error {
	return nil
}

func (a *App) Run() error {
	defer func() {
		a.log.Info("Shutting down application...")
		a.cancelFunc()
		closer.CloseAll()
		closer.Wait()
		a.log.Info("Application shutdown complete")
	}()

	errChan := make(chan error, 2)

	// Start the Kafka consumer if it's enabled
	if a.KafkaConsumer != nil && a.KafkaConsumer.ConsumerIsEnabled() {
		go func() {
			a.log.Info("Starting Kafka consumer...")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel() // Ensure the cancel function is called to prevent context leak
			a.KafkaConsumer.Start(ctx)
		}()
	}

	go func() {
		a.log.Info("Starting worker...")
		if err := a.worker.Start(); err != nil {
			a.log.Error(fmt.Sprintf("Worker error: %v", err))
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	}
}
