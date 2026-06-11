package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/semmidev/restful-template/internal/app"
	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/banner"
	"github.com/semmidev/restful-template/internal/shared/observability"
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

	if err := cfg.Validate(); err != nil {
		logger.Error("configuration validation failed", "err", err)
		return fmt.Errorf("config validation: %w", err)
	}

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

	taskProcessor, err := app.SetupWorker(cfg, logger)
	if err != nil {
		return fmt.Errorf("worker setup: %w", err)
	}

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		taskProcessor.Shutdown()
		close(done)
	}()

	select {
	case <-done:
		logger.Info("worker stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Error("worker shutdown timed out")
		return fmt.Errorf("graceful shutdown timed out")
	}

	return nil
}
