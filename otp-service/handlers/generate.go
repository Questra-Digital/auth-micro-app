package handlers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"otp-service/config"
	"otp-service/models"
	"otp-service/redis"
	"otp-service/utils"
	"time"
	"log"
	"os"
)

func GenerateOTPHandler(c *gin.Context) {
	logger := utils.NewLogger()

	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.LogGenerateFailure(c, req.Email, "", "Valid email is required", 0)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
		return
	}
	email := req.Email

	// Get or generate session ID
	sessionID, err := c.Cookie("session_id")
	var session *models.OTPSession
	if err != nil || sessionID == "" {
		sessionID = utils.GenerateSessionID()
		c.SetCookie("session_id", sessionID, int(config.OTPTTL.Seconds()), "/", "", false, true)
		session = &models.OTPSession{
			Email:     email,
			SessionID: sessionID,
			Resends:   0,
			Attempts:  0,
			CreatedAt: time.Now(),
		}
	} else {
		sessionValue, err := redis.GetSession(sessionID)
		if err != nil {
			if err.Error() == "redis: nil" {
				logger.LogGenerateFailure(c, email, sessionID, "Session against this sessionId does not exist", 0)
				c.JSON(http.StatusBadRequest, gin.H{"error": "Session against this sessionId does not exist"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error while retrieving session"})
			return
		}
		session = &sessionValue
		session.Resends++
		if session.Resends >= config.MaxResends {
			redis.DeleteSession(sessionID)
			c.SetCookie("session_id", "", -1, "/", "", false, true)
			logger.LogGenerateBlocked(c, email, sessionID, "Maximum resends exceeded", session.Resends)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum OTP resends exceeded"})
			return
		}
		// Log resend event
		logger.LogGenerateResend(c, email, sessionID, session.OTPHash, session.Resends)
	}

	// Generate and hash OTP
	otp := utils.GenerateSecureOTP(config.OTPLength)
	hashedOTP, err := utils.HashOTP(otp)
	if err != nil {
		logger.LogGenerateFailure(c, email, sessionID, "Failed to hash OTP", session.Resends)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash OTP"})
		return
	}

	// Log OTP value in development mode only
	if os.Getenv("APP_ENV") != "production" {
		log.Printf("[DEVELOPMENT] Generated OTP for %s: %s", email, otp)
	}

	// Update session data
	session.SessionID = sessionID
	session.OTPHash = hashedOTP
	session.CreatedAt = time.Now()

	if err := redis.StoreSession(*session, config.OTPTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store session in Redis"})
		return
	}

	logger.LogGenerateSuccess(c, email, sessionID, hashedOTP, session.Resends)
	c.JSON(http.StatusOK, gin.H{"success": "OTP generated successfully"})
}
