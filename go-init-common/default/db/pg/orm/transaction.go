package orm

import (
	"context"
	"fmt"

	"gitlab.com/go-init/go-init-common/default/logger"
	"gorm.io/gorm"
)

// Transaction wraps a GORM transaction with a logger
type Transaction struct {
	Log *logger.Logger
	Tx  *gorm.DB
}

func (t *Transaction) Enfold(ctx context.Context, err *error) {
	if r := recover(); r != nil {
		t.Log.ErrorContext(ctx, "Unexpected panic", logger.Any("recovery", r))

		*err = t.Rollback()

		if *err != nil {
			t.Log.ErrorContext(ctx, "Tx rollback failed", logger.Error(*err))
		}

		return
	}

	if *err != nil {
		*err = t.Rollback()

		if *err != nil {
			t.Log.ErrorContext(ctx, "Tx rollback failed", logger.Error(*err))
		}
		return
	}

	*err = t.Commit()

	if *err != nil {
		t.Log.ErrorContext(ctx, "Tx commit failed", logger.Error(*err))
	}
}

// Commit commits the transaction
func (t *Transaction) Commit() error {
	t.Log.Debug("Committing transaction")
	return t.Tx.Commit().Error
}

// Rollback rolls back the transaction
func (t *Transaction) Rollback() error {
	t.Log.Debug("Rolling back transaction")
	return t.Tx.Rollback().Error
}

// GetOne retrieves a single record with optional preloading
func GetOne[T PersInterface](ctx context.Context, tx *Transaction, model T, preload PersInterface) (T, error) {
	tx.Log.DebugContext(ctx, "Fetching single record", logger.Any("model", model.Name()))
	var result T
	err := tx.Tx.WithContext(ctx).Preload(preload.String()).First(&result).Error
	if err != nil {
		tx.Log.ErrorContext(ctx, "Failed to get one", logger.Error(err))
		return result, fmt.Errorf("failed to get one: %w", err)
	}
	return result, nil
}

// GetMany retrieves multiple records with optional preloading
func GetMany[T PersInterface](ctx context.Context, tx *Transaction, model T, preload PersInterface) ([]T, error) {
	tx.Log.DebugContext(ctx, "Fetching multiple records", logger.Any("model", model.Name()))
	var results []T
	err := tx.Tx.WithContext(ctx).Preload(preload.String()).Find(&results).Error
	if err != nil {
		tx.Log.ErrorContext(ctx, "Failed to get many", logger.Error(err))
		return nil, fmt.Errorf("failed to get many: %w", err)
	}
	return results, nil
}

// Insert creates a new record
func (t *Transaction) Insert(ctx context.Context, model PersInterface) (GenericID, error) {
	t.Log.DebugContext(ctx, "Inserting record", logger.Any("model", model.Name()))
	err := t.Tx.WithContext(ctx).Create(model).Error
	if err != nil {
		t.Log.ErrorContext(ctx, "Failed to insert", logger.Error(err))
		return nil, fmt.Errorf("failed to insert: %w", err)
	}
	return model.GenericID(), nil
}

// FullUpdate fully updates an existing record
func (t *Transaction) FullUpdate(ctx context.Context, model PersInterface) (GenericID, error) {
	t.Log.DebugContext(ctx, "Fully updating record", logger.Any("model", model.Name()))
	err := t.Tx.WithContext(ctx).Save(model).Error
	if err != nil {
		t.Log.ErrorContext(ctx, "Failed to full update", logger.Error(err))
		return nil, fmt.Errorf("failed to full update: %w", err)
	}
	return model.GenericID(), nil
}

// Update updates specific fields of an existing record
func (t *Transaction) Update(ctx context.Context, model PersInterface, fields ...string) (GenericID, error) {
	t.Log.DebugContext(ctx, "Updating record", logger.Any("model", model.Name()), logger.Any("fields", fields))
	err := t.Tx.WithContext(ctx).Model(model).Select(fields).Updates(model).Error
	if err != nil {
		t.Log.ErrorContext(ctx, "Failed to update", logger.Error(err))
		return nil, fmt.Errorf("failed to update: %w", err)
	}
	return model.GenericID(), nil
}

// Delete removes a record
func (t *Transaction) Delete(ctx context.Context, model PersInterface) (GenericID, error) {
	t.Log.DebugContext(ctx, "Deleting record", logger.Any("model", model.Name()))
	err := t.Tx.WithContext(ctx).Delete(model).Error
	if err != nil {
		t.Log.ErrorContext(ctx, "Failed to delete", logger.Error(err))
		return nil, fmt.Errorf("failed to delete: %w", err)
	}
	return model.GenericID(), nil
}
