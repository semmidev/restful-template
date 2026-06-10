package infrastructure

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	"github.com/sony/gobreaker/v2"
)

var (
	// DBBreaker is a shared circuit breaker for PostgreSQL queries.
	// It trips after 5 consecutive failures and stays half-open for 10 seconds.
	DBBreaker *gobreaker.CircuitBreaker[any]

	// RedisBreaker is a shared circuit breaker for Redis cache queries.
	RedisBreaker *gobreaker.CircuitBreaker[any]
)

func init() {
	dbSettings := gobreaker.Settings{
		Name:        "PostgresDB",
		MaxRequests: 3,                // Requests allowed in half-open state
		Interval:    30 * time.Second, // Cyclic period for clearing counts in closed state
		Timeout:     10 * time.Second, // Time in open state before transitioning to half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			// Trip the breaker if we have at least 5 failures consecutively
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: isDBErrorSuccess,
	}
	DBBreaker = gobreaker.NewCircuitBreaker[any](dbSettings)

	redisSettings := gobreaker.Settings{
		Name:        "RedisCache",
		MaxRequests: 3,
		Interval:    15 * time.Second,
		Timeout:     5 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		IsSuccessful: isRedisErrorSuccess,
	}
	RedisBreaker = gobreaker.NewCircuitBreaker[any](redisSettings)
}

func isDBErrorSuccess(err error) bool {
	if err == nil {
		return true
	}
	// pgx.ErrNoRows is a normal query result (row not found), not a DB failure
	if errors.Is(err, pgx.ErrNoRows) {
		return true
	}
	// Client context cancellation is not a database failure
	if errors.Is(err, context.Canceled) {
		return true
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// pgconn.PgError indicates PostgreSQL server is alive and responding.
		// Class 57 (Operator Intervention), Class 58 (System Error), and Class XX (Internal Error)
		// represent database-level crashes or shutdowns, so they count as failures.
		// Other classes (like Class 23 Integrity Violations, Class 22 Data Exceptions, etc.)
		// indicate the database is healthy and responding to queries, so they do NOT count as failures.
		if len(pgErr.Code) >= 2 {
			class := pgErr.Code[:2]
			if class == "57" || class == "58" || class == "XX" {
				return false
			}
		}
		return true
	}

	return false
}

func isRedisErrorSuccess(err error) bool {
	if err == nil {
		return true
	}
	// redis.Nil indicates a cache miss, which is a normal business outcome, not a failure
	if errors.Is(err, redis.Nil) {
		return true
	}
	// Client context cancellation is not a Redis failure
	if errors.Is(err, context.Canceled) {
		return true
	}
	return false
}

