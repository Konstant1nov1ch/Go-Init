package pool

import (
	"context"

	"gitlab.com/go-init/go-init-common/default/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type key string

const (
	TxKey key = "tx"
)

type pgPool struct {
	dbc *pgxpool.Pool
	log *logger.Logger
}

// Exec выполняет SQL-запрос
func (p *pgPool) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	p.log.DebugContext(ctx, "sql expression", logger.String("sql", sql), logger.Any("args", arguments))
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Exec(ctx, sql, arguments...)
	}
	return p.dbc.Exec(ctx, sql, arguments...)
}

// Query выполняет SQL-запрос и возвращает набор строк
func (p *pgPool) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	p.log.DebugContext(ctx, "sql expression", logger.String("sql", sql), logger.Any("args", args))
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.Query(ctx, sql, args...)
	}
	return p.dbc.Query(ctx, sql, args...)
}

// QueryRow выполняет SQL-запрос и возвращает одну строку
func (p *pgPool) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	p.log.DebugContext(ctx, "sql expression", logger.String("sql", sql), logger.Any("args", args))
	tx, ok := ctx.Value(TxKey).(pgx.Tx)
	if ok {
		return tx.QueryRow(ctx, sql, args...)
	}
	return p.dbc.QueryRow(ctx, sql, args...)
}

// BeginTx начинает новую транзакцию
func (p *pgPool) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	p.log.InfoContext(ctx, "start transaction")
	return p.dbc.BeginTx(ctx, txOptions)
}

// Ping проверяет соединение с базой данных
func (p *pgPool) Ping(ctx context.Context) error {
	return p.dbc.Ping(ctx)
}

// Close закрывает пул соединений
func (p *pgPool) Close() {
	p.log.Info("closing database pool")
	p.dbc.Close()
}

// MakeContext добавляет транзакцию в контекст
func MakeContext(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, TxKey, tx)
}
