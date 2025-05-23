package app

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go-init/config"
	"go-init/internal/database/request_repo/models"
	"go-init/internal/graphql"
	"go-init/internal/kafka"
	generatedGQL "go-init/pkg/api/graphql"

	database "gitlab.com/go-init/go-init-common/default/db/pg/orm"
	commonKafka "gitlab.com/go-init/go-init-common/default/kafka"

	dbRepo "go-init/internal/database"

	"gitlab.com/go-init/go-init-common/default/closer"
	grpcpkg "gitlab.com/go-init/go-init-common/default/grpcpkg"
	myhttp "gitlab.com/go-init/go-init-common/default/http"
	myserver "gitlab.com/go-init/go-init-common/default/http/server"
	"gitlab.com/go-init/go-init-common/default/logger"

	request_repo "go-init/internal/database/request_repo"
)

type App struct {
	cfg            *config.AppConfig
	log            *logger.Logger
	db             *database.AgentImpl
	dbManagerRepo  dbRepo.GoInitManagerRepository
	KafkaProducer  *commonKafka.ClientConfig
	graphqlService *graphql.Service
	srv            *http.Server
	grpcServer     *grpcpkg.GRPCServer
}

const (
	serviceName     = "go-init-manager"
	shutDownTimeOut = time.Second * 5
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
		a.initDb,
		a.initKafka,
		a.initGrpcServer,
		a.initServices,
		a.initHttpServer,
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
	k, err := commonKafka.NewClientConfig(&a.cfg.Kafka, a.log)
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

func (a *App) initHttpServer(ctx context.Context) error {
	// 1. Собираем ExecutableSchema из вашего проекта,
	//    предполагая, что у вас есть graph.NewExecutableSchema() и свой Resolver
	schema := generatedGQL.NewExecutableSchema(
		generatedGQL.Config{
			Resolvers: &generatedGQL.Resolver{
				Service: a.graphqlService,
			},
		},
	)

	// 2. Создаём кастомный GraphQL-хендлер через пакет mygraphql
	gqlHandler := myserver.NewGraphQLServer(schema)

	// 3. Подготовим ( handler для метрик.
	//    Когда захотите Prometheus / OTEL - тут подключаете
	metricsHandler := http.NotFoundHandler()

	// 4. Собираем middleware (логирование, CORS и т.д.) через пакет myhttp
	middlewares := myhttp.CollectHandlers(
	// например, myAuthMiddleware, myLoggerMiddleware...
	)

	s := myserver.NewServer(
		&a.cfg.HttpServ,
		gqlHandler,
		metricsHandler,
		middlewares,
	)
	closer.Add(func() error {
		cancelCtx, cancel := context.WithTimeout(ctx, shutDownTimeOut)
		defer cancel()
		if err := s.Shutdown(cancelCtx); err != nil {
			return fmt.Errorf("faild to stop server: %w", err)
		}
		a.log.Info("Http server stopped")
		return nil
	})

	a.srv = &s.Server
	return nil
}

func (a *App) runHttpServer() error {
	a.log.Info(fmt.Sprintf("Запуск HTTP сервера на %s", a.srv.Addr))
	err := a.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		a.log.Error(fmt.Sprintf("Ошибка HTTP сервера: %v", err))
		return err
	}
	return nil
}

func (a *App) initGrpcServer(_ context.Context) error {
	grpcServer, err := grpcpkg.NewGRPCServer(a.cfg.GrpcServ)
	if err != nil {
		return fmt.Errorf("failed to initialize gRPC server: %w", err)
	}
	a.grpcServer = grpcServer
	return nil
}

func (a *App) runGrpcServer() error {
	a.log.Info(fmt.Sprintf("Запуск gRPC сервера на %s", a.cfg.GrpcServ.Port))
	if err := a.grpcServer.Start(a.cfg.GrpcServ.Port); err != nil {
		a.log.Error(fmt.Sprintf("Ошибка gRPC сервера: %v", err))
		return err
	}
	return nil
}

func (a *App) initServices(_ context.Context) error {
	if a.log == nil || a.cfg == nil || a.db == nil || a.KafkaProducer == nil {
		a.log.Error("One or more dependencies are not initialized")
		return fmt.Errorf("dependencies not initialized")
	}

	// Initialize the repository with schema name
	dbManagerRepo := request_repo.NewRepository(a.db, a.log, a.cfg.Database.Schema)
	a.dbManagerRepo = dbManagerRepo

	// Initialize the GraphQL service
	a.graphqlService = graphql.New(a.log, a.cfg.HttpServ.Name, dbManagerRepo, a.db, a.KafkaProducer)

	// Reuse the same repository instance for Kafka consumers
	// Initialize the Kafka consumer for archive-ready events
	archiveConsumer := kafka.NewArchiveConsumerService(a.log, dbManagerRepo)

	// Register the archive consumer with Kafka
	if a.KafkaProducer != nil && a.KafkaProducer.ConsumerIsEnabled() {
		// Проверяем наличие топиков в конфигурации
		if len(a.cfg.Kafka.ConsumerConfig.Topic) > 0 {
			// Получаем ID топика из конфигурации
			topicID := a.cfg.Kafka.ConsumerConfig.Topic[0].Id
			a.log.InfoContext(context.Background(), "Registering archive consumer for topic ID",
				logger.String("topic_id", topicID))

			// Регистрируем consumer для этого топика
			err := a.KafkaProducer.RegisterConsumerWorkersByTopic(topicID, archiveConsumer)
			if err != nil {
				a.log.ErrorContext(context.Background(), "Failed to register archive Kafka consumer",
					logger.Error(err))
				return err
			}
			a.log.InfoContext(context.Background(), "Archive consumer registered successfully",
				logger.String("topic_id", topicID))
		} else {
			a.log.ErrorContext(context.Background(), "No Kafka topics found in configuration")
			return fmt.Errorf("no Kafka topics found in configuration")
		}

		// Start consuming messages in a separate goroutine
		a.log.InfoContext(context.Background(), "Starting Kafka consumer...")
		go a.KafkaProducer.Start(context.Background())
		a.log.InfoContext(context.Background(), "Kafka consumer started successfully")
	} else {
		a.log.InfoContext(context.Background(), "Kafka consumer is disabled, skipping archive consumer registration")
	}

	return nil
}

func (a *App) initDb(ctx context.Context) error {
	agent, err := database.NewAgent(&a.cfg.Database, a.log)
	a.db = agent

	if err != nil {
		a.log.Error(err.Error())
		return err
	}

	if a.cfg.Database.AutoMigrate {
		err := a.db.Migrate(
			models.Models...,
		)
		if err != nil {
			a.log.Error(fmt.Sprintf("Migration failed: %v", err))
			return err
		}
	}
	return nil
}

func (a *App) Run() error {
	defer func() {
		closer.CloseAll()
		closer.Wait()
	}()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		err := a.runHttpServer()
		if err != nil {
			a.log.Error(fmt.Sprintf("Failed to start http server: %v", err))
		}
	}()

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
