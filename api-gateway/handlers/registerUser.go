package handlers

import (
	"api-gateway/config"
	"api-gateway/redis"
	"api-gateway/utils"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func RegisterUserHandler(c *gin.Context) {
	log := utils.NewLogger()

	// Step 1: Get sessionId from cookie
	sessionID, err := c.Cookie("sessionId")
	if err != nil || sessionID == "" {
		log.Warn("Missing or invalid sessionId cookie")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid sessionId"})
		return
	}

	// Step 2: Fetch session data from Redis
	sessionData, err := redis.GetSessionData(sessionID)
	if err != nil || len(sessionData) == 0 {
		log.Warn("Invalid sessionId: %s", sessionID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}

	// Step 3: Check clientID matches
	clientID := c.ClientIP()
	if sessionData["clientID"] != clientID {
		log.Warn("Session clientID mismatch: got %s, expected %s", sessionData["clientID"], clientID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session does not belong to this client"})
		return
	}

	// Step 4: Check session is verified
	if sessionData["verified"] != "true" {
		log.Warn("Session not verified for sessionId=%s", sessionID)
		c.JSON(http.StatusForbidden, gin.H{"error": "Session not verified"})
		return
	}

	// Step 5: Parse registration request
	var req RegisterUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid registration payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Ensure email matches session email
	sessionEmail := sessionData["email"]
	if sessionEmail != req.Email {
		log.Warn("Email mismatch: session email = %s, request email = %s", sessionEmail, req.Email)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email does not match session"})
		return
	}

	// Step 6: Forward request to AuthorizationServer
	authURL := fmt.Sprintf("%s/registerUser", config.AppConfig.AuthorizationService)
	jsonBody, _ := json.Marshal(req)
	forwardReq, err := http.NewRequest("POST", authURL, bytes.NewReader(jsonBody))
	if err != nil {
		log.Error("Failed to create request to AuthorizationServer: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	forwardReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(forwardReq)
	if err != nil {
		log.Error("Failed to reach AuthorizationServer: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "Authorization service unreachable"})
		return
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusCreated {
		// Delete session from Redis
		err := redis.DeleteSession(sessionID)
		if err != nil {
			log.Warn("Failed to delete session from Redis after registration: %v", err)
		}
		// Delete sessionId cookie
		cookie := &http.Cookie{
			Name:     "sessionId",
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
		}
		http.SetCookie(c.Writer, cookie)
	}

	c.Data(resp.StatusCode, "application/json", respBody)
} 