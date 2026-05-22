package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/semmidev/restful-template/internal/domain"
)

type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new domain.CacheRepository backed by Redis.
func NewCacheRepository(client *redis.Client) domain.CacheRepository {
	return &cacheRepository{client: client}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", domain.ErrNotFound
	}
	return val, err
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}
