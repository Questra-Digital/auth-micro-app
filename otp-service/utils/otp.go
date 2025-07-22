package utils

import (
	"crypto/rand"
)

func GenerateSecureOTP(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	for i := 0; i < length; i++ {
		bytes[i] = (bytes[i] % 10) + '0'
	}
	return string(bytes)
}
