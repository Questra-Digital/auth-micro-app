package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	APP_MODE string

	Port string

	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFromName string

	RedisAddr     string
	RedisPassword string

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string

	RateLimitPerSecond int

	RabbitMQURL   string
	RabbitMQQueue string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found or failed to load, using environment variables only")
	}

	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))
	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
	rateLimit, _ := strconv.Atoi(getEnv("RATE_LIMIT_PER_SECOND", "5"))

	AppConfig = &Config{
		Port:               getEnv("PORT", "8080"),
		SMTPHost:           getEnv("SMTP_HOST", ""),
		SMTPPort:           smtpPort,
		SMTPUsername:       getEnv("SMTP_USERNAME", ""),
		SMTPPassword:       getEnv("SMTP_PASSWORD", ""),
		SMTPFromName:       getEnv("SMTP_FROM_NAME", "Email Service"),
		RedisAddr:          getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnv("REDIS_PASSWORD", ""),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             dbPort,
		DBUser:             getEnv("DB_USER", "postgres"),
		DBPassword:         getEnv("DB_PASSWORD", ""),
		DBName:             getEnv("DB_NAME", "email_audit"),
		RateLimitPerSecond: rateLimit,
		RabbitMQURL:        getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQQueue:      getEnv("RABBITMQ_QUEUE", "email_queue"),
		APP_MODE:           getEnv("APP_MODE", "development"),
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
