package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	// PostgreSQL
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// Redis
	RedisHost     string
	RedisPort     int
	RedisPassword string
	RedisDB       int

	// App
	AppEnv          string
	OTPEventTTLDays int
	OtpServicePort int
	EmailServiceUrl string
}

var AppConfig Config

func InitConfig() {
	var err error

	AppConfig = Config{
		DBHost:     getEnv("DB_HOST", "localhost"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", "otp_audit"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		RedisHost:     getEnv("REDIS_HOST", "localhost"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		AppEnv:        getEnv("APP_ENV", "development"),
		EmailServiceUrl: getEnv("Email_Service_URL","http://localhost:8082"),
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

	// Parse OTP_EVENT_TTL_DAYS
	AppConfig.OTPEventTTLDays, err = parseEnvInt("OTP_EVENT_TTL_DAYS", 30)
	if err != nil {
		log.Fatalf("Invalid OTP_EVENT_TTL_DAYS: %v", err)
	}

	// Parse OTP_SERVICE_PORT
	AppConfig.OtpServicePort, err = parseEnvInt("OTP_SERVICE_PORT", 8081)
	if err != nil {
		log.Fatalf("Invalid OTP_SERVICE_PORT: %v", err)
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