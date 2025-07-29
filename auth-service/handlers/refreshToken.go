
package handlers

import (
	"auth-server/redis"
	"auth-server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"auth-server/models"
	"strings"
)

type RefreshTokenRequest struct {
	GrantType    string `json:"grant_type" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func RefreshAccessToken(c *gin.Context) {
	var req RefreshTokenRequest
	logger := utils.NewLogger()

	// Validate input
	if err := c.ShouldBindJSON(&req); err != nil || req.GrantType != "refresh_token" {
		logger.Warn("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	tokenID := req.RefreshToken
	data, err := redis.GetRefreshToken(tokenID)
	if err != nil || data == nil {
		logger.Warn("Invalid or expired refresh token: %v", err)
		logger.LogAuditRecord(models.AuditRecord{
			UserID:      "",
			Action:      models.TokenIssued,
			Status:      models.StatusFailure,
			ClientIP:    c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
			Description: "Invalid or expired refresh token.",
			Scopes:      "",
		})
		
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}

	// Generate new access token
	accessToken, err := utils.GenerateJWT(data.UserID, data.Email, data.Scopes)
	if err != nil {
		logger.Warn("Failed to generate access token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate access token"})
		return
	}

	logger.Info("Access token generated successfully")
	logger.LogAuditRecord(models.AuditRecord{
		UserID:      data.UserID,
		Action:      models.TokenIssued,
		Status:      models.StatusSuccess,
		ClientIP:    c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		Description: "Access token and Refresh token issued.",
		Scopes:      strings.Join(data.Scopes, ","),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": tokenID,
	})
}
