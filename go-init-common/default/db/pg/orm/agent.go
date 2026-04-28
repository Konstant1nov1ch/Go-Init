package orm

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gitlab.com/go-init/go-init-common/default/db/pg"
	db "gitlab.com/go-init/go-init-common/default/db/pg"
	"gitlab.com/go-init/go-init-common/default/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Init represents the result of a database initialization attempt
type Init struct {
	Conn  *gorm.DB
	Error error
}

// Agent defines the interface for database operations
type Agent interface {
	BeginTx() *Transaction
	Migrate(services ...any) error
}

// AgentImpl is the concrete implementation of the Agent interface
type AgentImpl struct {
	db  *gorm.DB
	log *logger.Logger
	Agent
}

// NewAgent создает новый AgentImpl с GORM-соединением, ожидая максимум timeoutSec
func NewAgent(dbConf *db.Config, log *logger.Logger) (*AgentImpl, error) {
	if dbConf == nil {
		return nil, fmt.Errorf("database config cannot be nil")
	}
	if log == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	wait := dbConf.InitTimeout
	if wait <= 0 {
		wait = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()

	initCh := make(chan Init, 1)

	// Запускаем инициализацию GORM в горутине
	go func() {
		initCh <- newDBConnection(dbConf)
	}()

	// Ждем результат или таймаут
	var initResult Init
	select {
	case initResult = <-initCh:
		// Если при инициализации была ошибка, возвращаем её
		if initResult.Error != nil {
			return nil, fmt.Errorf("failed to init gorm DB: %w", initResult.Error)
		}
	case <-ctx.Done():
		return nil, fmt.Errorf("timeout waiting for gorm initialization: %v", ctx.Err())
	}

	// Успешно инициализировали GORM
	return &AgentImpl{
		db:  initResult.Conn,
		log: log,
	}, nil
}

// init initializes a GORM connection using the provided configuration
func newDBConnection(conf *pg.Config) Init {
	dsn := pg.BuildDsn(conf)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.Schema + ".",
			SingularTable: true,
		},
	})
	if err != nil {
		return Init{
			Conn:  nil,
			Error: err,
		}
	}
	sqlDB, err := db.DB()
	if err != nil {
		return Init{
			Conn:  nil,
			Error: err,
		}
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(0)
	sqlDB.SetConnMaxIdleTime(30 * time.Minute)

	return Init{Conn: db, Error: nil}
}

// BeginTx starts a new database transaction
func (a *AgentImpl) BeginTx(ctx context.Context) (*Transaction, error) {
	tx := a.db.Begin(&sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if tx.Error != nil {
		a.log.ErrorContext(ctx, "Error creating tx", logger.Error(tx.Error))
		return nil, tx.Error
	}

	a.log.DebugContext(ctx, "Tx created")

	return &Transaction{Log: a.log, Tx: tx}, nil
}

// Migrate performs automatic migration for the provided models
func (a *AgentImpl) Migrate(services ...interface{}) error {
	return a.db.AutoMigrate(services...)
}

// DB returns the underlying GORM DB instance
func (a *AgentImpl) DB() *gorm.DB {
	return a.db
}
