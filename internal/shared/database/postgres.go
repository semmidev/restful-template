package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/exaring/otelpgx"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/migrations"
)

func NewDB(ctx context.Context, cfg config.Database) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}

	connConfig.Tracer = otelpgx.NewTracer()

	connStr := stdlib.RegisterConnConfig(connConfig)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return sqlx.NewDb(db, "pgx"), nil
}

func RunMigrations(dsn string, direction string) error {
	d, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return err
	}
	defer func() {
		_, _ = m.Close()
	}()

	if direction == "down" {
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	} else {
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return err
		}
	}

	return nil
}
