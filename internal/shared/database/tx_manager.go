package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

// InjectTx injects a pgx.Tx into the context.
func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// ExtractTx retrieves a pgx.Tx from the context if it exists.
func ExtractTx(ctx context.Context) pgx.Tx {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return nil
}

// TxManager defines the interface for running operations within a transaction.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// PostgresTxManager provides transaction management using pgxpool.
type PostgresTxManager struct {
	pool *pgxpool.Pool
}

// NewPostgresTxManager creates a new transaction manager.
func NewPostgresTxManager(pool *pgxpool.Pool) *PostgresTxManager {
	return &PostgresTxManager{pool: pool}
}

// RunInTx executes the given function within a database transaction.
// If the function returns an error or panics, the transaction is rolled back.
func (tm *PostgresTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	ctxWithTx := InjectTx(ctx, tx)

	if err := fn(ctxWithTx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
