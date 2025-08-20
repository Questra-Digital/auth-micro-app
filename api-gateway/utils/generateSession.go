package utils

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

func GenerateSessionID() (string, error) {
	bytes := make([]byte, 16) // 128-bit session ID
	_, err := rand.Read(bytes)
	if err != nil {
		return "", errors.New("failed to generate secure session ID")
	}
	return hex.EncodeToString(bytes), nil
}
