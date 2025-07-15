package middleware

import (
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
	"otp-service/redis"
	"otp-service/config"
	"net/http"
)

// RedisLimiter holds the limiter instance
var RedisLimiter *limiter.Limiter

// InitRateLimiter initializes the Redis-backed rate limiter
func InitRateLimiter() error {
	// Get a Redis client
	rdb := redis.GetClient()

	// Use ulule's Redis store adapter
	store, err := redisStore.NewStoreWithOptions(rdb, limiter.StoreOptions{
		Prefix:   "rate_limiter",
		MaxRetry: 3,
	})
	if err != nil {
		return err
	}

	// Define rate: 3 requests per minute
	rate, err := limiter.NewRateFromFormatted(config.RateLimitPerMinute)
	if err != nil {
		return err
	}

	// Create limiter instance
	RedisLimiter = limiter.New(store, rate)
	return nil
}

// RateLimitMiddleware is a Gin middleware that enforces rate limiting per IP
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		// Get rate limiting context for this IP
		context, err := RedisLimiter.Get(c, ip)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			return
		}

		// Check if limit is exceeded
		if context.Reached {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		// Allow request
		c.Next()
	}
}
