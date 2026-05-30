package app

import (
	"context"
	"log/slog"
	"net/http"

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

	return server.Handler(), cleanup, nil
}


