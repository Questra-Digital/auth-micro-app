package models

import "time"

type OTPSession struct {
	SessionID string    `json:"session_id"`
	OTPHash   string    `json:"otp_hash"`
	CreatedAt time.Time `json:"created_at"`
	Attempts  int       `json:"attempts"`
	Resends   int       `json:"resends"`
}

