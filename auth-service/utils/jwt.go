package utils

import (
	"time"
	"github.com/golang-jwt/jwt/v5"
	"auth-server/config"
)

var jwtSecret = []byte(config.AppConfig.JWTSecret) // You should load this from env or config

type CustomClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Scopes []string `json:"scopes"`
	jwt.RegisteredClaims
}

// GenerateJWT generates a signed JWT access token
func GenerateJWT(userID, email string, scopes []string, duration time.Duration) (string, error) {
	claims := CustomClaims{
		UserID: userID,
		Email:  email,
		Scopes: scopes,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}