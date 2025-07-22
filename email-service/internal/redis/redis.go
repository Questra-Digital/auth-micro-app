package redis

import (
	"context"
	"email-service/internal/config"
	"email-service/internal/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

// GetClient returns the singleton Redis client instance
func GetClient() *redis.Client {
	if redisClient == nil {
		InitRedisClient()
	}
	return redisClient
}

// InitRedisClient initializes the Redis client with configuration
func InitRedisClient() error {
	cfg := config.AppConfig

	redisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       0,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis: %v", err)
		return err
	}

	logger.Info("Redis client initialized successfully")
	return nil
}

// CloseRedisClient closes the Redis client connection
func CloseRedisClient() {
	if redisClient != nil {
		if err := redisClient.Close(); err != nil {
			logger.Error("Error closing Redis client: %v", err)
		} else {
			logger.Info("Redis client closed successfully")
		}
	}
} 