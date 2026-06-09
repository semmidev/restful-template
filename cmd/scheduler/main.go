package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
	"github.com/semmidev/restful-template/internal/shared/banner"
	"github.com/semmidev/restful-template/internal/shared/database"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/shared/redis"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg := config.Load()
	logger := observability.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.App.Env)
	slog.SetDefault(logger)

	shutdownTelemetry, err := observability.InitTelemetry(ctx, "scheduler", cfg.App.Version, cfg.Telemetry.OTLPEndpoint)
	if err != nil {
		logger.Error("failed to initialize telemetry", "err", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				logger.Error("failed to shutdown telemetry", "err", err)
			}
		}()
	}

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("db connect failed: %w", err)
	}
	defer pool.Close()

	rdb, _, err := redis.NewClient(ctx, cfg.Redis.DSN)
	if err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer func() { _ = rdb.Close() }()

	tokenRepo := auth.NewTokenRepository(pool)
	authJob := auth.NewAuthJob(tokenRepo, logger)

	todoRepo := todos.NewTodoRepository(pool)
	cacheRepo := redis.NewCacheRepository(rdb)
	todoJob := todos.NewTodoJob(todoRepo, cacheRepo, logger)

	// 5-minute TTL prevents a dead scheduler replica from holding a lock past a
	// reasonable job window and blocking the next election.
	locker := redis.NewRedisLocker(rdb, "gocron:lock:", 5*time.Minute)

	s, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
		gocron.WithLogger(logger),
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Run at the top of every hour.
	_, err = s.NewJob(
		gocron.CronJob("0 * * * *", false),
		gocron.NewTask(authJob.CleanupExpiredTokens),
	)
	if err != nil {
		return fmt.Errorf("failed to register cleanup job: %w", err)
	}

	// Run every 10 minutes.
	_, err = s.NewJob(
		gocron.CronJob("*/10 * * * *", false),
		gocron.NewTask(todoJob.EscalateUrgency),
	)
	if err != nil {
		return fmt.Errorf("failed to register todo escalation job: %w", err)
	}

	banner.Print(cfg.App.Name+" (Scheduler)", cfg.App.Version, []banner.Field{
		{Key: "env", Value: cfg.App.Env},
		{Key: "jobs", Value: "2 (cleanup tokens, escalate urgency)"},
		{Key: "lock", Value: "redis (5m TTL)"},
	})

	s.Start()
	logger.Info("scheduler started successfully", "jobs", len(s.Jobs()))

	// Run once immediately so tokens/urgency that expired/changed while the scheduler was down
	// are cleared/escalated without waiting.
	authJob.CleanupExpiredTokens()
	todoJob.EscalateUrgency()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	if err := s.Shutdown(); err != nil {
		logger.Error("scheduler shutdown failed", "err", err)
		return fmt.Errorf("scheduler shutdown failed: %w", err)
	}

	logger.Info("scheduler stopped")
	return nil
}
