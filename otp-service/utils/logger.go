package utils

import (
	"fmt"
	"log"
	"otp-service/config"
	"otp-service/models"
	"time"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Logger provides centralized logging functionality
type Logger struct {
	db *gorm.DB
	isProduction bool
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	return &Logger{
		db: config.DB,
		isProduction: os.Getenv("APP_ENV") == "production",
	}
}

// LogGenerateSuccess logs a successful OTP generation
func (l *Logger) LogGenerateSuccess(c *gin.Context, email, sessionID, otpHash string, resends int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeGenerate,
		EventStatus: models.EventStatusSuccess,
		Email:       email,
		SessionID:   sessionID,
		OTPHash:     otpHash,
		Resends:     resends,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogGenerateFailure logs a failed OTP generation
func (l *Logger) LogGenerateFailure(c *gin.Context, email, sessionID, errorMsg string, resends int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeGenerate,
		EventStatus: models.EventStatusFailed,
		Email:       email,
		SessionID:   sessionID,
		ErrorMsg:    errorMsg,
		Resends:     resends,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogGenerateBlocked logs when OTP generation is blocked (e.g., max resends exceeded)
func (l *Logger) LogGenerateBlocked(c *gin.Context, email, sessionID, errorMsg string, resends int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeGenerate,
		EventStatus: models.EventStatusBlocked,
		Email:       email,
		SessionID:   sessionID,
		ErrorMsg:    errorMsg,
		Resends:     resends,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogGenerateResend logs when an OTP is resent
func (l *Logger) LogGenerateResend(c *gin.Context, email, sessionID, otpHash string, resends int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeGenerate,
		EventStatus: models.EventStatusSuccess,
		Email:       email,
		SessionID:   sessionID,
		OTPHash:     otpHash,
		Resends:     resends,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogVerifySuccess logs a successful OTP verification
func (l *Logger) LogVerifySuccess(c *gin.Context, email, sessionID, otpHash string, attempts int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeVerify,
		EventStatus: models.EventStatusSuccess,
		Email:       email,
		SessionID:   sessionID,
		OTPHash:     otpHash,
		Attempts:    attempts,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogVerifyFailure logs a failed OTP verification
func (l *Logger) LogVerifyFailure(c *gin.Context, email, sessionID, otpHash, errorMsg string, attempts int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeVerify,
		EventStatus: models.EventStatusFailed,
		Email:       email,
		SessionID:   sessionID,
		OTPHash:     otpHash,
		ErrorMsg:    errorMsg,
		Attempts:    attempts,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogVerifyBlocked logs when OTP verification is blocked (e.g., max attempts exceeded)
func (l *Logger) LogVerifyBlocked(c *gin.Context, email, sessionID, otpHash, errorMsg string, attempts int) {
	event := &models.OTPEvent{
		EventType:   models.EventTypeVerify,
		EventStatus: models.EventStatusBlocked,
		Email:       email,
		SessionID:   sessionID,
		OTPHash:     otpHash,
		ErrorMsg:    errorMsg,
		Attempts:    attempts,
		ExpiresAt:   time.Now().Add(GetRetentionPeriod()),
		CreatedAt:   time.Now(),
	}
	l.logEvent(c, event)
}

// LogRateLimit logs a rate limiting event
func (l *Logger) LogRateLimit(c *gin.Context, endpoint, method, rateLimitType string, requestCount, limit int, window string, blocked bool) {
	event := &models.RateLimitEvent{
		IPAddress:     c.ClientIP(),
		UserAgent:     c.GetHeader("User-Agent"),
		Endpoint:      endpoint,
		Method:        method,
		RateLimit:     rateLimitType,
		RequestCount:  requestCount,
		Limit:         limit,
		Window:        window,
		Blocked:       blocked,
		ExpiresAt:     time.Now().Add(GetRateLimitRetentionPeriod()),
		CreatedAt:     time.Now(),
	}
	l.logRateLimitEvent(c, event)
}

// logRateLimitEvent logs a rate limit event to both console and database
func (l *Logger) logRateLimitEvent(c *gin.Context, event *models.RateLimitEvent) {
	if !l.isProduction {
		// Development: log rate limit details
		logMsg := fmt.Sprintf("[RATE_LIMIT] %s - IP: %s, Endpoint: %s %s, Count: %d/%d, Blocked: %t", 
			time.Now().Format("2006-01-02 15:04:05"), event.IPAddress, event.Method, event.Endpoint, 
			event.RequestCount, event.Limit, event.Blocked)
		log.Println(logMsg)
	} else {
		// Production: log minimal rate limit info
		logMsg := fmt.Sprintf("[RATE_LIMIT] %s - Endpoint: %s %s, Blocked: %t", 
			time.Now().Format("2006-01-02 15:04:05"), event.Method, event.Endpoint, event.Blocked)
		log.Println(logMsg)
	}

	// Database logging (async to avoid blocking the request)
	go func() {
		if err := l.logRateLimitToDatabase(c, event); err != nil {
			if !l.isProduction {
				log.Printf("Failed to log rate limit to database: %v", err)
			}
		}
	}()
}

// logRateLimitToDatabase logs the rate limit event to PostgreSQL using GORM
func (l *Logger) logRateLimitToDatabase(c *gin.Context, event *models.RateLimitEvent) error {
	return l.db.Create(event).Error
}

// logEvent logs an event to both console and database
func (l *Logger) logEvent(c *gin.Context, event *models.OTPEvent) {
	if !l.isProduction {
		// Development: log everything
		logMsg := fmt.Sprintf("[%s] %s - Session: %s, Email: %s, Status: %s", 
			event.EventType, time.Now().Format("2006-01-02 15:04:05"), event.SessionID, event.Email, event.EventStatus)
		if event.ErrorMsg != "" {
			logMsg += fmt.Sprintf(", Error: %s", event.ErrorMsg)
		}
		if event.OTPHash != "" {
			logMsg += ", OTPHash: [REDACTED IN PRODUCTION]"
		}
		log.Println(logMsg)
	} else {
		// Production: log only minimal info, no sensitive fields
		logMsg := fmt.Sprintf("[%s] %s - Status: %s", 
			event.EventType, time.Now().Format("2006-01-02 15:04:05"), event.EventStatus)
		log.Println(logMsg)
	}

	// Database logging (async to avoid blocking the request)
	go func() {
		if err := l.logToDatabase(c, event); err != nil {
			if !l.isProduction {
				log.Printf("Failed to log to database: %v", err)
			}
		}
	}()
}

// logToDatabase logs the event to PostgreSQL using GORM
func (l *Logger) logToDatabase(c *gin.Context, event *models.OTPEvent) error {
	if c != nil {
		event.IPAddress = c.ClientIP()
		event.UserAgent = c.GetHeader("User-Agent")
	}
	return l.db.Create(event).Error
} 