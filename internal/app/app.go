package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hibiken/asynq"
	"github.com/semmidev/restful-template/internal/adm/user"
	"github.com/semmidev/restful-template/internal/auth"
	"github.com/semmidev/restful-template/internal/config"
	httpserver "github.com/semmidev/restful-template/internal/http"
	"github.com/semmidev/restful-template/internal/shared/asynqtask"
	"github.com/semmidev/restful-template/internal/shared/database"
	"github.com/semmidev/restful-template/internal/shared/email/smtp"
	jwtpkg "github.com/semmidev/restful-template/internal/shared/jwt"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/shared/policy"
	redispkg "github.com/semmidev/restful-template/internal/shared/redis"
	"github.com/semmidev/restful-template/internal/todo"
)

// Setup wires all application dependencies and returns an http.Handler
// along with a cleanup function to close resources (like DB pools).
func Setup(ctx context.Context, cfg config.Config, logger *slog.Logger) (http.Handler, func(), error) {
	db, err := database.NewDB(ctx, cfg.Database)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		return nil, nil, err
	}

	// Multi-replica deploys set DATABASE_RUN_MIGRATIONS=false and run migrations
	// as an init-container to avoid advisory-lock contention and startup latency.
	if cfg.Database.RunMigrations {
		if err := database.RunMigrations(cfg.Database.DSN, "up"); err != nil {
			logger.Error("migrate failed", "err", err)
			_ = db.Close()
			return nil, nil, err
		}
	}

	if err := policy.Init(ctx); err != nil {
		logger.Error("failed to init OPA policy engine", "err", err)
		_ = db.Close()
		return nil, nil, err
	}

	rdb, limiter, err := redispkg.NewClient(ctx, cfg.Redis.DSN)
	if err != nil {
		logger.Error("redis connect failed", "err", err)
		_ = db.Close()
		return nil, nil, err
	}

	var distributor *asynqtask.Distributor

	cleanup := func() {
		if distributor != nil {
			_ = distributor.Close()
		}
		if rdb != nil {
			_ = rdb.Close()
		}
		_ = db.Close()
	}

	userRepo := auth.NewUserRepository(db)
	todoRepo := todo.NewTodoRepository(db)
	tokenRepo := auth.NewTokenRepository(db)
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
	txManager := database.NewPostgresTxManager(db)

	todoSvc := todo.NewTodoService(todoRepo, cacheRepo, tracerAdapter)

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
	distributor = asynqtask.NewDistributor(clientOpt)
	authDistributor := auth.NewTaskDistributor(distributor)

	authSvc := auth.NewAuthService(userRepo, tokenSvc, tokenRepo, todoSvc, txManager, tracerAdapter, authDistributor, cfg.Google)

	usersRepo := user.NewUserRepository(db)
	usersSvc := user.NewUserService(usersRepo, txManager, tracerAdapter)

	healthCheckers := map[string]httpserver.HealthChecker{
		"postgres": func(hctx context.Context) error {
			return db.PingContext(hctx)
		},
		"redis": func(hctx context.Context) error {
			return rdb.Ping(hctx).Err()
		},
	}

	server := httpserver.NewServer(cfg, logger, authSvc, todoSvc, usersSvc, tokenSvc, limiter, healthCheckers, clientOpt)

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

	emailSender, err := smtp.NewSender(cfg.SMTP)
	if err != nil {
		return nil, fmt.Errorf("failed to init smtp sender: %w", err)
	}
	authWorker := auth.NewAuthWorker(logger, emailSender, cfg.App.URL)
	processor.AddTask(auth.TaskSendWelcomeEmail, authWorker.HandleSendWelcomeEmail())

	return processor, nil
}
