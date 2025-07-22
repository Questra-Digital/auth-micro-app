package config

import (
	"fmt"
	"log"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"otp-service/models"
)

var DB *gorm.DB
var sqlDBInstance *gorm.DB // for tracking underlying *sql.DB

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		Host:     AppConfig.DBHost,
		Port:     AppConfig.DBPort,//Convert int to string
		User:     AppConfig.DBUser,
		Password: AppConfig.DBPassword,
		DBName:   AppConfig.DBName,
		SSLMode:  AppConfig.DBSSLMode,
	}
}

// InitDatabase connects to the existing database (does not create it)
func InitDatabase() error {
	config := GetDatabaseConfig()

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	DB = db
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Connection pool settings
	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	log.Println("Connected to PostgreSQL successfully")
	if err := CreateTable(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

// CloseDatabaseConnection safely closes the database connection
func CloseDatabaseConnection() {
	if DB == nil {
		return
	}
	sqlDB, err := DB.DB()
	if err != nil {
		log.Printf("Failed to retrieve sql.DB for closing: %v", err)
		return
	}
	if err := sqlDB.Close(); err != nil {
		log.Printf("Error while closing DB connection: %v", err)
	} else {
		log.Println("Database connection closed successfully")
	}
}

func CreateTable() error {
	err := DB.AutoMigrate(&models.OTPEvent{})
	if err != nil {
		return fmt.Errorf("failed to auto-migrate tables: %w", err)
	}
	log.Println("Database tables created/verified successfully")
	return nil
}