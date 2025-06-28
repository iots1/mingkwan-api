package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
)

type CacheManager struct {
	client *redis.Client
}

func NewCacheManager(client *redis.Client) *CacheManager {
	return &CacheManager{client: client}
}

func (cm *CacheManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := cm.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		utils.Logger.Error(
			"CacheManager: Failed to set cache key",
			zap.String("key", key),
			zap.Any("value", value),
			zap.Duration("expiration", expiration),
			zap.Error(err),
		)
		return fmt.Errorf("failed to set cache key '%s': %w", key, err) // ยังคงคืนค่า error กลับไป
	}
	utils.Logger.Debug("CacheManager: Cache key set successfully", zap.String("key", key))
	return nil
}

func (cm *CacheManager) Get(ctx context.Context, key string) (string, error) {
	val, err := cm.client.Get(ctx, key).Result()
	if err == redis.Nil {
		utils.Logger.Info("CacheManager: Cache key not found", zap.String("key", key))
		return "", errors.New("cache key not found") // สามารถใช้ errors.New แทน fmt.Errorf ได้ถ้าไม่ต้องการ format
	}
	if err != nil {
		utils.Logger.Error(
			"CacheManager: Failed to get cache key",
			zap.String("key", key),
			zap.Error(err),
		)
		return "", fmt.Errorf("failed to get cache key '%s': %w", key, err) // ยังคงคืนค่า error กลับไป
	}
	utils.Logger.Debug("CacheManager: Cache key retrieved successfully", zap.String("key", key))
	return val, nil
}

func (cm *CacheManager) Delete(ctx context.Context, key string) error {
	err := cm.client.Del(ctx, key).Err()
	if err != nil {
		utils.Logger.Error(
			"CacheManager: Failed to delete cache key",
			zap.String("key", key),
			zap.Error(err),
		)
		return fmt.Errorf("failed to delete cache key '%s': %w", key, err)
	}
	utils.Logger.Debug("CacheManager: Cache key deleted successfully", zap.String("key", key))
	return nil
}
