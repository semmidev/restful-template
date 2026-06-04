package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/hibiken/asynq"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/banner"
	"github.com/semmidev/restful-template/internal/shared/observability"
	"github.com/semmidev/restful-template/internal/worker"
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

	shutdownTelemetry, err := observability.InitTelemetry(ctx, cfg.App.Name+"-worker", cfg.App.Version, cfg.Telemetry.OTLPEndpoint)
	if err != nil {
		logger.Error("failed to initialize telemetry", "err", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				logger.Error("failed to shutdown telemetry", "err", err)
			}
		}()
	}

	redisOpt, err := asynq.ParseRedisURI(cfg.Redis.DSN)
	if err != nil {
		return fmt.Errorf("failed to parse redis dsn: %w", err)
	}

	clientOpt, ok := redisOpt.(asynq.RedisClientOpt)
	if !ok {
		return errors.New("parsed redis dsn is not a RedisClientOpt")
	}

	taskProcessor := worker.NewRedisTaskProcessor(
		clientOpt,
		logger,
	)

	banner.Print(cfg.App.Name, cfg.App.Version, []banner.Field{
		{Key: "env", Value: cfg.App.Env},
		{Key: "service", Value: "worker"},
		{Key: "redis_dsn", Value: cfg.Redis.DSN},
	})

	errChan := make(chan error, 1)
	go func() {
		logger.Info("worker starting")
		if err := taskProcessor.Start(); err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errChan:
		return fmt.Errorf("worker error: %w", err)
	}

	taskProcessor.Shutdown()

	logger.Info("worker stopped")
	return nil
}
