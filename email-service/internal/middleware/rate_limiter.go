package middleware

import (
	"fmt"
	"email-service/internal/config"
	"email-service/internal/logger"
	"email-service/internal/redis"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
	"net/http"
)

// RedisLimiter holds the limiter instance
var RedisLimiter *limiter.Limiter

// InitRateLimiter initializes the Redis-backed rate limiter
func InitRateLimiter() error {
	logger.Info("Initializing rate limiter...")

	// Get a Redis client
	rdb := redis.GetClient()

	// Use ulule's Redis store adapter
	store, err := redisStore.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:   "rate_limiter",
		MaxRetry: 3,
	})
	if err != nil {
		logger.Error("Failed to create Redis store for rate limiter: %v", err)
		return err
	}

	// Define rate: (RateLimitPerSecond) requests per second (from config)
	rate, err := limiter.NewRateFromFormatted(fmt.Sprintf("%d-S", config.AppConfig.RateLimitPerSecond))
	if err != nil {
		logger.Error("Failed to create rate limit: %v", err)
		return err
	}

	// Create limiter instance
	RedisLimiter = limiter.New(store, rate)
	
	logger.Info("Rate limiter initialized successfully with %d requests per second", config.AppConfig.RateLimitPerSecond)
	return nil
}

// RateLimitMiddleware is a Gin middleware that enforces rate limiting per IP
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Get rate limiting context for this IP
		context, err := RedisLimiter.Get(c, ip)
		if err != nil {
			// Log rate limiter error
			logger.Error("Rate limiter error for IP %s: %v", ip, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			return
		}

		// Check if limit is exceeded
		if context.Reached {
			// Log blocked rate limit event
			logger.SecureInfo("Rate limit exceeded for IP %s: %d/%d requests used", 
				ip, int(context.Limit), int(context.Limit))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		// Log allowed rate limit event
		logger.SecureInfo("Rate limit check passed for IP %s: %d/%d requests remaining", 
			ip, int(context.Remaining), int(context.Limit))

		// Allow request
		c.Next()
	}
}
