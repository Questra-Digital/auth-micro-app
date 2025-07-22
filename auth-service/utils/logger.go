package utils

import (
	"auth-server/config"
	"auth-server/models"
	"log"
	"os"
	"time"
	"sync"
	"github.com/gin-gonic/gin"
)

type Logger struct {
	env string
}

var (
	instance *Logger
	once     sync.Once
)

// NewLogger returns a singleton logger instance
func NewLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			env: config.AppConfig.AppEnv,
		}
	})
	return instance
}

// Info logs informational messages
func (l *Logger) Info(format string, args ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Printf("[INFO] "+format, args...)
}

// Warn logs warning messages
func (l *Logger) Warn(format string, args ...interface{}) {
	log.SetOutput(os.Stdout)
	log.Printf("[WARN] "+format, args...)
}

// Debug logs debug messages, but skips in production
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.env == "production" {
		return // Suppress in production
	}
	log.SetOutput(os.Stdout)
	log.Printf("[DEBUG] "+format, args...)
}

// LogAuditRecord logs and stores an audit trail into the database
func (l *Logger) LogAuditRecord(record models.AuditRecord) {
	record.Timestamp = time.Now()

	if err := config.AuditDB.Create(&record).Error; err != nil {
		l.Warn("Failed to store audit record: %v", err)
	} else {
		l.Info("Audit log stored: user=%s, action=%s, status=%s", record.UserID, record.Action, record.Status)
	}
}

// LogRateLimit logs information about a rate-limited request without exposing sensitive data in production.
func (l *Logger) LogRateLimit(c *gin.Context, params models.RateLimitParams) {
	msg := "Rate limit reached: Method=%s, Endpoint=%s, Requests=%d/%d, Window=%s"
	if params.Blocked {
		msg += " [BLOCKED]"
	}

	if l.env == "production" {
		l.Warn(msg,
			params.Method,
			params.Endpoint,
			params.RequestCount,
			params.Limit,
			params.Window,
		)
	} else {
		ip := c.ClientIP()
		msg += ", IP=%s, RateLimitType=%s"
		l.Debug(msg,
			params.Method,
			params.Endpoint,
			params.RequestCount,
			params.Limit,
			params.Window,
			ip,
			params.RateLimit,
		)
	}
}
