package handlers

import (
	"api-gateway/api"
	"api-gateway/models"
	"api-gateway/redis"
	"api-gateway/utils"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
)

func SignUpHandler(c *gin.Context) {
	log := utils.NewLogger()
	otpClient := api.NewOTPClient()

	// Extract request context info
	reqCtx := models.RequestContext{
		IP:     c.ClientIP(),
		Method: c.Request.Method,
		Path:   c.FullPath(),
	}

	// Parse email from body
	var body struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Email == "" {
		log.Warn("Missing email in request body")

		msg := "Missing email"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			models.ActionSignup,
			nil,
			nil,
			reqCtx,
			http.StatusBadRequest,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Validate email format
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(body.Email) {
		log.Warn("Invalid email format: %s", body.Email)

		msg := "Invalid email format"
		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Generate session ID
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Error("Failed to generate session ID: %v", err)

		msg := "Internal server error"
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	// Store session data in Redis (clientID as a field in the session hash)
	clientID := c.ClientIP()
	if err := redis.StoreSessionData(sessionID, clientID, "", body.Email, "", 15*time.Minute); err != nil {
		log.Error("Failed to store session data in Redis: %v", err)

		msg := "Internal server error"
		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	// Set cookie with standardized name
	cookie := &http.Cookie{
		Name:     "sessionId",
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int((15 * time.Minute).Seconds()), // this token will live only for 15 mins.
	}
	http.SetCookie(c.Writer, cookie)

	// Request OTP using OTP client
	resp, err := otpClient.RequestOTP(body.Email, sessionID)
	if err != nil {
		log.Error("Request to OTP service failed: %v", err)

		msg := "OTP service unreachable"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			models.ActionSignup,
			&clientID,
			&sessionID,
			reqCtx,
			http.StatusBadGateway,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)

	// Prepare audit entry
	msg := "Signup attempt"
	auditEntry := log.NewAuditEntry(
		models.EventGroupAuth,
		models.ActionSignup,
		&clientID,
		&sessionID,
		reqCtx,
		resp.StatusCode,
		&msg,
	)

	if resp.StatusCode == http.StatusOK {
		log.Info("OTP sent successfully to %s", body.Email)
		successMsg := "OTP sent successfully"
		auditEntry.Message = &successMsg
		log.LogAuditEntry(auditEntry)
		c.Data(http.StatusOK, "application/json", respBody)
	} else {
		log.Warn("OTP service responded with status %d", resp.StatusCode)
		errMsg := fmt.Sprintf("Failed to initiate signup: %s", string(respBody))
		auditEntry.Message = &errMsg
		log.LogAuditEntry(auditEntry)
		c.JSON(http.StatusBadGateway, gin.H{
			"error":  "Failed to initiate signup",
			"detail": string(respBody),
		})
	}
}
