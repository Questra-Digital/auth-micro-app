// internal/cleanup/audit_purger.go
package cleanup

import (
    "time"

    "github.com/robfig/cron/v3"
    "email-service/internal/config"
    "email-service/internal/logger"
    "email-service/internal/models"
)

// StartAuditPurger kicks off a daily job (at 2 AM) that deletes
// any EmailAudit rows older than 7 days.
func StartAuditPurger() {
    c := cron.New()
    // cron spec: minute hour day-of-month month day-of-week
    _, err := c.AddFunc("0 2 * * *", func() {
        cutoff := time.Now().Add(-7 * 24 * time.Hour)

        res := config.DB.
            Where("timestamp < ?", cutoff).
            Delete(&models.EmailAudit{})

        if err := res.Error; err != nil {
            logger.Error("[AuditPurger] purge failed: %v", err)
            return
        }

        logger.SecureInfo("[AuditPurger] purged %d old audit records", res.RowsAffected)
        // And always log an info line:
        logger.Info("[AuditPurger] purged %d old audit records", res.RowsAffected)
    })
    if err != nil {
        logger.Error("failed to schedule audit purger: %v", err)
    }
    c.Start()
}
