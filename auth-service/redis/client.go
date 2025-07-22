package redis

import (
	"time"
	"fmt"
	"auth-server/config"
	"context"
	"github.com/redis/go-redis/v9"
	"auth-server/utils"
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