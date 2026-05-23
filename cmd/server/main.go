package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/semmidev/restful-template/internal/app"
	"github.com/semmidev/restful-template/internal/config"
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

	shutdownTelemetry, err := observability.InitTelemetry(ctx, cfg.App.Name, cfg.App.Version, cfg.Telemetry.OTLPEndpoint)
	if err != nil {
		logger.Error("failed to initialize telemetry", "err", err)
	} else {
		defer func() {
			if err := shutdownTelemetry(context.Background()); err != nil {
				logger.Error("failed to shutdown telemetry", "err", err)
			}
		}()
	}

	handler, cleanup, err := app.Setup(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("app setup failed: %w", err)
	}
	defer cleanup()

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           handler,
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errChan := make(chan error, 1)
	go func() {
		logger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-errChan:
		return fmt.Errorf("server error: %w", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	logger.Info("server stopped")
	return nil
}
