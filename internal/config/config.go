// Package config loads application configuration from a .env file and
// environment variables using Viper. Environment variables always take
// precedence over values in the .env file.
//
// Variable naming convention:
//
//	APP_ENV         → Config.App.Env
//	HTTP_PORT       → Config.HTTP.Port
//	DATABASE_DSN    → Config.Database.DSN
//	JWT_SECRET      → Config.JWT.Secret
//	LOG_LEVEL       → Config.Log.Level
package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config is the root application configuration.
type Config struct {
	App       App
	HTTP      HTTP
	Database  Database
	Redis     Redis
	JWT       JWT
	Log       Log
	Telemetry Telemetry
}

type App struct {
	Env         string
	Name        string
	Description string
	Version     string
}

type HTTP struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

type Database struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type Redis struct {
	DSN string
}

type JWT struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Log struct {
	Level  string
	Format string
}

type Telemetry struct {
	OTLPEndpoint string
}

// Load reads configuration from `.env` (if present) and then overlays any
// matching OS environment variables. Defaults are applied for all fields.
func Load() Config {
	v := viper.New()

	// Read .env file (optional — silently ignored if missing)
	v.SetConfigFile(".env")
	v.SetConfigType("dotenv")
	_ = v.ReadInConfig()

	// OS env vars always win (uses the same key names, case-insensitive)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// ── Defaults ─────────────────────────────────────────────────────────────
	v.SetDefault("APP_ENV", "development")
	v.SetDefault("APP_NAME", "restful-template")
	v.SetDefault("APP_DESCRIPTION", "Production-ready Todo REST API built with Huma v2 + Chi.")
	v.SetDefault("APP_VERSION", "1.0.0")

	v.SetDefault("HTTP_PORT", "8080")
	v.SetDefault("HTTP_READ_TIMEOUT", "15s")
	v.SetDefault("HTTP_WRITE_TIMEOUT", "15s")
	v.SetDefault("HTTP_IDLE_TIMEOUT", "60s")
	v.SetDefault("HTTP_SHUTDOWN_TIMEOUT", "10s")

	v.SetDefault("DATABASE_DSN", "postgres://todo:todo@localhost:5432/todo?sslmode=disable")
	v.SetDefault("DATABASE_MAX_OPEN_CONNS", 25)
	v.SetDefault("DATABASE_MAX_IDLE_CONNS", 5)
	v.SetDefault("DATABASE_CONN_MAX_LIFETIME", "5m")

	v.SetDefault("REDIS_DSN", "redis://localhost:6379/0")

	v.SetDefault("JWT_SECRET", "change-me-in-production-min-32-bytes!")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	v.SetDefault("TELEMETRY_OTLP_ENDPOINT", "localhost:4317")

	return Config{
		App: App{
			Env:         v.GetString("APP_ENV"),
			Name:        v.GetString("APP_NAME"),
			Description: v.GetString("APP_DESCRIPTION"),
			Version:     v.GetString("APP_VERSION"),
		},
		HTTP: HTTP{
			Port:            v.GetString("HTTP_PORT"),
			ReadTimeout:     mustDuration(v, "HTTP_READ_TIMEOUT"),
			WriteTimeout:    mustDuration(v, "HTTP_WRITE_TIMEOUT"),
			IdleTimeout:     mustDuration(v, "HTTP_IDLE_TIMEOUT"),
			ShutdownTimeout: mustDuration(v, "HTTP_SHUTDOWN_TIMEOUT"),
		},
		Database: Database{
			DSN:             v.GetString("DATABASE_DSN"),
			MaxOpenConns:    v.GetInt("DATABASE_MAX_OPEN_CONNS"),
			MaxIdleConns:    v.GetInt("DATABASE_MAX_IDLE_CONNS"),
			ConnMaxLifetime: mustDuration(v, "DATABASE_CONN_MAX_LIFETIME"),
		},
		Redis: Redis{
			DSN: v.GetString("REDIS_DSN"),
		},
		JWT: JWT{
			Secret:     v.GetString("JWT_SECRET"),
			AccessTTL:  mustDuration(v, "JWT_ACCESS_TTL"),
			RefreshTTL: mustDuration(v, "JWT_REFRESH_TTL"),
		},
		Log: Log{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
		Telemetry: Telemetry{
			OTLPEndpoint: v.GetString("TELEMETRY_OTLP_ENDPOINT"),
		},
	}
}

// mustDuration parses a duration string from viper, falling back to viper's
// default if the value is invalid. It never panics.
func mustDuration(v *viper.Viper, key string) time.Duration {
	s := v.GetString(key)
	d, err := time.ParseDuration(s)
	if err != nil {
		// Fall back to the registered default
		d, _ = time.ParseDuration(v.GetString(key))
	}
	return d
}
