package utils

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"log"
	"otp-service/config"
	"otp-service/models"
	"time"
	"sync"
)


// Logger provides centralized logging functionality
type Logger struct {
	db           *gorm.DB
	isProduction bool
}

var (
	loggerInstance *Logger
	once           sync.Once
)

// NewLogger returns a singleton Logger instance
func NewLogger() *Logger {
	once.Do(func() {
		loggerInstance = &Logger{
			db:           config.DB,
			isProduction: config.AppConfig.AppEnv == "production",
		}
	})
	return loggerInstance
}

type OTPEventParams struct {
	Email       string
	SessionID   string
	EventType   string
	EventStatus string
	OTPHash     string
	Msg         string
	Attempts    int
	Resends     int
}

type RateLimitParams struct {
	Endpoint     string
	Method       string
	RateLimit    string
	RequestCount int
	Limit        int
	Window       string
	Blocked      bool
}

// LogOTPEvent logs an OTP event to both console and database
func (l *Logger) LogOTPEvent(c *gin.Context, p OTPEventParams) {
	// Apply defaults if fields are missing
	if p.EventType == "" {
		p.EventType = "otp_event"
	}
	if p.EventStatus == "" {
		p.EventStatus = "unknown"
	}
	if p.SessionID == "" {
		p.SessionID = "NULL"
	}
	if p.Email == "" {
		p.Email = "NULL"
	}
	now := time.Now()

	event := &models.OTPEvent{
		Email:       p.Email,
		SessionID:   p.SessionID,
		EventType:   p.EventType,
		EventStatus: p.EventStatus,
		OTPHash:     p.OTPHash,
		Msg:         p.Msg,
		Attempts:    p.Attempts,
		Resends:     p.Resends,
		CreatedAt:   now,
	}

	// Logging to console
	if !l.isProduction {
		logMsg := fmt.Sprintf(
			"[INFO] [%s] %s - Session: %s, Email: %s, Status: %s",
			event.EventType, now.Format("2006-01-02 15:04:05"), event.SessionID, event.Email, event.EventStatus,
		)
		if event.Msg != "" {
			logMsg += fmt.Sprintf(", Error: %s", event.Msg)
		}
		if event.OTPHash != "" {
			logMsg += ", OTPHash: [REDACTED IN PRODUCTION]"
		}
		log.Println(logMsg)
	} else {
		log.Printf(
			"[INFO] [%s] %s - Status: %s",
			event.EventType, now.Format("2006-01-02 15:04:05"), event.EventStatus,
		)
	}

	// Async DB logging
	go func() {
		if err := l.logToDatabase(c, event); err != nil && !l.isProduction {
			log.Printf("[WARN] Failed to log to database: %v", err)
		}
	}()
}

func (l *Logger) LogRateLimit(c *gin.Context, params RateLimitParams) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	if l.isProduction {
		log.Printf("[INFO] [RATE_LIMIT] %s - %s %s, Blocked: %t",
			timestamp, params.Method, params.Endpoint, params.Blocked)
	} else {
		log.Printf("[INFO] [RATE_LIMIT] %s - IP: %s, UA: %s, %s %s, Type: %s, Count: %d/%d, Window: %s, Blocked: %t",
			timestamp,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
			params.Method,
			params.Endpoint,
			params.RateLimit,
			params.RequestCount,
			params.Limit,
			params.Window,
			params.Blocked,
		)
	}
}

// logToDatabase logs the event to PostgreSQL using GORM
func (l *Logger) logToDatabase(c *gin.Context, event *models.OTPEvent) error {
	if c != nil {
		event.IPAddress = c.ClientIP()
		event.UserAgent = c.GetHeader("User-Agent")
	}
	return l.db.Create(event).Error
}
