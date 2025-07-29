package handlers

import (
	"api-gateway/config"
	"api-gateway/models"
	"api-gateway/redis"
	"api-gateway/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"regexp"
	"time"
)

func SignUpHandler(c *gin.Context) {
	log := utils.NewLogger()

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
		log.Warn("Invalid or missing email in request body")

		msg := "Invalid or missing email"
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

	// Generate session ID
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Error("Failed to generate session ID: %v", err)

		msg := "Internal server error"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			models.ActionSignup,
			nil,
			nil,
			reqCtx,
			http.StatusInternalServerError,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	// Store session data in Redis (clientID as a field in the session hash)
	clientID := c.ClientIP()
	if err := redis.StoreSessionData(sessionID, clientID, "", body.Email,15*time.Minute); err != nil {
		log.Error("Failed to store session data in Redis: %v", err)

		msg := "Internal server error"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			models.ActionSignup,
			nil,
			&sessionID,
			reqCtx,
			http.StatusInternalServerError,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

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
		MaxAge:   int((15 * time.Minute).Seconds()),// this token will live only for 15 mins.
	}
	http.SetCookie(c.Writer, cookie)

	// Prepare OTP request
	otpPayload := map[string]string{"email": body.Email}
	otpBody, _ := json.Marshal(otpPayload)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/otp/generate", config.AppConfig.OtpService), bytes.NewBuffer(otpBody))
	if err != nil {
		log.Error("Failed to create request to OTP service: %v", err)

		msg := "Internal server error"
		auditEntry := log.NewAuditEntry(
			models.EventGroupAuth,
			models.ActionSignup,
			nil,
			&sessionID,
			reqCtx,
			http.StatusInternalServerError,
			&msg,
		)
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)

	// Call OTP service
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	msg := "Signup attempt"
	auditEntry := log.NewAuditEntry(
		models.EventGroupAuth,
		models.ActionSignup,
		&clientID,
		&sessionID,
		reqCtx,
		0,
		&msg,
	)

	if err != nil {
		log.Error("Request to OTP service failed: %v", err)
		auditEntry.StatusCode = http.StatusBadGateway
		failMsg := "OTP service unreachable"
		auditEntry.Message = &failMsg
		log.LogAuditEntry(auditEntry)

		c.JSON(http.StatusBadGateway, gin.H{"error": failMsg})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	auditEntry.StatusCode = resp.StatusCode

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
