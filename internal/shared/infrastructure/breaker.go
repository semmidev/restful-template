package infrastructure

import (
	"time"

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
	}
	RedisBreaker = gobreaker.NewCircuitBreaker[any](redisSettings)
}
