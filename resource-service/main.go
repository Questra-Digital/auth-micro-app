package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"resource-service/config"
	"resource-service/handlers"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

// gracefulShutdown handles OS signals and cleans up resources
func gracefulShutdown(srv *http.Server, cleanupStop chan struct{}, cleanupDone chan struct{}) {
	log.Println("Shutting down Resource API server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	close(cleanupStop)
	<-cleanupDone
}

func main() {
	// Initialize config
	config.InitConfig()

	// Setup Gin router and register routes
	r := gin.Default()

	// Resource CRUD routes
	r.GET("/resources", handlers.GetAllResources)
	r.GET("/resources/:id", handlers.GetResource)
	r.POST("/resources", handlers.CreateResource)
	r.PUT("/resources/:id", handlers.UpdateResource)
	r.DELETE("/resources/:id", handlers.DeleteResource)

	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(config.AppConfig.ResourceServicePort),
		Handler: r,
	}

	// Listen for shutdown signals in a goroutine
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		gracefulShutdown(srv, make(chan struct{}), make(chan struct{}))
		os.Exit(0)
	}()

	log.Println("Resource API server running on :" + strconv.Itoa(config.AppConfig.ResourceServicePort))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}
