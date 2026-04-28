package pool

import (
	"context"
	"time"

	"gitlab.com/go-init/go-init-common/default/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.com/go-init/go-init-common/default/db/pg"
)

type pgClient struct {
	db pg.DB // Используем интерфейс без указателя
}

// New создает новый клиент для работы с пулом соединений PostgreSQL
func New(ctx context.Context, conf *pg.Config, log *logger.Logger) (pg.Client, error) {
	dsn := pg.BuildDsn(conf)

	pool, err := createPool(ctx, dsn)
	if err != nil {
		log.ErrorContext(ctx, "failed to create pool", logger.Error(err))
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		log.ErrorContext(ctx, "failed to ping database", logger.Error(err))
		pool.Close() // Закрываем пул в случае ошибки
		return nil, err
	}

	client := &pgClient{
		db: &pgPool{
			dbc: pool,
			log: log,
		},
	}
	return client, nil
}

// createPool создает пул соединений с настройками
func createPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	// Настройки пула (можно вынести в конфигурацию)
	config.MaxConnLifetime = 20 * time.Minute          // Максимальное количество соединений
	config.MaxConnIdleTime = 5 * time.Minute           // Минимальное количество соединений
	config.HealthCheckPeriod = 1 * time.Minute         // Максимальное время жизни соединения
	config.ConnConfig.ConnectTimeout = 5 * time.Second // Время простоя соединения

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}
	return pool, nil
}

// DB возвращает интерфейс для работы с базой данных
func (c *pgClient) DB() pg.DB {
	return c.db
}

// Close закрывает пул соединений
func (c *pgClient) Close() {
	if c.db != nil {
		c.db.Close()
	}
}
