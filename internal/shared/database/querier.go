package database

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/shared/infrastructure"
)

// Querier defines the common methods between *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type circuitBreakerQuerier struct {
	q Querier
}

func (c *circuitBreakerQuerier) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	res, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return c.q.Exec(ctx, sql, arguments...)
	})
	if err != nil {
		return pgconn.CommandTag{}, err
	}
	tag, ok := res.(pgconn.CommandTag)
	if !ok {
		return pgconn.CommandTag{}, nil
	}
	return tag, nil
}

func (c *circuitBreakerQuerier) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	res, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return c.q.Query(ctx, sql, args...)
	})
	if err != nil {
		return nil, err
	}
	rows, ok := res.(pgx.Rows)
	if !ok {
		return nil, nil
	}
	return rows, nil
}

type circuitBreakerRow struct {
	row pgx.Row
}

func (r *circuitBreakerRow) Scan(dest ...any) error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, r.row.Scan(dest...)
	})
	return err
}

func (c *circuitBreakerQuerier) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	row := c.q.QueryRow(ctx, sql, args...)
	return &circuitBreakerRow{row: row}
}

// GetDB returns the pgx.Tx from context if a transaction was injected by TxManager;
// otherwise it returns the default pool. Repositories call this on every query
// so transaction propagation via context is transparent to callers.
// It wraps the returned Querier in a Circuit Breaker to protect the database.
func GetDB(ctx context.Context, pool *pgxpool.Pool) Querier {
	var q Querier = pool
	if tx := ExtractTx(ctx); tx != nil {
		q = tx
	}
	return &circuitBreakerQuerier{q: q}
}
