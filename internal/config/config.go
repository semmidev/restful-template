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
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App       App
	HTTP      HTTP
	Database  Database
	Redis     Redis
	JWT       JWT
	Log       Log
	Telemetry Telemetry
	CORS      CORS
	SMTP      SMTP
	Asynqmon  Asynqmon
	Google    Google
}

type Google struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

type Asynqmon struct {
	Username string
	Password string
}

type SMTP struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
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
	RunMigrations   bool // set false to skip auto-migration (e.g. in multi-replica deploys)
}

type Redis struct {
	DSN string
}

type JWT struct {
	Secret     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
	Issuer     string // iss claim
	Audience   string // aud claim
}

type Log struct {
	Level  string
	Format string
}

type Telemetry struct {
	OTLPEndpoint string
}

// CORS holds allowed origin configuration.
// AllowedOrigins is a comma-separated list of origins (e.g. "https://app.example.com").
// Use "*" to allow all origins (development only).
type CORS struct {
	AllowedOrigins []string
}

// Load reads configuration from `.env` (if present) and then overlays any
// matching OS environment variables. Defaults are applied for all fields.
func Load() Config {
	v := viper.New()

	v.SetConfigFile(".env")
	v.SetConfigType("dotenv")
	_ = v.ReadInConfig()

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
	v.SetDefault("DATABASE_RUN_MIGRATIONS", false)

	v.SetDefault("REDIS_DSN", "redis://localhost:6379/0")

	v.SetDefault("JWT_SECRET", "change-me-in-production-min-32-bytes!")
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")
	v.SetDefault("JWT_ISSUER", "restful-template")
	v.SetDefault("JWT_AUDIENCE", "restful-template")

	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	v.SetDefault("SMTP_HOST", "localhost")
	v.SetDefault("SMTP_PORT", 1025)
	v.SetDefault("SMTP_USERNAME", "")
	v.SetDefault("SMTP_PASSWORD", "")
	v.SetDefault("SMTP_FROM", "noreply@todo.local")

	v.SetDefault("TELEMETRY_OTLP_ENDPOINT", "localhost:4317")

	v.SetDefault("CORS_ALLOWED_ORIGINS", "*") // override in production

	v.SetDefault("ASYNQMON_USERNAME", "admin")
	v.SetDefault("ASYNQMON_PASSWORD", "admin")

	v.SetDefault("GOOGLE_CLIENT_ID", "")
	v.SetDefault("GOOGLE_CLIENT_SECRET", "")
	v.SetDefault("GOOGLE_REDIRECT_URI", "")

	rawOrigins := v.GetString("CORS_ALLOWED_ORIGINS")
	var allowedOrigins []string
	for _, o := range strings.Split(rawOrigins, ",") {
		o = strings.TrimSpace(o)
		if o != "" {
			allowedOrigins = append(allowedOrigins, o)
		}
	}

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
			RunMigrations:   v.GetBool("DATABASE_RUN_MIGRATIONS"),
		},
		Redis: Redis{
			DSN: v.GetString("REDIS_DSN"),
		},
		JWT: JWT{
			Secret:     v.GetString("JWT_SECRET"),
			AccessTTL:  mustDuration(v, "JWT_ACCESS_TTL"),
			RefreshTTL: mustDuration(v, "JWT_REFRESH_TTL"),
			Issuer:     v.GetString("JWT_ISSUER"),
			Audience:   v.GetString("JWT_AUDIENCE"),
		},
		Log: Log{
			Level:  v.GetString("LOG_LEVEL"),
			Format: v.GetString("LOG_FORMAT"),
		},
		Telemetry: Telemetry{
			OTLPEndpoint: v.GetString("TELEMETRY_OTLP_ENDPOINT"),
		},
		CORS: CORS{
			AllowedOrigins: allowedOrigins,
		},
		SMTP: SMTP{
			Host:     v.GetString("SMTP_HOST"),
			Port:     v.GetInt("SMTP_PORT"),
			Username: v.GetString("SMTP_USERNAME"),
			Password: v.GetString("SMTP_PASSWORD"),
			From:     v.GetString("SMTP_FROM"),
		},
		Asynqmon: Asynqmon{
			Username: v.GetString("ASYNQMON_USERNAME"),
			Password: v.GetString("ASYNQMON_PASSWORD"),
		},
		Google: Google{
			ClientID:     v.GetString("GOOGLE_CLIENT_ID"),
			ClientSecret: v.GetString("GOOGLE_CLIENT_SECRET"),
			RedirectURI:  v.GetString("GOOGLE_REDIRECT_URI"),
		},
	}
}

// mustDuration parses a duration string from viper.
// It panics on misconfiguration — fail-fast at startup is correct for timeouts.
// A zero/missing timeout would silently create an infinite timeout, which is
// a reliability hazard in production.
func mustDuration(v *viper.Viper, key string) time.Duration {
	s := v.GetString(key)
	d, err := time.ParseDuration(s)
	if err != nil {
		panic(fmt.Sprintf("config: invalid duration for %s=%q: %v", key, s, err))
	}
	return d
}
