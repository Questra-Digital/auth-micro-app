package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strconv"
	"api-gateway/config"
	"api-gateway/redis"
	"api-gateway/utils"
	"api-gateway/handlers"
	"api-gateway/middleware"
	"github.com/gin-gonic/gin"
)

// gracefulShutdown handles OS signals and cleans up resources
func gracefulShutdown(srv *http.Server, cleanupStop chan struct{}, cleanupDone chan struct{}) {
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	close(cleanupStop)
	<-cleanupDone
	config.CloseDatabaseConnection()
	redis.CloseRedis()
}

func main() {
	// Initialize config
	config.InitConfig()

	// Initialize database
	if err := config.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Redis
	redis.InitRedis()

	// Initialize Rate Limiter
	if err := middleware.InitRateLimiter(); err != nil {
		log.Fatalf("Failed to initialize rate limiter: %v", err)
	}

	// Start cleanup cron
	cleanupStop := make(chan struct{})
	cleanupDone := make(chan struct{})
	go func() {
		utils.StartCleanup(cleanupStop)
		close(cleanupDone)
	}()

	// Setup Gin router and register routes
	r := gin.Default()
	r.Use(middleware.RateLimitMiddleware())

	r.POST("/signup", handlers.SignUpHandler)
	r.POST("/verify-otp", handlers.VerifyOTPHandler)

	srv := &http.Server{
		Addr:    ":"+strconv.Itoa(config.AppConfig.ApiGatewayPort),
		Handler: r,
	}

	// Listen for shutdown signals in a goroutine
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		gracefulShutdown(srv, cleanupStop, cleanupDone)
		os.Exit(0)
	}()

	log.Println("Server running on :"+strconv.Itoa(config.AppConfig.ApiGatewayPort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}