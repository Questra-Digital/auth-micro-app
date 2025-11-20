package handlers

import (
	"auth-server/config"
	"auth-server/models"
	"auth-server/redis"
	"auth-server/utils"
	"net/http"
	"strconv"
	"strings"
	"time"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"errors"
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

	if errors.Is(err, gorm.ErrRecordNotFound) {
		
		logger.Info("User not found, creating a new one")

		// Generate a new API key for the user
		apiKey, err := utils.GenerateAPIKey()
		if err != nil {
			logger.Warn("Failed to generate API key: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "message": "Failed to generate API key"})
			return
		}


		user = models.User{
			Email: req.Email,
			Role:  models.RoleUser, // Default role
			Keys: &models.Keys{
				APIKey: apiKey,
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		logger.Info("New API key generated for user %s: %s", user.Email, apiKey)
		logger.LogAuditRecord(models.AuditRecord{
			UserID:      strconv.FormatInt(int64(user.ID), 10),
			Action:      models.TokenIssued,
			Status:      models.StatusSuccess,
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Description: "New API key generated for user",
			Scopes:      "",
		})

		if err := config.UserDB.Create(&user).Error; err != nil {
			logger.Warn("Failed to create user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		logger.Info("New user created: %s", user.Email)
	}

	// Define scopes based on user role
	var scopes []string
	if user.Role == models.RoleAdmin {
		scopes = []string{"read", "write", "admin"}
	} else {
		scopes = []string{"read", "write"}
	}

	token, err := utils.GenerateJWT(strconv.FormatInt(user.ID, 10), user.Email, scopes)
	if err != nil {
		logger.Warn("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshTokenID := utils.GenerateRefreshTokenID()
	redis.StoreRefreshToken(refreshTokenID, strconv.FormatInt(user.ID,10), user.Email, scopes)
	// Log audit record
	action := models.TokenIssued
	description := "Access token and Refresh token issued."

	logger.LogAuditRecord(models.AuditRecord{
		UserID:      strconv.FormatInt(user.ID, 10),
		Action:      action,
		Status:      models.StatusSuccess,
		ClientIP:    c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		Description: description,
		Scopes:      strings.Join(scopes, ","),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":                token,
		"refresh_token":               refreshTokenID,
		"refresh_token_duration_days": config.AppConfig.RefreshTokenDuration,
		"user_id":                     user.ID,
	})
}
