package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/infrastructure/database"
)

// Querier defines the common methods between *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// getDb returns the pgx.Tx from context if it exists, otherwise returns the default pool.
func getDb(ctx context.Context, pool *pgxpool.Pool) Querier {
	if tx := database.ExtractTx(ctx); tx != nil {
		return tx
	}
	return pool
}
