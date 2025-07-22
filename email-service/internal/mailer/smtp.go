package mailer

import (
	"email-service/internal/config"
	"email-service/internal/logger"
	"email-service/internal/models"
	"fmt"
	"github.com/go-mail/mail"
)

// SendEmail sends an email using SMTP with HTML content.
func SendEmail(data models.EmailJob) error {
	cfg := config.AppConfig

	m := mail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("%s <%s>", cfg.SMTPFromName, cfg.SMTPUsername))
	m.SetHeader("To", data.To)
	m.SetHeader("Subject", data.Subject)
	m.SetBody("text/html", data.HTMLBody)

	d := mail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUsername, cfg.SMTPPassword)
	d.StartTLSPolicy = mail.MandatoryStartTLS

	if err := d.DialAndSend(m); err != nil {
		logger.Error("Failed to send email to %s: %v", data.To, err)
		return fmt.Errorf("failed to send email to %s: %w", data.To, err)
	}

	logger.SecureInfo("Email successfully sent to: %s", data.To)
	return nil
}
