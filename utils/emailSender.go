package utils

import (
	"fmt"
	"gopkg.in/gomail.v2"
	"os"
)

func SendVerificationEmail(to, code string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", os.Getenv("SMTP_USER"))
	m.SetHeader("To", to)
	m.SetHeader("Subject", "Your Verification Code")
	m.SetBody("text/plain", fmt.Sprintf("Your verification code is: %s", code))

	port := 587
	d := gomail.NewDialer(os.Getenv("SMTP_HOST"), port, os.Getenv("SMTP_USER"), os.Getenv("SMTP_PASS"))

	return d.DialAndSend(m)
}