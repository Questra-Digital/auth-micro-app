package utils

import (
	"golang.org/x/crypto/bcrypt"
)

func HashOTP(otp string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	return string(bytes), err
}

func CompareOTP(hash, otp string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(otp)) == nil
}
