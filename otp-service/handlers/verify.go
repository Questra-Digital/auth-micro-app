package handlers

import (
	"net/http"
	"otp-service/config"
	"otp-service/redis"
	"otp-service/utils"

	"github.com/gin-gonic/gin"
)

func VerifyOTPHandler(c *gin.Context) {
	logger := utils.NewLogger()

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		logEventAndRespond(c, logger, "Missing X-Session-ID header", "verify", "failure", "", "", 0, 0, http.StatusBadRequest)
		return
	}

	var req struct {
		Email string `json:"email" binding:"required,email"`
		OTP   string `json:"otp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logEventAndRespond(c, logger, "Invalid request payload", "verify", "failure", "", "", 0, 0, http.StatusBadRequest)
		return
	}

	session, err := redis.GetSession(sessionID)
	if err != nil {
		logEventAndRespond(c, logger, "Session expired or invalid", "verify", "failure", "", "", 0, 0, http.StatusUnauthorized)
		return
	}

	if session.Email != req.Email {
		logEventAndRespond(c, logger, "Email does not match session", "verify", "failure", session.Email, session.OTPHash, session.Attempts, session.Resends, http.StatusUnauthorized)
		return
	}

	if !utils.CompareOTP(session.OTPHash, req.OTP) {
		if err := redis.IncrementReattempts(sessionID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record attempt"})
			return
		}

		session.Attempts++
		if session.Attempts >= config.MaxAttempts {
			redis.DeleteSession(sessionID)
			logEventAndRespond(c, logger, "Maximum verification attempts exceeded", "verify", "failure", session.Email, session.OTPHash, session.Attempts, session.Resends, http.StatusTooManyRequests)
			return
		}

		logEventAndRespond(c, logger, "Invalid OTP", "verify", "failure", session.Email, session.OTPHash, session.Attempts, session.Resends, http.StatusUnauthorized)
		return
	}

	redis.DeleteSession(sessionID)
	logEventAndRespond(c, logger, "OTP verified successfully", "verify", "success", session.Email, session.OTPHash, session.Attempts, session.Resends, http.StatusOK)
}

func logEventAndRespond(
	c *gin.Context,
	logger *utils.Logger,
	msg, eventType, status, email, otpHash string,
	attempts, resends int,
	httpStatus int,
) {
	logger.LogOTPEvent(c, utils.OTPEventParams{
		SessionID:   c.GetHeader("X-Session-ID"),
		EventType:   eventType,
		EventStatus: status,
		Email:       email,
		OTPHash:     otpHash,
		Attempts:    attempts,
		Resends:     resends,
		Msg:         msg,
	})
	if status == "success" {
		c.JSON(httpStatus, gin.H{"message": msg})
	} else {
		c.JSON(httpStatus, gin.H{"error": msg})
	}
}
