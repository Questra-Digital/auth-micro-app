package middleware

import (
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
	"otp-service/redis"
	"otp-service/config"
	"otp-service/utils"
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
		logger := utils.NewLogger()

		// Get rate limiting context for this IP
		context, err := RedisLimiter.Get(c, ip)
		if err != nil {
			// Log rate limiter error
			logger.LogRateLimit(c, c.Request.URL.Path, c.Request.Method, "IP_BASED", 0, 0, "1m", false)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			return
		}

		// Check if limit is exceeded
		if context.Reached {
			// Log blocked rate limit event
			logger.LogRateLimit(c, c.Request.URL.Path, c.Request.Method, "IP_BASED", 
				int(context.Limit), int(context.Limit), "1m", true)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		// Log allowed rate limit event
		logger.LogRateLimit(c, c.Request.URL.Path, c.Request.Method, "IP_BASED", 
			int(context.Remaining), int(context.Limit), "1m", false)

		// Allow request
		c.Next()
	}
}
