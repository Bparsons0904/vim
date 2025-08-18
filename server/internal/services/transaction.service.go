package services

import (
	"context"
	"fmt"
	"server/internal/database"
	"server/internal/logger"

	"gorm.io/gorm"
)

type contextKey string

const TRANSACTION_KEY contextKey = "transaction"

// Context utilities for transaction injection
func WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, TRANSACTION_KEY, tx)
}

func GetTransaction(ctx context.Context) (*gorm.DB, bool) {
	tx, ok := ctx.Value(TRANSACTION_KEY).(*gorm.DB)
	return tx, ok
}

// TransactionService handles database transactions using context injection
type TransactionService struct {
	db  database.DB
	log logger.Logger
}

func NewTransactionService(db database.DB) *TransactionService {
	return &TransactionService{
		db:  db,
		log: logger.New("TransactionService"),
	}
}

// Execute runs the provided function within a database transaction
// The function receives a context with the transaction injected
// Automatically handles commit/rollback based on function result
// Panics are converted to errors unless rollback fails (which crashes service for data safety)
func (ts *TransactionService) Execute(ctx context.Context, fn func(context.Context) error) (err error) {
	log := ts.log.Function("Execute")

	// Begin transaction
	tx := ts.db.SQLWithContext(ctx).Begin()
	if tx.Error != nil {
		return log.Err("failed to begin transaction", tx.Error)
	}

	// Create context with transaction
	txCtx := WithTransaction(ctx, tx)

	// Handle panics with intelligent rollback
	defer func() {
		if r := recover(); r != nil {
			panicErr := fmt.Errorf("panic during transaction: %v", r)
			log.Er("panic during transaction, rolling back", panicErr)

			if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
				// Critical failure - data integrity at risk, crash service
				log.Er("CRITICAL: failed to rollback after panic", rollbackErr, "panic", r)
				panic(fmt.Sprintf("transaction rollback failed: %v (original panic: %v)", rollbackErr, r))
			}

			// Rollback succeeded - convert panic to error for graceful handling
			log.Info("transaction rolled back successfully after panic")
			err = panicErr
		}
	}()

	// Execute the function with transaction context
	if err := fn(txCtx); err != nil {
		if rollbackErr := tx.Rollback().Error; rollbackErr != nil {
			log.Er("failed to rollback transaction", rollbackErr, "originalError", err.Error())
		} else {
			log.Info("transaction rolled back due to error", "error", err.Error())
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return log.Err("failed to commit transaction", err)
	}

	log.Info("transaction completed successfully")
	return nil
}