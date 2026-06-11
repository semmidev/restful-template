package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/semmidev/restful-template/internal/shared/cache"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	"github.com/semmidev/restful-template/internal/shared/infrastructure"
)

type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository creates a new cache.CacheRepository backed by Redis.
func NewCacheRepository(client *redis.Client) cache.CacheRepository {
	return &cacheRepository{client: client}
}

func (r *cacheRepository) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	_, err := infrastructure.RedisBreaker.Execute(func() (any, error) {
		return nil, r.client.Set(ctx, key, value, expiration).Err()
	})
	return err
}

func (r *cacheRepository) Get(ctx context.Context, key string) (string, error) {
	res, err := infrastructure.RedisBreaker.Execute(func() (any, error) {
		return r.client.Get(ctx, key).Result()
	})
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", apperrors.ErrNotFound
		}
		return "", err
	}
	val, ok := res.(string)
	if !ok {
		return "", errors.New("type assertion to string failed")
	}
	return val, nil
}

func (r *cacheRepository) Delete(ctx context.Context, key string) error {
	_, err := infrastructure.RedisBreaker.Execute(func() (any, error) {
		return nil, r.client.Del(ctx, key).Err()
	})
	return err
}
