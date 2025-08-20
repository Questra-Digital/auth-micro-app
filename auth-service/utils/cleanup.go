package utils

import (
	"log"
	"time"
	"auth-server/config"
	"auth-server/models"
	"github.com/robfig/cron/v3"
)

// StartCleanup schedules a daily cleanup of old audit records based on TTL
func StartCleanup(stop <-chan struct{}) {
	c := cron.New()

	_, err := c.AddFunc("0 2 * * *", func() {
		db := config.AuditDB
		if db == nil {
			log.Println("[CLEANUP] DB not initialized, skipping cleanup.")
			return
		}

		cutoff := time.Now().AddDate(0, 0, -config.AppConfig.AuditTTLDays)
		result := db.Where("timestamp < ?", cutoff).Delete(&models.AuditRecord{})
		if result.Error != nil {
			log.Printf("[CLEANUP] Error deleting old audit records: %v", result.Error)
		} else {
			log.Printf("[CLEANUP] Deleted %d audit records older than %v", result.RowsAffected, cutoff)
		}
	})

	if err != nil {
		log.Printf("[CLEANUP] Failed to schedule cleanup job: %v", err)
		return
	}

	c.Start()
	log.Println("[CLEANUP] Scheduled daily audit record cleanup at 2 AM.")

	<-stop
	log.Println("[CLEANUP] Stopping cleanup cron job.")
	c.Stop()
}
