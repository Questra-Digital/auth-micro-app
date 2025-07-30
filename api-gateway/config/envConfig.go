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
	RedisHost       string
	RedisPort       int
	RedisPassword   string
	RedisDB         int
	SessionTTLHours int // in hours

	// App
	AppEnv       string
	AuditTTLDays int
	RateLimit    int

	//Deployed Services
	OtpService           string
	ApiGatewayPort       int
	AuthorizationService string
	ResourceServiceURL   string
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

		RedisHost:            getEnv("REDIS_HOST", "localhost"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		AppEnv:               getEnv("APP_ENV", "development"),
		OtpService:           getEnv("OTP_SERVICE_URL", "http://otp-service:8080"),
		AuthorizationService: getEnv("AUTHORIZATION_SERVICE_URL", "http://auth-service:8083"),
		ResourceServiceURL:   getEnv("RESOURCE_SERVICE_URL", "http://resource-service:8084"),
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

	// Parse Session TTL in hours
	AppConfig.SessionTTLHours, err = parseEnvInt("SESSION_TTL_HOURS", 24)
	if err != nil {
		log.Fatalf("Invalid SESSION_TTL_HOURS: %v", err)
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

	AppConfig.ApiGatewayPort, err = parseEnvInt("API_GATEWAY_PORT", 8080)
	if err != nil {
		log.Fatalf("API_GATEWAY_PORT: %v", err)
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
