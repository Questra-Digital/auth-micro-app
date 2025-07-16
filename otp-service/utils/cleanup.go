package utils

import (
	"log"
	"otp-service/config"
	"otp-service/models"
	"os"
	"strconv"
	"time"
	"gorm.io/gorm"
)

// CleanupService handles automatic cleanup of expired audit records
type CleanupService struct {
	db *gorm.DB
}

// NewCleanupService creates a new cleanup service
func NewCleanupService() *CleanupService {
	return &CleanupService{
		db: config.DB,
	}
}

// StartCleanupJob starts the background cleanup job
func (cs *CleanupService) StartCleanupJob() {
	go func() {
		ticker := time.NewTicker(1 * time.Hour) // Run every hour
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := cs.cleanupExpiredRecords(); err != nil {
					log.Printf("Cleanup job failed: %v", err)
				}
			}
		}
	}()
}

// cleanupExpiredRecords removes expired records from both audit tables
func (cs *CleanupService) cleanupExpiredRecords() error {
	now := time.Now()
	
	// Clean up expired OTP events
	if err := cs.db.Where("expires_at < ?", now).Delete(&models.OTPEvent{}).Error; err != nil {
		return err
	}

	// Clean up expired rate limit events
	if err := cs.db.Where("expires_at < ?", now).Delete(&models.RateLimitEvent{}).Error; err != nil {
		return err
	}

	log.Printf("Cleanup job completed at %s", now.Format("2006-01-02 15:04:05"))
	return nil
}

// GetRetentionPeriod returns the retention period based on environment
func GetRetentionPeriod() time.Duration {
	// Try to get from environment variable first
	if retentionStr := os.Getenv("OTP_RETENTION_DAYS"); retentionStr != "" {
		if days, err := strconv.Atoi(retentionStr); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	// Default retention periods
	otpRetention := 30 * 24 * time.Hour // 30 days for OTP events

	// In production, you might want shorter retention
	if os.Getenv("APP_ENV") == "production" {
		otpRetention = 7 * 24 * time.Hour // 7 days for OTP events
	}

	return otpRetention
}

// GetRateLimitRetentionPeriod returns the retention period for rate limit events
func GetRateLimitRetentionPeriod() time.Duration {
	// Try to get from environment variable first
	if retentionStr := os.Getenv("RATE_LIMIT_RETENTION_DAYS"); retentionStr != "" {
		if days, err := strconv.Atoi(retentionStr); err == nil && days > 0 {
			return time.Duration(days) * 24 * time.Hour
		}
	}

	rateLimitRetention := 7 * 24 * time.Hour // 7 days for rate limit events

	// In production, shorter retention
	if os.Getenv("APP_ENV") == "production" {
		rateLimitRetention = 3 * 24 * time.Hour // 3 days for rate limit events
	}

	return rateLimitRetention
} 