package main
import (
	"github.com/gin-gonic/gin"
	"otp-service/handlers"
	"otp-service/middleware"
	"otp-service/redis"
	"log"
)

func main() {

	redis.InitRedis()

	if err := middleware.InitRateLimiter(); err != nil {
		log.Fatal("Rate limiter init failed:", err)
	}
	
	r := gin.Default()

	r.POST("/otp/generate", middleware.RateLimitMiddleware(), handlers.GenerateOTPHandler)
	r.POST("/otp/verify", middleware.RateLimitMiddleware(), handlers.VerifyOTPHandler)

	log.Println("OTP service running on :8080")
	r.Run(":8080")
}
