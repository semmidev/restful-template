package cache

import (
	"context"
	"time"
)

// CacheRepository defines an interface for a generic key-value store.
type CacheRepository interface {
	// Set stores a value with an optional expiration. (0 means no expiration)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Get retrieves a value by key. Returns an error if not found.
	Get(ctx context.Context, key string) (string, error)

	// Delete removes a value by key.
	Delete(ctx context.Context, key string) error
}
