package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// CacheManager provides methods for interacting with a Redis cache.
type CacheManager struct {
	client *redis.Client
}

// NewCacheManager creates a new CacheManager instance.
func NewCacheManager(client *redis.Client) *CacheManager {
	return &CacheManager{client: client}
}

// Set stores a key-value pair in Redis with an expiration.
func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := cm.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set cache key '%s': %w", key, err)
	}
	return nil
}

// Get retrieves a value from Redis by key.
func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("cache key '%s' not found", key)
	}
	if err != nil {
		return "", fmt.Errorf("failed to get cache key '%s': %w", key, err)
	}
	return val, nil
}

// Delete removes a key from Redis.
func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	err := cm.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete cache key '%s': %w", key, err)
	}
	return nil
}

// You could add more methods like Incr, Decr, Expire, etc.
