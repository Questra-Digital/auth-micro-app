package handlers

import (
	"auth-server/config"
	"auth-server/models"
	"auth-server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

func RegisterUser(c *gin.Context) {
	var req RegisterRequest
	logger := utils.NewLogger()

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("Invalid request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var existing models.User
	if err := config.UserDB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already registered"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Warn("Failed to hash password: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	newUser := models.User{
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	if err := config.UserDB.Create(&newUser).Error; err != nil {
		logger.Warn("Failed to create user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register user"})
		return
	}

	logger.Info("User registered: %s", newUser.Email)
	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user_id": newUser.ID,
	})
}
