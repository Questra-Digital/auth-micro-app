package handlers

import (
	"api-gateway/api"
	"api-gateway/models"
	"api-gateway/redis"
	"api-gateway/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type otpVerifyRequest struct {
	OTP   string `json:"otp"`
	Email string `json:"email"`
}

func VerifyOTPHandler(c *gin.Context) {
	log := utils.NewLogger()
	otpClient := api.NewOTPClient()
	authClient := api.NewAuthClient()

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
		log.Warn("Missing sessionId cookie")

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

	// Step 5: Verify OTP using API client
	resp, err := otpClient.VerifyOTP(otpReq.OTP, email, sessionID)
	if err != nil {
		log.Error("OTP service request failed: %v", err)

		msg := "OTP service unreachable"
		audit.StatusCode = http.StatusBadGateway
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}

	// Step 6: Check OTP verification response
	respBody, _ := api.ReadResponseBody(resp)
	if resp.StatusCode != http.StatusOK {
		log.Warn("OTP verification failed with status=%d", resp.StatusCode)

		msg := "OTP verification failed"
		audit.StatusCode = resp.StatusCode
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.Data(resp.StatusCode, "application/json", respBody)
		return
	}

	// Step 7: Delete session after successful OTP verification
	if err := redis.DeleteSession(sessionID); err != nil {
		log.Error("Failed to delete session after OTP verification: %v", err)
		// Continue anyway as OTP verification was successful
	}

	log.Info("OTP verified successfully for sessionId=%s", sessionID)

	// Step 8: Delete the previous sessionId from the cookie
	c.SetCookie("sessionId", "", -1, "/", "", true, true)

	// Step 9: Get access token using API client
	authResp, err := authClient.GetAccessToken(email)
	if err != nil {
		log.Error("Auth service request failed: %v", err)

		msg := "Auth service unreachable"
		audit.StatusCode = http.StatusBadGateway
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(http.StatusBadGateway, gin.H{"error": msg})
		return
	}

	// Step 10: Process auth response
	authRespBody, _ := api.ReadResponseBody(authResp)
	if authResp.StatusCode != http.StatusOK {
		log.Error("Auth service responded with status %d: %s", authResp.StatusCode, string(authRespBody))

		msg := "Failed to get access token"
		audit.StatusCode = authResp.StatusCode
		audit.Message = &msg
		log.LogAuditEntry(audit)

		c.JSON(authResp.StatusCode, gin.H{"error": msg})
		return
	}

	// Step 11: Parse auth response and extract tokens
	authResponse, err := api.ParseAuthResponse(authRespBody)
	if err != nil {
		log.Error("Failed to parse auth response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid auth response format"})
		return
	}

	accessToken, refreshToken, refreshTokenDuration, err := api.ExtractTokensAndDuration(authResponse)
	if err != nil {
		log.Error("Missing tokens in auth response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid auth response - missing tokens"})
		return
	}

	// Step 12: Generate a new sessionId
	newSessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Error("Failed to generate new session ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Step 13: Store the access token, refresh token, and refreshTokenId in Redis
	// Use refresh token duration for session TTL
	sessionTTL := time.Duration(refreshTokenDuration) * 24 * time.Hour
	if err := redis.StoreSessionData(newSessionID, clientID, accessToken, email, refreshToken, sessionTTL); err != nil {
		log.Error("Failed to store new session data: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Step 14: Send the new sessionId as a cookie to the client
	// Cookie MaxAge should match session data TTL (refresh token duration)
	newCookie := &http.Cookie{
		Name:     "sessionId",
		Value:    newSessionID,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()), // Use refresh token duration
	}
	http.SetCookie(c.Writer, newCookie)

	// Step 15: Respond with HTTP status 200 OK and a success message
	msg := "OTP verification successful"
	audit.StatusCode = http.StatusOK
	audit.Message = &msg
	log.LogAuditEntry(audit)

	c.JSON(http.StatusOK, gin.H{
		"message": "OTP verification successful",
		"status":  "success",
	})
}
