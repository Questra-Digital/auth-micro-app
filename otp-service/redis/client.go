package redis

import (
	"context"
	"otp-service/models"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"time"
	"os"
	"strconv"
	"fmt"
)

var rdb *redis.Client

func InitRedis() {
	addr := fmt.Sprintf("%s:%s", getEnv("REDIS_HOST", "localhost"), getEnv("REDIS_PORT", "6379"))
	password := getEnv("REDIS_PASSWORD", "")
	dbStr := getEnv("REDIS_DB", "0")
	db, err := strconv.Atoi(dbStr)
	if err != nil {
		db = 0
	}

	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
}

func GetClient() *redis.Client {
	return rdb
}

func StoreSession(session models.OTPSession, ttl time.Duration) error {
	data, err := json.Marshal(session)
	if err != nil {
		return err
	}
	return rdb.Set(context.Background(), session.SessionID, data, ttl).Err()
}

func GetSession(sessionID string) (models.OTPSession, error) {
	val, err := rdb.Get(context.Background(), sessionID).Result()
	if err != nil {
		return models.OTPSession{}, err
	}
	var session models.OTPSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return models.OTPSession{}, err
	}
	return session, nil
}

func DeleteSession(sessionID string) {
	rdb.Del(context.Background(), sessionID)
}

// IncrementReattempts increments the Attempts field of an OTPSession in Redis, preserving TTL
func IncrementReattempts(sessionID string) error {
	ctx := context.Background()

	// Fetch existing session data
	val, err := rdb.Get(ctx, sessionID).Result()
	if err != nil {
		return err
	}

	var session models.OTPSession
	if err := json.Unmarshal([]byte(val), &session); err != nil {
		return err
	}

	// Increment Attempts
	session.Attempts++

	// Get current TTL (preserve original expiration)
	ttl, err := rdb.TTL(ctx, sessionID).Result()
	if err != nil || ttl <= 0 {
		return err // TTL missing or expired
	}

	// Save updated session with the same TTL
	return StoreSession(session, ttl)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
