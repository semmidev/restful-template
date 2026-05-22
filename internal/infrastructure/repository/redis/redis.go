package redis

import (
	"context"

	"github.com/go-redis/redis_rate/v10"
	"github.com/redis/go-redis/v9"
)

// NewClient establishes a connection to Redis and creates a rate limiter.
func NewClient(ctx context.Context, dsn string) (*redis.Client, *redis_rate.Limiter, error) {
	redisOpts, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, nil, err
	}
	rdb := redis.NewClient(redisOpts)

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, nil, err
	}

	limiter := redis_rate.NewLimiter(rdb)

	return rdb, limiter, nil
}
