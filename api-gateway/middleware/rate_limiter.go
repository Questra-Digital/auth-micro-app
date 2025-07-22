package middleware

import (
	"api-gateway/config"
	"api-gateway/redis"
	"api-gateway/utils"
	"fmt"
	"github.com/gin-gonic/gin"
	limiter "github.com/ulule/limiter/v3"
	redisStore "github.com/ulule/limiter/v3/drivers/store/redis"
	"log"
	"net/http"
	"api-gateway/models"
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

	// Convert integer to limiter format string (e.g., "3-M")
	rateLimitFormatted := fmt.Sprintf("%d-M", config.AppConfig.RateLimit)
	// Use it with limiter
	rate, err := limiter.NewRateFromFormatted(rateLimitFormatted)
	if err != nil {
		log.Fatal("Invalid rate format:", err)
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
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Rate limiter error"})
			return
		}

		// Log data
		params := models.RateLimitParams{
			Endpoint:     c.Request.URL.Path,
			Method:       c.Request.Method,
			RateLimit:    "IP_BASED",
			RequestCount: int(context.Limit - context.Remaining),
			Limit:        int(context.Limit),
			Window:       "1m",
			Blocked:      context.Reached,
		}

		// Check if limit is exceeded
		if context.Reached {
			logger.LogRateLimit(c, params)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			return
		}

		logger.LogRateLimit(c, params)

		// Allow request
		c.Next()
	}
}
