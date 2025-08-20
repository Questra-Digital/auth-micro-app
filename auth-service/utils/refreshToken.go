package utils
import "github.com/google/uuid"

func GenerateRefreshTokenID() string {
	return uuid.New().String()
}