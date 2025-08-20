package config

import (
	"fmt"
	"log"
	"time"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"auth-server/models"
)

var DB *gorm.DB
var sqlDBInstance *gorm.DB // for tracking underlying *sql.DB
var UserDB *gorm.DB
var AuditDB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func GetDatabaseConfig(dbName string) DatabaseConfig {
	return DatabaseConfig{
		Host:     AppConfig.DBHost,
		Port:     AppConfig.DBPort,
		User:     AppConfig.DBUser,
		Password: AppConfig.DBPassword,
		DBName:   dbName,
		SSLMode:  AppConfig.DBSSLMode,
	}
}

// InitDatabase connects to the existing database (does not create it)
func InitDatabase() error {
	// User DB
	userConfig := GetDatabaseConfig(AppConfig.UserDBName)
	userDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		userConfig.Host, userConfig.Port, userConfig.User, userConfig.Password, userConfig.DBName, userConfig.SSLMode,
	)
	udb, err := gorm.Open(postgres.Open(userDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to user database: %w", err)
	}
	UserDB = udb

	// Ensure user_role enum exists
	if err := ensureUserRoleEnum(UserDB); err != nil {
		return fmt.Errorf("failed to ensure user_role enum: %w", err)
	}

	// Audit DB
	auditConfig := GetDatabaseConfig(AppConfig.AuditDBName)
	auditDSN := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		auditConfig.Host, auditConfig.Port, auditConfig.User, auditConfig.Password, auditConfig.DBName, auditConfig.SSLMode,
	)
	adb, err := gorm.Open(postgres.Open(auditDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to audit database: %w", err)
	}
	AuditDB = adb

	// Connection pool settings for both
	for _, db := range []*gorm.DB{UserDB, AuditDB} {
		sqlDB, err := db.DB()
		if err != nil {
			return fmt.Errorf("failed to get sql.DB: %w", err)
		}
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)
	}

	log.Println("Connected to PostgreSQL user and audit databases successfully")
	if err := CreateTable(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	return nil
}

func ensureUserRoleEnum(db *gorm.DB) error {
	return db.Exec(`DO $$ BEGIN IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN CREATE TYPE user_role AS ENUM ('user', 'admin'); END IF; END $$;`).Error
}

// CloseDatabaseConnection safely closes the database connection
func CloseDatabaseConnection() {
	for _, db := range []*gorm.DB{UserDB, AuditDB} {
		if db == nil {
			continue
		}
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Failed to retrieve sql.DB for closing: %v", err)
			continue
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error while closing DB connection: %v", err)
		} else {
			log.Println("Database connection closed successfully")
		}
	}
}

func CreateTable() error {
	if err := UserDB.AutoMigrate(&models.User{}); err != nil {
		return fmt.Errorf("failed to auto-migrate user table: %w", err)
	}
	if err := AuditDB.AutoMigrate(&models.AuditRecord{}); err != nil {
		return fmt.Errorf("failed to auto-migrate audit table: %w", err)
	}
	log.Println("Database tables created/verified successfully")
	return nil
}