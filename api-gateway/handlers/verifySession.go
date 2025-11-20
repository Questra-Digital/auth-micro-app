package handlers

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"strings"
	"api-gateway/redis"
	"strconv"
)

func VerifySession(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Authorization header is required"})
		return
	}

	//Expected format Bearer <sessionId>
	sessionId := strings.Split(authHeader, " ")[1]
	if sessionId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Session ID is required"})
		return
	}

	// Validate session from Redis
	sessionData, err := redis.GetSessionData(sessionId)
	if err != nil || len(sessionData) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "message": "Invalid session"})
		return
	}

	userIdint, err := strconv.ParseInt(sessionData["userID"], 10, 64)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Session validated", "data":map[string]interface{}{"user_id": userIdint}})
}