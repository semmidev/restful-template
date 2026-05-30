package app

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/semmidev/restful-template/internal/config"
	delivery "github.com/semmidev/restful-template/internal/delivery/http"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
	"github.com/semmidev/restful-template/internal/shared/database"
	jwtpkg "github.com/semmidev/restful-template/internal/shared/jwt"
	"github.com/semmidev/restful-template/internal/shared/observability"
	redispkg "github.com/semmidev/restful-template/internal/shared/redis"
)

// Setup wires all application dependencies and returns an http.Handler
// along with a cleanup function to close resources (like DB pools).
func Setup(ctx context.Context, cfg config.Config, logger *slog.Logger) (http.Handler, func(), error) {
	// Initialize Postgres Pool
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		return nil, nil, err
	}

	// run migrations only when explicitly enabled (default: true).
	// In multi-replica deployments set DATABASE_RUN_MIGRATIONS=false and run
	// migrations as a separate init-container / CLI step to avoid advisory-lock
	// contention and startup latency.
	if cfg.Database.RunMigrations {
		if err := database.RunMigrations(cfg.Database.DSN, "up"); err != nil {
			logger.Error("migrate failed", "err", err)
			pool.Close()
			return nil, nil, err
		}
	}

	// Initialize Redis
	rdb, limiter, err := redispkg.NewClient(ctx, cfg.Redis.DSN)
	if err != nil {
		logger.Error("redis connect failed", "err", err)
		pool.Close()
		return nil, nil, err
	}

	cleanup := func() {
		if rdb != nil {
			_ = rdb.Close()
		}
		pool.Close()
	}

	// Repositories
	userRepo := auth.NewUserRepository(pool)
	todoRepo := todos.NewTodoRepository(pool)
	tokenRepo := auth.NewTokenRepository(pool)
	cacheRepo := redispkg.NewCacheRepository(rdb)

	// Services & Adapters
	// pass issuer and audience so JWTs carry iss/aud claims
	tokenSvc := jwtpkg.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)
	tracerAdapter := observability.NewOtelTracer("usecase")
	txManager := database.NewPostgresTxManager(pool)

	// Usecases
	todoSvc := todos.NewTodo(todoRepo, cacheRepo, tracerAdapter)
	authSvc := auth.NewAuth(userRepo, tokenSvc, tokenRepo, todoSvc, txManager, tracerAdapter)

	// build health checkers so the /health endpoint probes real deps
	healthCheckers := map[string]delivery.HealthChecker{
		"postgres": func(hctx context.Context) error {
			return pool.Ping(hctx)
		},
		"redis": func(hctx context.Context) error {
			return rdb.Ping(hctx).Err()
		},
	}

	// Server Handler
	server := delivery.NewServer(cfg, logger, authSvc, todoSvc, tokenSvc, limiter, healthCheckers)

	// background goroutine that periodically purges expired refresh
	// tokens from the DB. Without this the table grows unboundedly because
	// DeleteRefreshToken only removes tokens on explicit use (rotation).
	// The goroutine respects ctx cancellation and exits on graceful shutdown.
	go startRefreshTokenCleanup(ctx, pool, logger)

	return server.Handler(), cleanup, nil
}

// startRefreshTokenCleanup runs a periodic job that deletes expired refresh
// tokens from the `refresh_tokens` table. It runs hourly and stops when
// ctx is cancelled (i.e., on graceful shutdown).
func startRefreshTokenCleanup(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Run once immediately at startup to clear any accumulated backlog
	doCleanup(ctx, pool, logger)

	for {
		select {
		case <-ticker.C:
			doCleanup(ctx, pool, logger)
		case <-ctx.Done():
			return
		}
	}
}

func doCleanup(ctx context.Context, pool *pgxpool.Pool, logger *slog.Logger) {
	tag, err := pool.Exec(ctx, "DELETE FROM refresh_tokens WHERE expires_at < NOW()")
	if err != nil {
		logger.Error("refresh token cleanup failed", "err", err)
		return
	}
	if n := tag.RowsAffected(); n > 0 {
		logger.Info("cleaned up expired refresh tokens", "count", n)
	}
}
