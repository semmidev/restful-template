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

	// Initialize Postgres Pool
	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("db connect failed: %w", err)
	}
	defer pool.Close()

	// Initialize Redis
	rdb, _, err := redis.NewClient(ctx, cfg.Redis.DSN)
	if err != nil {
		logger.Error("failed to connect to redis", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	// Initialize Repositories and Jobs
	tokenRepo := auth.NewTokenRepository(pool)
	authJob := auth.NewAuthJob(tokenRepo, logger)

	// Initialize Redis Locker (5 minutes TTL to prevent deadlocks)
	locker := redis.NewRedisLocker(rdb, 5*time.Minute)

	// Initialize Scheduler
	s, err := gocron.NewScheduler(
		gocron.WithDistributedLocker(locker),
	)
	if err != nil {
		return fmt.Errorf("failed to create scheduler: %w", err)
	}

	// Register Jobs
	// Run hourly (0th minute of every hour)
	_, err = s.NewJob(
		gocron.CronJob("0 * * * *", false),
		gocron.NewTask(authJob.CleanupExpiredTokens),
	)
	if err != nil {
		return fmt.Errorf("failed to register cleanup job: %w", err)
	}

	// Start Scheduler
	s.Start()
	logger.Info("scheduler started successfully", "jobs", len(s.Jobs()))

	// Run once immediately on startup to clear backlog
	authJob.CleanupExpiredTokens()

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("shutdown signal received")

	// Graceful shutdown of scheduler
	if err := s.Shutdown(); err != nil {
		logger.Error("scheduler shutdown failed", "err", err)
		return fmt.Errorf("scheduler shutdown failed: %w", err)
	}

	logger.Info("scheduler stopped")
	return nil
}
