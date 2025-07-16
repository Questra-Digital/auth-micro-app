package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"otp-service/models"
)

var DB *gorm.DB

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// GetDatabaseConfig returns database configuration from environment variables
func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "otp_audit"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}
}

// InitDatabase initializes the GORM database connection
func InitDatabase() error {
	config := GetDatabaseConfig()
	
	// First connect to default postgres database to create our target database
	defaultDSN := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.SSLMode)
	
	defaultDB, err := gorm.Open(postgres.Open(defaultDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}

	// Create the target database if it doesn't exist
	var count int64
	defaultDB.Raw("SELECT 1 FROM pg_database WHERE datname = ?", config.DBName).Count(&count)
	if count == 0 {
		createSQL := fmt.Sprintf("CREATE DATABASE %s", config.DBName)
		if err := defaultDB.Exec(createSQL).Error; err != nil {
			return fmt.Errorf("failed to create database %s: %w", config.DBName, err)
		}
		log.Printf("Database '%s' created successfully", config.DBName)
	}

	// Close the default connection
	sqlDB, err := defaultDB.DB()
	if err == nil {
		sqlDB.Close()
	}
	
	// Now connect to our target database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)
	
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), // Set to logger.Silent for production
	})
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Get the underlying sql.DB object to configure connection pool
	sqlDB, err = DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Database connection established successfully")
	return nil
}

// CreateTables creates the necessary tables using GORM AutoMigrate
func CreateTables() error {
	// AutoMigrate will create tables and add missing columns
	err := DB.AutoMigrate(&models.OTPEvent{}, &models.RateLimitEvent{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}

	log.Println("Database tables created/verified successfully")
	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
} 