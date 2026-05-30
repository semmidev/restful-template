package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-co-op/gocron/v2"
	redisclient "github.com/redis/go-redis/v9"
)

// ErrFailedToAcquireLock is returned when the locker fails to acquire a lock via Redis
var ErrFailedToAcquireLock = errors.New("failed to acquire lock")

// RedisLocker implements gocron.Locker using Redis SetNX
type RedisLocker struct {
	client *redisclient.Client
	ttl    time.Duration
}

// NewRedisLocker creates a new RedisLocker for gocron.
// The ttl is the expiration time of the lock to prevent deadlocks if the instance crashes.
func NewRedisLocker(client *redisclient.Client, ttl time.Duration) *RedisLocker {
	return &RedisLocker{
		client: client,
		ttl:    ttl,
	}
}

// Lock attempts to acquire a lock for the given key.
func (l *RedisLocker) Lock(ctx context.Context, key string) (gocron.Lock, error) {
	lockKey := fmt.Sprintf("gocron:lock:%s", key)
	
	// Use SetNX to acquire the lock only if it doesn't already exist
	acquired, err := l.client.SetNX(ctx, lockKey, "locked", l.ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error while acquiring lock: %w", err)
	}

	if !acquired {
		return nil, ErrFailedToAcquireLock
	}

	return &RedisLock{
		client: l.client,
		key:    lockKey,
	}, nil
}

// RedisLock implements gocron.Lock
type RedisLock struct {
	client *redisclient.Client
	key    string
}

// Unlock releases the lock
func (l *RedisLock) Unlock(ctx context.Context) error {
	err := l.client.Del(ctx, l.key).Err()
	if err != nil {
		return fmt.Errorf("failed to release lock: %w", err)
	}
	return nil
}
