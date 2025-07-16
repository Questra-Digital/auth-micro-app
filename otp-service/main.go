package main
import (
	"github.com/gin-gonic/gin"
	"otp-service/handlers"
	"otp-service/middleware"
	"otp-service/redis"
	"otp-service/config"
	"otp-service/utils"
	"log"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if present
	_ = godotenv.Load()

	redis.InitRedis()

	if err := middleware.InitRateLimiter(); err != nil {
		log.Fatal("Rate limiter init failed:", err)
	}

	// Initialize database
	if err := config.InitDatabase(); err != nil {
		log.Fatal("Database init failed:", err)
	}

	// Create database tables
	if err := config.CreateTables(); err != nil {
		log.Fatal("Database table creation failed:", err)
	}

	// Start cleanup service for automatic record expiration
	cleanupService := utils.NewCleanupService()
	cleanupService.StartCleanupJob()
	log.Println("Cleanup service started - records will expire automatically")
	
	r := gin.Default()

	// OTP endpoints
	r.POST("/otp/generate", middleware.RateLimitMiddleware(), handlers.GenerateOTPHandler)
	r.POST("/otp/verify", middleware.RateLimitMiddleware(), handlers.VerifyOTPHandler)

	log.Println("OTP service running on :8080")
	r.Run(":8080")
}
