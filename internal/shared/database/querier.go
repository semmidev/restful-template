package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Querier defines the common methods between *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// GetDB returns the pgx.Tx from context if a transaction was injected by TxManager;
// otherwise it returns the default pool. Repositories call this on every query
// so transaction propagation via context is transparent to callers.
func GetDB(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx := ExtractTx(ctx); tx != nil {
		return tx
	}
	return pool
}
