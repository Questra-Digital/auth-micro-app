package redis

import (
	"time"
	"fmt"
	"api-gateway/config"
	"context"
	"github.com/redis/go-redis/v9"
	"api-gateway/utils"
)

var rdb *redis.Client
var ctx = context.Background()
var logger = utils.NewLogger()

var sessionTTL time.Duration

func InitRedis() {
	addr := fmt.Sprintf("%s:%d", config.AppConfig.RedisHost, config.AppConfig.RedisPort)
	password := config.AppConfig.RedisPassword
	db := config.AppConfig.RedisDB
	
	rdb = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	sessionTTL = time.Duration(config.AppConfig.RedisTTL) * time.Second
}

func GetClient() *redis.Client {
	return rdb
}

func CloseRedis() error {
	if rdb != nil {
		return rdb.Close()
	}
	return nil
}

// StoreSessionData stores all session data in a hash with optional expiry
func StoreSessionData(sessionID, clientID, jwtToken, email string, expiry ...time.Duration) error {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	fields := map[string]interface{}{
		"clientID": clientID,
		"token": jwtToken,
		"email": email,
	}
	// Use custom expiry if provided, otherwise fall back to sessionTTL
	expireAfter := sessionTTL
	if len(expiry) > 0 {
		expireAfter = expiry[0]
	}
	pipe := rdb.TxPipeline()
	pipe.HSet(ctx, hashKey, fields)
	pipe.Expire(ctx, hashKey, expireAfter)
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Failed to store session data for sessionID %s: %v", sessionID, err)
		return err
	}
	logger.Debug("Stored session data for sessionID %s", sessionID)
	return nil
}

// GetSessionData retrieves all session data from the hash
func GetSessionData(sessionID string) (map[string]string, error) {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	data, err := rdb.HGetAll(ctx, hashKey).Result()
	if err != nil {
		logger.Error("Failed to get session data for sessionID %s: %v", sessionID, err)
		return nil, err
	}
	if len(data) == 0 {
		logger.Warn("No session data found for sessionID %s", sessionID)
	}
	return data, nil
}

// UpdateSessionField updates a single field in the session hash
func UpdateSessionField(sessionID, field, value string) error {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	pipe := rdb.TxPipeline()
	pipe.HSet(ctx, hashKey, field, value)
	pipe.Expire(ctx, hashKey, sessionTTL)
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Failed to update field %s for sessionID %s: %v", field, sessionID, err)
		return err
	}
	logger.Debug("Updated field %s for sessionID %s", field, sessionID)
	return nil
}

func DeleteSession(sessionID string) error {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	pipe := rdb.TxPipeline()
	pipe.Del(ctx, hashKey)
	_, err := pipe.Exec(ctx)
	if err != nil {
		logger.Error("Failed to delete sessionID %s: %v", sessionID, err)
		return err
	}
	logger.Debug("Deleted sessionID %s", sessionID)
	return nil
}

// StoreJWTForSession updates the token field in the session hash
func StoreJWTForSession(sessionID, jwtToken string) error {
	return UpdateSessionField(sessionID, "token", jwtToken)
}

// GetJWTForSession retrieves the token field from the session hash
func GetJWTForSession(sessionID string) (string, error) {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	val, err := rdb.HGet(ctx, hashKey, "token").Result()
	if err != nil {
		logger.Error("Failed to get JWT for sessionID %s: %v", sessionID, err)
		return "", err
	}
	return val, nil
}

// StoreEmailForSession updates the email field in the session hash
func StoreEmailForSession(sessionID, email string) error {
	return UpdateSessionField(sessionID, "email", email)
}

// GetEmailForSession retrieves the email field from the session hash
func GetEmailForSession(sessionID string) (string, error) {
	hashKey := fmt.Sprintf("session:%s", sessionID)
	val, err := rdb.HGet(ctx, hashKey, "email").Result()
	if err != nil {
		logger.Error("Failed to get email for sessionID %s: %v", sessionID, err)
		return "", err
	}
	return val, nil
}