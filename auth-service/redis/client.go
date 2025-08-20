package redis

import (
	"auth-server/config"
	"context"
	"encoding/json"
	"fmt"
	"time"
	"github.com/redis/go-redis/v9"
)

var rdb *redis.Client

var sessionTTL time.Duration

type RefreshTokenData struct {
	UserID    string   `json:"user_id"`
	Email     string   `json:"email"`
	Scopes    []string `json:"scopes"`
}

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

func StoreRefreshToken(tokenID string, userID string, email string, scopes []string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", tokenID)

	data := RefreshTokenData{
		UserID:    userID,
		Email:     email,
		Scopes:    scopes,
	}

	ttl := time.Duration(config.AppConfig.RefreshTokenDuration) * time.Hour

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, jsonData, ttl).Err()
}

func GetRefreshToken(tokenID string) (*RefreshTokenData, error) {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", tokenID)

	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data RefreshTokenData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func DeleteRefreshToken(tokenID string) error {
	ctx := context.Background()
	key := fmt.Sprintf("refresh_token:%s", tokenID)
	return rdb.Del(ctx, key).Err()
}
