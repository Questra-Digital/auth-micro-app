package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email-service/internal/api"
	"email-service/internal/cleanup"
	"email-service/internal/config"
	"email-service/internal/logger"
	"email-service/internal/middleware"
	"email-service/internal/models"
	"email-service/internal/queue"
	"email-service/internal/redis"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration and dependencies
	config.LoadConfig()
	logger.InitLogger(config.AppConfig.APP_MODE == "development")
	
	// Initialize database
	if err := config.InitDatabase(); err != nil {
		logger.Error("Failed to initialize database: %v", err)
		os.Exit(1)
	}
	
	// Initialize Redis
	if err := redis.InitRedisClient(); err != nil {
		logger.Error("Failed to initialize Redis: %v", err)
		os.Exit(1)
	}
	
	// Initialize rate limiter
	if err := middleware.InitRateLimiter(); err != nil {
		logger.Error("Failed to initialize rate limiter: %v", err)
		os.Exit(1)
	}
	
	// Initialize RabbitMQ
	if err := queue.InitRabbitMQ(); err != nil {
		logger.Error("Failed to initialize RabbitMQ: %v", err)
		os.Exit(1)
	}
	
	// Start email consumer
	if err := queue.StartEmailConsumer(); err != nil {
		logger.Error("Failed to start email consumer: %v", err)
		os.Exit(1)
	}
	
	cleanup.StartAuditPurger()

	// Start HTTP server
	srv := startServer()

	// Handle graceful shutdown
	gracefulShutdown(srv)
}

// startServer configures and starts the HTTP server
func startServer() *http.Server {
	router := gin.Default()
	
	// Add rate limiting middleware
	router.Use(middleware.RateLimitMiddleware())
	
	router.POST("/send-otp", api.SendOTPHandler)

	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	logger.Info("HTTP server started on port %s", config.AppConfig.Port)
	return srv
}

// gracefulShutdown handles OS signals and cleans up resources
func gracefulShutdown(srv *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	queue.StopEmailConsumer()

	if models.EmailChannel != nil {
		if err := models.EmailChannel.Close(); err != nil {
			logger.Error("Error closing RabbitMQ channel: %v", err)
		}
		logger.Info("RabbitMQ channel closed")
	}

	// Close Redis connection
	redis.CloseRedisClient()

	// Close DB connection
	if sqlDB, err := config.DB.DB(); err != nil {
		logger.Error("Failed to retrieve raw DB from GORM: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Failed to close database connection: %v", err)
		} else {
			logger.Info("Database connection closed")
		}
	}

	logger.Info("Server exited gracefully")
}