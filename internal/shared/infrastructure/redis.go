// internal/shared/infrastructure/redis.go
package infrastructure

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9" // Import the Redis client library
)

// RedisClient wraps the Redis client and provides connection management
type RedisClient struct {
	client   *redis.Client
	addr     string
	password string
	db       int
}

// NewRedisClient creates a new RedisClient instance.
// It doesn't establish the connection yet, only sets up the configuration.
func NewRedisClient(addr, password string, db int) *RedisClient {
	return &RedisClient{
		addr:     addr,
		password: password,
		db:       db,
	}
}

// Connect establishes a connection to the Redis server.
// It returns the *redis.Client instance or an error.
func (rc *RedisClient) Connect(ctx context.Context) (*redis.Client, error) {
	rc.client = redis.NewClient(&redis.Options{
		Addr:     rc.addr,
		Password: rc.password,
		DB:       rc.db,
		// PoolSize: 10, // You can configure connection pool size here
	})

	// Ping the Redis server to verify connection
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	status := rc.client.Ping(pingCtx)
	if status.Err() != nil {
		// Close client if ping fails
		if closeErr := rc.client.Close(); closeErr != nil {
			log.Printf("Error closing Redis client after failed ping: %v", closeErr)
		}
		return nil, fmt.Errorf("failed to ping Redis: %w", status.Err())
	}

	log.Println("Successfully connected to Redis!")
	return rc.client, nil
}

// GetClient returns the underlying *redis.Client.
// Call this *after* a successful Connect().
func (rc *RedisClient) GetClient() *redis.Client {
	if rc.client == nil {
		log.Fatal("Redis client is not connected. Call Connect() first.")
		return nil // Or return an error
	}
	return rc.client
}

// Disconnect closes the Redis connection.
func (rc *RedisClient) Disconnect() {
	if rc.client != nil {
		if err := rc.client.Close(); err != nil {
			log.Printf("Error disconnecting from Redis: %v", err)
		} else {
			log.Println("Disconnected from Redis.")
		}
	}
}
