package mailer

import (
	"bytes"
	"email-service/internal/config"
	"html/template"
	"path/filepath"
)

type OTPTemplateData struct {
	OTP     string
	AppName string
}

// ParseOTPTemplate loads and renders the HTML template with dynamic data.
func ParseOTPTemplate(otp string) (string, error) {
	tmplPath := filepath.Join("templates", "otp_email.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	data := OTPTemplateData{
		OTP:     otp,
		AppName: config.AppConfig.SMTPFromName,
	}

	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
