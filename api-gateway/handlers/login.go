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
	"time"

	"github.com/gin-gonic/gin"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(c *gin.Context) {
	log := utils.NewLogger()

	// Step 1: Parse login request
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Warn("Invalid login payload: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// Step 2: Generate new sessionId
	sessionID, err := utils.GenerateSessionID()
	if err != nil {
		log.Error("Failed to generate session ID: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	clientID := c.ClientIP()

	// Step 3: Forward login request to AuthorizationServer
	authURL := fmt.Sprintf("%s/login", config.AppConfig.AuthorizationService)
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

	if resp.StatusCode == http.StatusOK {
		// Parse JWT from response
		var respData map[string]interface{}
		if err := json.Unmarshal(respBody, &respData); err == nil {
			if token, ok := respData["access_token"].(string); ok {
				// Store session data in Redis (verified=true, with JWT)
				err := redis.StoreSessionData(sessionID, clientID, token, req.Email, true)
				if err != nil {
					log.Warn("Failed to store session data in Redis: %v", err)
				}
				// Set sessionId cookie
				cookie := &http.Cookie{
					Name:     "sessionId",
					Value:    sessionID,
					Path:     "/",
					HttpOnly: true,
					Secure:   true,
					SameSite: http.SameSiteLaxMode,
					MaxAge:   int((1 * time.Hour).Seconds()),
				}
				http.SetCookie(c.Writer, cookie)
			}
		}
	}

	c.Data(resp.StatusCode, "application/json", []byte(`"Log In Successful!"`))

} 