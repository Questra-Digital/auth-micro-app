package handlers

import (
	"auth-server/config"
	"auth-server/models"
	"auth-server/utils"
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func LoginHandler(c *gin.Context) {
	var req LoginRequest
	logger := utils.NewLogger()

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid login request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var user models.User
	err := config.UserDB.Where("email = ?", req.Email).First(&user).Error
	if err != nil {
		logger.LogAuditRecord(models.AuditRecord{
			UserID:    req.Email,
			Action:    models.LoginFailure,
			Status:    models.StatusFailure,
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Description: "Email not found",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		logger.LogAuditRecord(models.AuditRecord{
			UserID:    user.ID,
			Action:    models.LoginFailure,
			Status:    models.StatusFailure,
			ClientIP:  c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Description: "Incorrect password",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Define scopes based on user role
	var scopes []string
	scopes = []string{"read", "write", "admin"}
	
	token, err := utils.GenerateJWT(user.ID, user.Email, scopes, time.Hour*1)
	if err != nil {
		logger.Warn("Token generation failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	logger.LogAuditRecord(models.AuditRecord{
		UserID:    user.ID,
		Action:    models.LoginSuccess,
		Status:    models.StatusSuccess,
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Description: "User logged in successfully",
		Scopes:    models.ScopesToString(toScopeTypes(scopes)),
	})

	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
}

// helper to convert []string to []ScopeType
func toScopeTypes(strs []string) []models.ScopeType {
	scopes := make([]models.ScopeType, len(strs))
	for i, s := range strs {
		scopes[i] = models.ScopeType(s)
	}
	return scopes
}
