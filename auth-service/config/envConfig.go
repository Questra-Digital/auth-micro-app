package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// PostgreSQL
	DBHost      string
	DBPort      int
	DBUser      string
	DBPassword  string
	DBSSLMode   string
	AuditDBName string
	UserDBName  string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int
	RedisTTL      int

	// App
	AppEnv       string
	AuditTTLDays int
	RateLimit    int

	JWTSecret string

	AppPort int

	// Token durations
	AccessTokenDuration  int // in hours
	RefreshTokenDuration int // in days
}

var AppConfig Config

func InitConfig() {
	var err error

	AppConfig = Config{
		DBHost:      getEnv("DB_HOST", "localhost"),
		DBUser:      getEnv("DB_USER", "postgres"),
		DBPassword:  getEnv("DB_PASSWORD", ""),
		DBSSLMode:   getEnv("DB_SSLMODE", "disable"),
		AuditDBName: getEnv("AUDIT_DB_NAME", "audit"),
		UserDBName:  getEnv("USER_DB_NAME", "user"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		AppEnv:        getEnv("APP_ENV", "development"),
		JWTSecret:     getEnv("JWT_SECRET", ""),
	}

	if AppConfig.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set and non-empty")
	}
	// Parse DB_PORT
	AppConfig.DBPort, err = parseEnvInt("DB_PORT", 5432)
	if err != nil {
		log.Fatalf("Invalid DB_PORT: %v", err)
	}

	// Parse REDIS_PORT
	AppConfig.RedisPort, err = parseEnvInt("REDIS_PORT", 6379)
	if err != nil {
		log.Fatalf("Invalid REDIS_PORT: %v", err)
	}

	// Parse REDIS_DB
	AppConfig.RedisDB, err = parseEnvInt("REDIS_DB", 0)
	if err != nil {
		log.Fatalf("Invalid REDIS_DB: %v", err)
	}

	// Parse REDIS_TTL
	AppConfig.RedisTTL, err = parseEnvInt("REDIS_TTL", 86400)
	if err != nil {
		log.Fatalf("Invalid REDIS_TTL: %v", err)
	}

	// Parse Audit_TTL_Days
	AppConfig.AuditTTLDays, err = parseEnvInt("Audit_TTL_Days", 30)
	if err != nil {
		log.Fatalf("Invalid Audit_TTL_Days: %v", err)
	}

	AppConfig.RateLimit, err = parseEnvInt("Rate_Limit_Per_Minute", 3)
	if err != nil {
		log.Fatalf("Rate_Limit_Per_Minute: %v", err)
	}

	AppConfig.AppPort, err = parseEnvInt("AS_PORT", 8083)
	if err != nil {
		log.Fatalf("AS_Port: %v", err)
	}

	// Parse token durations
	AppConfig.AccessTokenDuration, err = parseEnvInt("ACCESS_TOKEN_DURATION_HOURS", 1)
	if err != nil {
		log.Fatalf("Invalid ACCESS_TOKEN_DURATION_HOURS: %v", err)
	}

	AppConfig.RefreshTokenDuration, err = parseEnvInt("REFRESH_TOKEN_DURATION_DAYS", 7)
	if err != nil {
		log.Fatalf("Invalid REFRESH_TOKEN_DURATION_DAYS: %v", err)
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func parseEnvInt(key string, defaultVal int) (int, error) {
	if val := os.Getenv(key); val != "" {
		return strconv.Atoi(val)
	}
	return defaultVal, nil
}
