package api

import (
	"email-service/internal/logger"
	"email-service/internal/mailer"
	"email-service/internal/models"
	"email-service/internal/queue"
	"github.com/gin-gonic/gin"
	"net/http"
	"net/mail"
)

// isEmailValid returns true if addr is a syntactically valid email.
func isEmailValid(addr string) bool {
    _, err := mail.ParseAddress(addr)
    return err == nil
}

func SendOTPHandler(c *gin.Context) {
	var req models.OTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: ensure valid email and 6-digit OTP"})
		logger.Error("Invalid OTP request: %v", err)
		return
	}

	// Validate email format
    if !isEmailValid(req.Email) {
        logger.Error("Invalid email format: %s", req.Email)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email address"})
        return
    }

	// Log email attempt
	logger.LogEmailAudit(req.Email, "attempted")

	// Render HTML with provided OTP
	htmlBody, err := mailer.ParseOTPTemplate(req.OTP)
	if err != nil {
		logger.Error("Template error: %v", err)
		logger.LogEmailAudit(req.Email, "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to render email"})
		return
	}

	// Prepare and publish job
	job := models.EmailJob{
		To:       req.Email,
		Subject:  "Your One-Time Password (OTP)",
		HTMLBody: htmlBody,
	}
	if err := queue.PublishEmailJob(job); err != nil {
		logger.Error("Failed to queue email job: %v", err)
		logger.LogEmailAudit(req.Email, "failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue email"})
		return
	}

	logger.SecureInfo("OTP email job queued for: %s", req.Email)
	logger.LogEmailAudit(req.Email, "queued")
	c.JSON(http.StatusOK, gin.H{"message": "OTP email queued successfully"})
}
