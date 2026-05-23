package main

import (
	"log/slog"
	"os"

	"github.com/semmidev/restful-template/internal/config"
	"github.com/semmidev/restful-template/internal/shared/database"
	"github.com/semmidev/restful-template/internal/shared/observability"
)

func main() {
	cfg := config.Load()
	logger := observability.NewLogger(cfg.Log.Level, "text", cfg.App.Env)

	direction := "up"
	if len(os.Args) > 1 {
		direction = os.Args[1]
	}

	if err := database.RunMigrations(cfg.Database.DSN, direction); err != nil {
		logger.Error("migrate", "err", err)
		os.Exit(1)
	}
	slog.Info("migrations applied", "direction", direction)
}
