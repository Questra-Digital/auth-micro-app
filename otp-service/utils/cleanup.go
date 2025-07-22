package utils

import (
	"log"
	"time"

	"otp-service/config"
	"otp-service/models"

	"github.com/robfig/cron/v3"
)

var cleanupCron *cron.Cron

func StartCleanupJob() {
	cleanupCron = cron.New()

	// Schedule job to run every day at 2:00 AM
	_, err := cleanupCron.AddFunc("0 2 * * *", func() {
		log.Println("[CRON] Running OTPEvent cleanup task...")

		ttlDays := config.AppConfig.OTPEventTTLDays
		threshold := time.Now().AddDate(0, 0, -ttlDays)

		result := config.DB.Where("created_at < ?", threshold).Delete(&models.OTPEvent{})
		if result.Error != nil {
			log.Printf("[CRON] Failed to delete old OTP events: %v", result.Error)
			return
		}

		log.Printf("[CRON] Deleted %d old OTP events.", result.RowsAffected)
	})

	if err != nil {
		log.Fatalf("[CRON] Failed to schedule cleanup job: %v", err)
	}

	cleanupCron.Start()
}

// StopCleanupJob gracefully stops the cleanup cron job
func StopCleanupJob() {
	if cleanupCron != nil {
		log.Println("[CRON] Stopping cleanup job...")
		ctx := cleanupCron.Stop()
		<-ctx.Done()
		log.Println("[CRON] Cleanup job stopped successfully")
	}
}
