package utils

import (
	"log"
	"time"
	"api-gateway/config"
	"api-gateway/models"
	"github.com/robfig/cron/v3"
)

// StartCleanup kicks off a daily job (at 2 AM) that deletes all audit records older than AuditTTLDays in the env
func StartCleanup(stop <-chan struct{}) {
	c := cron.New()
	_, err := c.AddFunc("0 2 * * *", func() {
		db := config.DB
		if db == nil {
			log.Println("[CLEANUP] DB not initialized, skipping cleanup.")
			return
		}
		cutoff := time.Now().AddDate(0, 0, -config.AppConfig.AuditTTLDays)
		result := db.Where("timestamp < ?", cutoff).Delete(&models.AuditLog{})
		if result.Error != nil {
			log.Printf("[CLEANUP] Error deleting old audit logs: %v", result.Error)
		} else {
			log.Printf("[CLEANUP] Deleted %d old audit logs (before %v)", result.RowsAffected, cutoff)
		}
	})
	if err != nil {
		log.Printf("[CLEANUP] Failed to schedule cleanup job: %v", err)
		return
	}
	c.Start()
	log.Println("[CLEANUP] Scheduled daily audit log cleanup at 2 AM.")

	// Wait for shutdown signal
	<-stop
	log.Println("[CLEANUP] Stopping cleanup cron job.")
	c.Stop()
}
