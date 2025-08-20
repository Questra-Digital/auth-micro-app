package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"otp-service/config"
	"otp-service/redis"
	"otp-service/utils"
)

func VerifyOTPHandler(c *gin.Context) {
	logger := utils.NewLogger()

	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		logger.LogVerifyFailure(c, "", "", "", "Missing session_id in cookie", 0)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing session_id in cookie"})
		return
	}

	var req struct {
		OTP string `json:"otp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OTP == "" {
		logger.LogVerifyFailure(c, "", sessionID, "", "OTP is required", 0)
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP is required"})
		return
	}

	session, err := redis.GetSession(sessionID)
	if err != nil {
		logger.LogVerifyFailure(c, "", sessionID, "", "Session expired or invalid", 0)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired or invalid"})
		return
	}

	if !utils.CompareOTP(session.OTPHash, req.OTP) {
		if err := redis.IncrementReattempts(sessionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attempt"})
			return
		}
		if session.Attempts >= config.MaxAttempts {
			redis.DeleteSession(sessionID)
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			logger.LogVerifyBlocked(c, session.Email, sessionID, session.OTPHash, "Maximum verification attempts exceeded", session.Attempts)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum verification attempts exceeded"})
			return
		}
		logger.LogVerifyFailure(c, session.Email, sessionID, session.OTPHash, "Invalid OTP", session.Attempts)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid OTP"})
		return
	}

	redis.DeleteSession(sessionID)
	c.SetCookie("session_id", "", -1, "/", "", false, true)
	logger.LogVerifySuccess(c, session.Email, sessionID, session.OTPHash, session.Attempts)
	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully"})
}
