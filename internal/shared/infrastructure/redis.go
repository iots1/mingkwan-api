package infrastructure

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/iots1/mingkwan-api/internal/shared/utils"
)

type RedisClient struct {
	client   *redis.Client
	addr     string
	password string
	db       int
}

func NewRedisClient(addr, password string, db int) *RedisClient {
	return &RedisClient{
		addr:     addr,
		password: password,
		db:       db,
	}
}

func (rc *RedisClient) Connect(ctx context.Context) (*redis.Client, error) {
	rc.client = redis.NewClient(&redis.Options{
		Addr:     rc.addr,
		Password: rc.password,
		DB:       rc.db,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := rc.client.Ping(pingCtx)
	if status.Err() != nil {
		if closeErr := rc.client.Close(); closeErr != nil {
			utils.Logger.Error("Error closing Redis client after failed ping", zap.Error(closeErr))
		}
		return nil, fmt.Errorf("failed to ping Redis: %w", status.Err())
	}

	utils.Logger.Info("Successfully connected to Redis!")
	return rc.client, nil
}

func (rc *RedisClient) GetClient() *redis.Client {
	if rc.client == nil {
		utils.Logger.Fatal("Redis client is not connected. Call Connect() first.")
		return nil
	}
	return rc.client
}

func (rc *RedisClient) Disconnect() {
	if rc.client != nil {
		if err := rc.client.Close(); err != nil {
			utils.Logger.Error("Error disconnecting from Redis", zap.Error(err))
		} else {
			utils.Logger.Info("Disconnected from Redis.")
		}
	}
}
