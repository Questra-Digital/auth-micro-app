package logger

import (
	"log"
	"os"
	"time"

	"email-service/internal/config"
	"email-service/internal/models"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugMode   bool
)

func InitLogger(isDebug bool) {
	debugMode = isDebug

	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
}

// Info logs general information
func Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		infoLogger.Printf(msg, args...)
	} else {
		infoLogger.Println(msg)
	}
}

// Error logs errors
func Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		errorLogger.Printf(msg, args...)
	} else {
		errorLogger.Println(msg)
	}
}

// SecureInfo avoids logging sensitive data in production
func SecureInfo(msg string, args ...interface{}) {
	if debugMode {
		Info(msg, args...)
	} 
}

// LogEmailAudit logs email metadata to PostgreSQL for traceability and analytics
func LogEmailAudit(recipient, status string) {
	audit := models.EmailAudit{
		From:   config.AppConfig.SMTPUsername,
		Recipient: recipient,
		Status:    status,
		Timestamp: time.Now(),
	}

	if err := config.DB.Create(&audit).Error; err != nil {
		Error("Failed to log email audit: %v", err)
		return
	}

	Info("Email audit logged: ID=%s, Recipient=%s, Status=%s", config.AppConfig.SMTPUsername, recipient, status)
}
