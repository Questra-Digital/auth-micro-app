package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"otp-service/config"
	"otp-service/handlers"
	"otp-service/middleware"
	"otp-service/redis"
	"otp-service/utils"
	"strconv"
	"github.com/gin-gonic/gin"
)

func main() {
	
	config.InitConfig()

	// Initialize database
	if err := config.InitDatabase(); err != nil {
		log.Fatal("Database init failed:", err)
	}

	redis.InitRedis()

	if err := middleware.InitRateLimiter(); err != nil {
		log.Fatal("Rate limiter init failed:", err)
	}

	utils.StartCleanupJob()

	r := gin.Default()

	// OTP endpoints
	r.POST("/otp/generate", middleware.RateLimitMiddleware(), handlers.GenerateOTPHandler)
	r.POST("/otp/verify", middleware.RateLimitMiddleware(), handlers.VerifyOTPHandler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(config.AppConfig.OtpServicePort),
		Handler: r,
	}

	// Start server in a goroutine
	go func() {
		log.Println("OTP service running on port", config.AppConfig.OtpServicePort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	// kill (no param) default sends syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need to add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Graceful shutdown
	shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited gracefully")
}

// shutdown performs graceful cleanup of all resources
func shutdown() {
	log.Println("Starting graceful shutdown...")

	// Stop the cleanup job
	utils.StopCleanupJob()

	// Close Redis connection
	if err := redis.CloseRedis(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	} else {
		log.Println("Redis connection closed successfully")
	}

	// Close database connection
	config.CloseDatabaseConnection()

	log.Println("All resources cleaned up successfully")
}
