package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/semmidev/restful-template/internal/shared/infrastructure"
)

// Row defines the interface for scanning a single row.
type Row interface {
	Scan(dest ...any) error
}

// Rows defines the interface for iterating and scanning multiple rows.
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// Querier defines the common methods between *sqlx.DB and *sqlx.Tx.
type Querier interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) Row
}

type sqlRowWrapper struct {
	row *sql.Row
}

func (w *sqlRowWrapper) Scan(dest ...any) error {
	return w.row.Scan(dest...)
}

type circuitBreakerRow struct {
	row Row
}

func (r *circuitBreakerRow) Scan(dest ...any) error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, r.row.Scan(dest...)
	})
	return err
}

type circuitBreakerRows struct {
	rows *sql.Rows
}

func (r *circuitBreakerRows) Next() bool {
	res, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return r.rows.Next(), nil
	})
	if err != nil {
		return false
	}
	val, ok := res.(bool)
	if !ok {
		return false
	}
	return val
}

func (r *circuitBreakerRows) Scan(dest ...any) error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, r.rows.Scan(dest...)
	})
	return err
}

func (r *circuitBreakerRows) Close() error {
	return r.rows.Close()
}

func (r *circuitBreakerRows) Err() error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, r.rows.Err()
	})
	return err
}

type circuitBreakerQuerier struct {
	q Querier
}

func (c *circuitBreakerQuerier) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, c.q.GetContext(ctx, dest, query, args...)
	})
	return err
}

func (c *circuitBreakerQuerier) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	_, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		return nil, c.q.SelectContext(ctx, dest, query, args...)
	})
	return err
}

func (c *circuitBreakerQuerier) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	res, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		r, err := c.q.ExecContext(ctx, query, args...)
		return r, err
	})
	if err != nil {
		return nil, err
	}
	val, ok := res.(sql.Result)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (c *circuitBreakerQuerier) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	res, err := infrastructure.DBBreaker.Execute(func() (any, error) {
		r, err := c.q.QueryContext(ctx, query, args...)
		return r, err
	})
	if err != nil {
		return nil, err
	}
	val, ok := res.(Rows)
	if !ok {
		return nil, nil
	}
	return val, nil
}

func (c *circuitBreakerQuerier) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	row := c.q.QueryRowContext(ctx, query, args...)
	return &circuitBreakerRow{row: row}
}

// GetDB returns the *sqlx.Tx from context if a transaction was injected by TxManager;
// otherwise it returns the default *sqlx.DB pool.
// It wraps the returned Querier in a Circuit Breaker to protect the database.
func GetDB(ctx context.Context, db *sqlx.DB) Querier {
	var q Querier = &sqlxDBWrapper{db: db}
	if tx := ExtractTx(ctx); tx != nil {
		q = &sqlxTxWrapper{tx: tx}
	}
	return &circuitBreakerQuerier{q: q}
}

// sqlxDBWrapper adapts *sqlx.DB to the Querier interface.
type sqlxDBWrapper struct {
	db *sqlx.DB
}

func (w *sqlxDBWrapper) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.db.GetContext(ctx, dest, query, args...)
}

func (w *sqlxDBWrapper) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.db.SelectContext(ctx, dest, query, args...)
}

func (w *sqlxDBWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.db.ExecContext(ctx, query, args...)
}

func (w *sqlxDBWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := w.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &circuitBreakerRows{rows: rows}, nil
}

func (w *sqlxDBWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return &sqlRowWrapper{row: w.db.QueryRowContext(ctx, query, args...)}
}

// sqlxTxWrapper adapts *sqlx.Tx to the Querier interface.
type sqlxTxWrapper struct {
	tx *sqlx.Tx
}

func (w *sqlxTxWrapper) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.tx.GetContext(ctx, dest, query, args...)
}

func (w *sqlxTxWrapper) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return w.tx.SelectContext(ctx, dest, query, args...)
}

func (w *sqlxTxWrapper) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return w.tx.ExecContext(ctx, query, args...)
}

func (w *sqlxTxWrapper) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := w.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &circuitBreakerRows{rows: rows}, nil
}

func (w *sqlxTxWrapper) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return &sqlRowWrapper{row: w.tx.QueryRowContext(ctx, query, args...)}
}
