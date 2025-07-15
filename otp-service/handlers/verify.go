package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"otp-service/config"
	"otp-service/redis"
	"otp-service/utils"
)

func VerifyOTPHandler(c *gin.Context) {
	// Retrieve session ID from cookie
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id in cookie"})
		return
	}

	// Bind OTP from request JSON
	var req struct {
		OTP string `json:"otp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OTP == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP is required"})
		return
	}

	// Retrieve session from Redis
	session, err := redis.GetSession(sessionID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
		return
	}

	// Check if max attempts reached
	if session.Attempts >= config.MaxAttempts {
		redis.DeleteSession(sessionID)
		c.SetCookie("session_id", "", -1, "/", "", false, true)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum verification attempts exceeded"})
		return
	}

	// Verify OTP
	if !utils.CompareOTP(session.OTPHash, req.OTP) {
		// Increment attempts safely
		if err := redis.IncrementReattempts(sessionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attempt"})
			return
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	// OTP is correct, cleanup session
	redis.DeleteSession(sessionID)
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	log.Printf("OTP verified for session: %s", sessionID)
	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}
