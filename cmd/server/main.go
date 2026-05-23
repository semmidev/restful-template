package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/semmidev/restful-template/internal/config"
	delivery "github.com/semmidev/restful-template/internal/delivery/http"
	"github.com/semmidev/restful-template/internal/infrastructure/jwt"
	"github.com/semmidev/restful-template/internal/infrastructure/repository/redis"
	"github.com/semmidev/restful-template/internal/modules/auth"
	"github.com/semmidev/restful-template/internal/modules/todos"
	"github.com/semmidev/restful-template/internal/shared/database"
	"github.com/semmidev/restful-template/internal/shared/observability"
)

func main() {
	cfg := config.Load()

	logger := observability.NewLogger(cfg.Log.Level, cfg.Log.Format, cfg.App.Env)
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	pool, err := database.NewPool(ctx, cfg.Database)
	if err != nil {
		logger.Error("db connect failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := database.RunMigrations(cfg.Database.DSN, "up"); err != nil {
		logger.Error("migrate failed", "err", err)
		os.Exit(1)
	}

	userRepo := auth.NewUserRepository(pool)
	todoRepo := todos.NewTodoRepository(pool)
	tokenRepo := auth.NewTokenRepository(pool)
	tokenSvc := jwt.NewJWTService(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.RefreshTTL)

	rdb, limiter, err := redis.NewClient(ctx, cfg.Redis.DSN)
	if err != nil {
		logger.Error("redis connect failed", "err", err)
		os.Exit(1)
	}
	defer rdb.Close()

	cacheRepo := redis.NewCacheRepository(rdb)
	tracerAdapter := observability.NewOtelTracer("usecase")

	txManager := database.NewPostgresTxManager(pool)
	todoSvc := todos.NewTodo(todoRepo, cacheRepo, tracerAdapter)
	authSvc := auth.NewAuth(userRepo, tokenSvc, tokenRepo, todoSvc, txManager, tracerAdapter)

	server := delivery.NewServer(cfg, logger, authSvc, todoSvc, tokenSvc, limiter)

	srv := &http.Server{
		Addr:              ":" + cfg.HTTP.Port,
		Handler:           server.Handler(),
		ReadTimeout:       cfg.HTTP.ReadTimeout,
		WriteTimeout:      cfg.HTTP.WriteTimeout,
		IdleTimeout:       cfg.HTTP.IdleTimeout,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("http server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	logger.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "err", err)
	}
	logger.Info("server stopped")
}
