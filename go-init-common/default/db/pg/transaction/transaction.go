package transaction

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.com/go-init/go-init-common/default/db/pg"
	"gitlab.com/go-init/go-init-common/default/db/pg/pool"
)

// manager реализует управление транзакциями
type manager struct {
	db pg.Transactor
}

// NewTransactionManager создает новый менеджер транзакций
func NewTransactionManager(db pg.Transactor) *manager {
	return &manager{db: db}
}

func (m *manager) transaction(ctx context.Context, opts pgx.TxOptions, fn pg.Handler) (err error) {
	// Проверяем, есть ли уже транзакция в контексте
	tx, ok := ctx.Value(pool.TxKey).(pgx.Tx)
	if ok {
		// Если транзакция уже есть, выполняем fn в текущем контексте
		return fn(ctx)
	}

	// Начинаем новую транзакцию
	tx, err = m.db.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Создаем контекст с транзакцией
	txCtx := pool.MakeContext(ctx, tx)

	// Отложенный блок для обработки паники и завершения транзакции
	defer func() {
		if p := recover(); p != nil {
			// Если произошла паника, откатываем транзакцию
			if rollbackErr := tx.Rollback(txCtx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback after panic: %w (panic: %v)", rollbackErr, p)
			} else {
				err = fmt.Errorf("transaction panicked: %v", p)
			}
			// Повторно вызываем панику, чтобы не прерывать выполнение
			panic(p)
		} else if err != nil {
			// Если есть ошибка, откатываем транзакцию
			if rollbackErr := tx.Rollback(txCtx); rollbackErr != nil {
				err = fmt.Errorf("failed to rollback: %w (original error: %v)", rollbackErr, err)
			}
		} else {
			// Если ошибок нет, коммитим транзакцию
			if commitErr := tx.Commit(txCtx); commitErr != nil {
				err = fmt.Errorf("failed to commit: %w", commitErr)
			}
		}
	}()

	// Выполняем переданную функцию
	err = fn(txCtx)
	return err
}

// ReadCommitted выполняет функцию handler в транзакции с уровнем изоляции "Read Committed"
func (m *manager) ReadCommitted(ctx context.Context, handler pg.Handler) error {
	return m.transaction(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}, handler)
}
