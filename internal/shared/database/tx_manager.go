package database

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type txKey struct{}

func InjectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

func ExtractTx(ctx context.Context) *sqlx.Tx {
	if tx, ok := ctx.Value(txKey{}).(*sqlx.Tx); ok {
		return tx
	}
	return nil
}

// TxManager defines the interface for running operations within a transaction.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type PostgresTxManager struct {
	db *sqlx.DB
}

func NewPostgresTxManager(db *sqlx.DB) *PostgresTxManager {
	return &PostgresTxManager{db: db}
}

// RunInTx executes fn inside a database transaction.
func (tm *PostgresTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	tx, err := tm.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-panic so original debugging stack trace is not swallowed
		} else if err != nil {
			_ = tx.Rollback()
		}
	}()

	ctxWithTx := InjectTx(ctx, tx)

	if err = fn(ctxWithTx); err != nil {
		return err
	}

	return tx.Commit()
}
