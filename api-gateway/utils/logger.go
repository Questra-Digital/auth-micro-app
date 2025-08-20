package utils

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
	"github.com/gin-gonic/gin"
	"api-gateway/config"
	"api-gateway/models"
)

// Logger struct encapsulates loggers
type Logger struct {
	infoLogger    *log.Logger
	errorLogger   *log.Logger
	warningLogger *log.Logger
	debugLogger   *log.Logger
	appEnv        string
}

// Singleton instance
var (
	instance *Logger
	once     sync.Once
)

// NewLogger returns a singleton Logger instance
func NewLogger() *Logger {
	once.Do(func() {
		instance = &Logger{
			infoLogger:    log.New(os.Stdout, "[INFO] ", log.LstdFlags),
			errorLogger:   log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
			warningLogger: log.New(os.Stdout, "[WARN] ", log.LstdFlags),
			debugLogger:   log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
			appEnv:        config.AppConfig.AppEnv,
		}
	})
	return instance
}

// Info logs general info
func (l *Logger) Info(msg string, args ...interface{}) {
	l.infoLogger.Println(format(msg, args...))
}

// Error logs error messages
func (l *Logger) Error(msg string, args ...interface{}) {
	l.errorLogger.Println(format(msg, args...))
}

// Warn logs warning messages
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.warningLogger.Println(format(msg, args...))
}

// Debug logs debug messages only in development mode
func (l *Logger) Debug(msg string, args ...interface{}) {
	if l.appEnv == "development" {
		l.debugLogger.Println(format(msg, args...))
	}
}

// format helper for string formatting
func format(msg string, args ...interface{}) string {
	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// LogAuditEntry logs an audit entry to the DB
func (l *Logger) LogAuditEntry(entry models.AuditLog) {
	db := config.DB
	if db == nil {
		l.Error("DB not initialized. Cannot log audit record.")
		return
	}

	if err := db.Create(&entry).Error; err != nil {
		l.Error("Failed to log audit event to DB: %v", err)
	} else {
		l.Debug("Audit event logged: Group=%s, Action=%s, Endpoint=%s", entry.EventGroup, entry.EventAction, entry.Endpoint)
	}
}

// NewAuditEntry creates a new models.AuditLog entry
func (l *Logger) NewAuditEntry(
	group models.EventGroup,
	action models.EventAction,
	userID, sessionID *string,
	r models.RequestContext,
	status int,
	msg *string,
) models.AuditLog {
	return models.AuditLog{
		EventGroup:  group,
		EventAction: action,
		UserID:      userID,
		SessionID:   sessionID,
		SourceIP:    r.IP,
		Endpoint:    r.Path,
		HTTPMethod:  r.Method,
		StatusCode:  status,
		ServiceName: "API-GATEWAY",
		Timestamp:   time.Now(),
		Message:     msg,
	}
}

// LogRateLimit logs information about a rate-limited request without exposing sensitive data in production.
func (l *Logger) LogRateLimit(c *gin.Context, params models.RateLimitParams) {
	msg := "Rate limit reached: Method=%s, Endpoint=%s, Requests=%d/%d, Window=%s"
	if params.Blocked {
		msg += " [BLOCKED]"
	}

	if l.appEnv == "production" {
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