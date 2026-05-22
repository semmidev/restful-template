package tests

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/infrastructure/database"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func SetupTestDatabase(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()

	dbContainer, err := postgres.Run(ctx,
		"docker.io/postgres:18-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := dbContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	// Register the uuidv7() function before running migrations.
	pcfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		t.Fatalf("failed to parse pool config: %v", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Run all embedded SQL migrations.
	if err := database.RunMigrations(connStr, "up"); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	cleanup := func() {
		pool.Close()
		if err := dbContainer.Terminate(ctx); err != nil {
			t.Logf("warning: failed to terminate postgres container: %v", err)
		}
	}

	return pool, cleanup
}
