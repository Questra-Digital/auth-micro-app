package main

import (
	"auth-server/config"
	"auth-server/handlers"
	"auth-server/middleware"
	"auth-server/redis"
	"auth-server/utils"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize App Config
	config.InitConfig()

	// Initialize Logger
	logger := utils.NewLogger()
	logger.Info("Starting Authorization Server...")

	// Initialize PostgreSQL
	if err := config.InitDatabase(); err != nil {
		logger.Warn("Database initialization failed: %v", err)
		os.Exit(1)
	}

	// Initialize Redis
	redis.InitRedis()

	// Initialize Rate Limiter
	if err := middleware.InitRateLimiter(); err != nil {
		logger.Warn("Failed to initialize rate limiter: %v", err)
		os.Exit(1)
	}

	// Start cleanup job
	stopCleanup := make(chan struct{})
	go utils.StartCleanup(stopCleanup)

	// Setup Gin
	r := gin.Default()

	// Apply middleware
	r.Use(middleware.RateLimitMiddleware())

	// Routes placeholder
	r.POST("/getAccessToken", handlers.GetAccessToken)
	r.POST("/refreshToken", handlers.RefreshAccessToken)

	// Start server in a goroutine
	go func() {
		port := ":" + strconv.Itoa(config.AppConfig.AppPort)
		logger.Info("Server running on %s", port)
		if err := r.Run(port); err != nil {
			logger.Warn("Gin server failed to start: %v", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown on Ctrl+C
	gracefulShutdown(stopCleanup)
}

func gracefulShutdown(stopCleanup chan struct{}) {
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop cleanup routine
	close(stopCleanup)

	// Cleanup
	config.CloseDatabaseConnection()
	if err := redis.CloseRedis(); err != nil {
		log.Printf("Error closing Redis: %v", err)
	}

	log.Println("Server shutdown complete.")
}
