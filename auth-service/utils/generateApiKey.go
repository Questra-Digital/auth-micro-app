package utils

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateAPIKey() (string, error) {
	// Generate a 64-byte API key
	apiKey := make([]byte, 64)
	_, err := rand.Read(apiKey)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(apiKey), nil
}