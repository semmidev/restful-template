package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/semmidev/restful-template/internal/config"
	delivery "github.com/semmidev/restful-template/internal/delivery/http"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
	"github.com/semmidev/restful-template/internal/shared/asynqtask"
	"github.com/semmidev/restful-template/internal/shared/database"
	jwtpkg "github.com/semmidev/restful-template/internal/shared/jwt"
	"github.com/semmidev/restful-template/internal/shared/observability"
	redispkg "github.com/semmidev/restful-template/internal/shared/redis"
)

// Setup wires all application dependencies and returns an http.Handler
// along with a cleanup function to close resources (like DB pools).
func Setup(ctx context.Context, cfg config.Config, logger *slog.Logger) (http.Handler, func(), error) {
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		return nil, nil, err
	}

	// Multi-replica deploys set DATABASE_RUN_MIGRATIONS=false and run migrations
	// as an init-container to avoid advisory-lock contention and startup latency.
	if cfg.Database.RunMigrations {
		if err := database.RunMigrations(cfg.Database.DSN, "up"); err != nil {
			logger.Error("migrate failed", "err", err)
			pool.Close()
			return nil, nil, err
		}
	}

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

	userRepo := auth.NewUserRepository(pool)
	todoRepo := todos.NewTodoRepository(pool)
	tokenRepo := auth.NewTokenRepository(pool)
	cacheRepo := redispkg.NewCacheRepository(rdb)

	// issuer and audience are embedded in JWT claims so tokens issued in one
	// environment are rejected in another (prevents substitution attacks).
	tokenSvc := jwtpkg.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.AccessTTL,
		cfg.JWT.RefreshTTL,
		cfg.JWT.Issuer,
		cfg.JWT.Audience,
	)
	tracerAdapter := observability.NewOtelTracer("usecase")
	txManager := database.NewPostgresTxManager(pool)

	todoSvc := todos.NewTodo(todoRepo, cacheRepo, tracerAdapter)

	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.DSN)
	if err != nil {
		logger.Error("invalid redis dsn for asynq", "err", err)
		return nil, nil, err
	}
	clientOpt, ok := redisOpt.(asynq.RedisClientOpt)
	if !ok {
		err := errors.New("parsed redis dsn is not a RedisClientOpt")
		logger.Error("invalid redis opt type", "err", err)
		return nil, nil, err
	}
	distributor := asynqtask.NewDistributor(clientOpt)
	authDistributor := auth.NewTaskDistributor(distributor)

	authSvc := auth.NewAuth(userRepo, tokenSvc, tokenRepo, todoSvc, txManager, tracerAdapter, authDistributor)

	healthCheckers := map[string]delivery.HealthChecker{
		"postgres": func(hctx context.Context) error {
			return pool.Ping(hctx)
		},
		"redis": func(hctx context.Context) error {
			return rdb.Ping(hctx).Err()
		},
	}

	server := delivery.NewServer(cfg, logger, authSvc, todoSvc, tokenSvc, limiter, healthCheckers)

	return server.Handler(), cleanup, nil
}

// SetupWorker wires task handlers and returns a ready-to-start Processor.
// All module task handlers are registered here via AddTask — this is the only
// place that knows about both the Processor and each module's handler functions.
// Mirrors Setup() so cmd/worker/main.go stays as thin as cmd/server/main.go.
func SetupWorker(cfg config.Config, logger *slog.Logger) (asynqtask.Processor, error) {
	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.DSN)
	if err != nil {
		return nil, fmt.Errorf("invalid redis dsn: %w", err)
	}

	clientOpt, ok := redisOpt.(asynq.RedisClientOpt)
	if !ok {
		return nil, errors.New("parsed redis dsn is not a RedisClientOpt")
	}

	processor := asynqtask.NewProcessor(clientOpt, logger)

	// ── auth module tasks ────────────────────────────────────────────────────
	processor.AddTask(auth.TaskSendWelcomeEmail, auth.HandleSendWelcomeEmail(logger))

	return processor, nil
}
