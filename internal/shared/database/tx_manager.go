package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

func InjectTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

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

type PostgresTxManager struct {
	pool *pgxpool.Pool
}

func NewPostgresTxManager(pool *pgxpool.Pool) *PostgresTxManager {
	return &PostgresTxManager{pool: pool}
}

// RunInTx executes fn inside a database transaction.
// The deferred Rollback is a no-op after a successful Commit, so it is safe
// to defer unconditionally — this avoids the common mistake of forgetting
// rollback on every early-return error path.
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
