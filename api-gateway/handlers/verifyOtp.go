package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"api-gateway/config"
	"api-gateway/models"
	"api-gateway/redis"
	"api-gateway/utils"
	"github.com/gin-gonic/gin"
)

type otpVerifyRequest struct {
	OTP   string `json:"otp"`
	Email string `json:"email"`
}

func VerifyOTPHandler(c *gin.Context) {
	log := utils.NewLogger()

	// Extract request context info
	reqCtx := models.RequestContext{
		IP:     c.ClientIP(),
		Method: c.Request.Method,
		Path:   c.FullPath(),
	}

	var userID *string
	var sessionID string

	// Prepare audit log entry base
	audit := log.NewAuditEntry(
		models.EventGroupAuth,
		models.ActionOTPVerified,
		userID,
		nil,
		reqCtx,
		http.StatusOK,
		nil,
	)

	// Step 1: Get sessionId from cookie
	var err error
	sessionID, err = c.Cookie("sessionId")
	if err != nil || sessionID == "" {
		log.Warn("Missing or invalid sessionId cookie")

		msg := "Missing sessionId cookie"
		audit.SessionID = nil
		audit.StatusCode = http.StatusBadRequest
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}
	audit.SessionID = &sessionID

	// Step 2: Identify client
	clientID := c.ClientIP()

	// Step 3: Validate session from Redis
	sessionData, err := redis.GetSessionData(sessionID)
	if err != nil || len(sessionData) == 0 {
		log.Warn("Invalid sessionId: %s", sessionID)

		msg := "Invalid session"
		audit.StatusCode = http.StatusUnauthorized
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}
	if sessionData["clientID"] != clientID {
		log.Warn("Session clientID mismatch: got %s, expected %s", sessionData["clientID"], clientID)

		msg := "Session does not belong to this client"
		audit.StatusCode = http.StatusUnauthorized
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
		return
	}

	email := sessionData["email"]
	if email == "" {
		log.Warn("Missing email in session data")

		msg := "Email not found in session"
		audit.StatusCode = http.StatusInternalServerError
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}


	// Step 4: Parse OTP from request
	var otpReq otpVerifyRequest
	if err := c.ShouldBindJSON(&otpReq); err != nil || otpReq.OTP == "" {
		log.Warn("Invalid OTP payload")

		msg := "Invalid OTP payload"
		audit.StatusCode = http.StatusBadRequest
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusBadRequest, gin.H{"error": msg})
		return
	}

	// Step 5: Send OTP verify request to OTP service
	otpReq.Email = email
	otpPayload, _ := json.Marshal(otpReq)
	otpServiceURL := fmt.Sprintf("%s/otp/verify", config.AppConfig.OtpService)

	req, err := http.NewRequest("POST", otpServiceURL, bytes.NewBuffer(otpPayload))
	if err != nil {
		log.Error("Failed to create request to OTP service: %v", err)

		msg := "Internal error preparing request"
		audit.StatusCode = http.StatusInternalServerError
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Error("OTP service request failed: %v", err)

		msg := "OTP service unreachable"
		audit.StatusCode = http.StatusBadGateway
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}
	defer resp.Body.Close()

	// Step 6: Log response status
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		log.Warn("OTP verification failed with status=%d", resp.StatusCode)

		msg := fmt.Sprintf("OTP verification failed with status=%d", resp.StatusCode)
		audit.StatusCode = resp.StatusCode
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.Data(resp.StatusCode, "application/json", body)
		return
	}

	// Step 7: Delete session after successful OTP verification
	if err := redis.DeleteSession(sessionID); err != nil {
		log.Error("Failed to delete session after OTP verification: %v", err)
		// Continue anyway as OTP verification was successful
	}

	log.Info("OTP verified successfully for sessionId=%s", sessionID)

	msg := "OTP verification successful"
	audit.StatusCode = http.StatusOK
	audit.Message = &msg
	log.LogAuditEntry(audit)

	//get access token from the AS.

	// Step 8: Respond to client
	c.Data(resp.StatusCode, "application/json", body)
}
