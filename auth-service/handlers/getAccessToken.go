package handlers

import (
	"auth-server/config"
	"auth-server/models"
	"auth-server/utils"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GetAccessTokenRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func GetAccessToken(c *gin.Context) {
	var req GetAccessTokenRequest
	logger := utils.NewLogger()

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid email format"})
		return
	}

	var user models.User
	err := config.UserDB.Where("email = ?", req.Email).First(&user).Error

	if err != nil {
		// User doesn't exist, create a new one
		newUser := models.User{
			Email: req.Email,
			Role:  models.RoleUser, // Default role
		}

		if err := config.UserDB.Create(&newUser).Error; err != nil {
			logger.Warn("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		user = newUser
		logger.Info("New user created: %s", user.Email)
	}

	// Define scopes based on user role
	var scopes []string
	if user.Role == models.RoleAdmin {
		scopes = []string{"read", "write", "admin"}
	} else {
		scopes = []string{"read", "write"}
	}

	token, err := utils.GenerateJWT(user.ID, user.Email, scopes, time.Hour*1)
	if err != nil {
		logger.Warn("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Log audit record
	action := models.TokenIssued
	description := "Access token issued via passwordless authentication"

	// Convert string scopes to ScopeType for audit logging
	scopeTypes := make([]models.ScopeType, len(scopes))
	for i, s := range scopes {
		scopeTypes[i] = models.ScopeType(s)
	}

	logger.LogAuditRecord(models.AuditRecord{
		UserID:      user.ID,
		Action:      action,
		Status:      models.StatusSuccess,
		ClientIP:    c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		Description: description,
		Scopes:      models.ScopesToString(scopeTypes),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
		"user_id":      user.ID,
		"email":        user.Email,
		"role":         user.Role,
	})
}
