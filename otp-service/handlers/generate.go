package handlers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"otp-service/config"
	"otp-service/models"
	"otp-service/redis"
	"otp-service/utils"
	"time"
	"bytes"
	"encoding/json"
)

func GenerateOTPHandler(c *gin.Context) {
	logger := utils.NewLogger()

	// Parse request body
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		params := utils.OTPEventParams{
			Email:       req.Email,
			EventType:   models.EventTypeGenerate,
			EventStatus: models.EventStatusFailed,
			Msg:         "Valid email is required",
		}
		logger.LogOTPEvent(c, params)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid email is required"})
		return
	}
	email := req.Email

	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		params := utils.OTPEventParams{
			Email:       email,
			EventType:   models.EventTypeGenerate,
			EventStatus: models.EventStatusFailed,
			Msg:         "Missing X-Session-ID header",
		}
		logger.LogOTPEvent(c, params)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing X-Session-ID header"})
		return
	}

	var session *models.OTPSession
	params := utils.OTPEventParams{
		Email:     email,
		SessionID: sessionID,
		EventType: models.EventTypeGenerate,
	}

	sessionValue, err := redis.GetSession(sessionID)
	if err != nil {
		if err.Error() == "redis: nil" {
			session = &models.OTPSession{
				Email:     email,
				SessionID: sessionID,
				Resends:   0,
				Attempts:  0,
				CreatedAt: time.Now(),
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Redis error while retrieving session"})
			return
		}
	} else {
		session = &sessionValue
		session.Resends++
		params.Resends = session.Resends

		if session.Resends >= config.MaxResends {
			redis.DeleteSession(sessionID)
			params.EventType = models.EventTypeRateLimit
			params.EventStatus = models.EventStatusBlocked
			params.Msg = "Maximum resends exceeded"
			logger.LogOTPEvent(c, params)
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Maximum OTP resends exceeded"})
			return
		}

		params.EventType = models.EventTypeResend
		params.EventStatus = models.EventStatusSuccess
		params.Msg = "OTP resent successfully"
		params.OTPHash = session.OTPHash
		logger.LogOTPEvent(c, params)
	}

	// Generate and hash OTP
	otp := utils.GenerateSecureOTP(config.OTPLength)
	hashedOTP, err := utils.HashOTP(otp)
	if err != nil {
		params.EventStatus = models.EventStatusFailed
		params.Msg = "Failed to hash OTP"
		logger.LogOTPEvent(c, params)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash OTP"})
		return
	}

	if config.AppConfig.AppEnv != "production" {
		log.Printf("[DEVELOPMENT] Generated OTP for %s: %s", email, otp)
	}

	// Update session
	session.OTPHash = hashedOTP
	session.CreatedAt = time.Now()

	if err := redis.StoreSession(*session, config.OTPTTL); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store session in Redis"})
		return
	}

	// Send OTP to email-service
	emailServiceURL := config.AppConfig.EmailServiceUrl + "/send-otp"
	emailReqBody, err := json.Marshal(map[string]string{
		"email": email,
		"otp":   otp,
	})
	if err != nil {
		redis.DeleteSession(sessionID)
		params.EventStatus = models.EventStatusFailed
		params.Msg = "Failed to marshal email request"
		logger.LogOTPEvent(c, params)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send OTP"})
		return
	}

	resp, err := http.Post(emailServiceURL, "application/json", bytes.NewBuffer(emailReqBody))
	if err != nil || resp.StatusCode != http.StatusOK {
		redis.DeleteSession(sessionID)
		params.EventStatus = models.EventStatusFailed
		params.Msg = "Failed to deliver OTP via email service"
		logger.LogOTPEvent(c, params)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to deliver OTP"})
		return
	}

	params.EventType = models.EventTypeGenerate
	params.EventStatus = models.EventStatusSuccess
	params.Msg = "OTP generated and sent successfully"
	params.OTPHash = hashedOTP
	params.Resends = session.Resends
	logger.LogOTPEvent(c, params)

	c.JSON(http.StatusOK, gin.H{"success": "OTP sent successfully"})
}