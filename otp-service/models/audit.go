package models

import (
	"time"
)

// OTPEvent represents an audit log entry for OTP-related events
type OTPEvent struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	SessionID   string    `json:"session_id" gorm:"type:varchar(255);not null;index"`
	Email       string    `json:"email" gorm:"type:varchar(255);not null;index"`
	EventType   string    `json:"event_type" gorm:"type:varchar(50);not null;index"`
	EventStatus string    `json:"event_status" gorm:"type:varchar(50);not null"`
	OTPHash     string    `json:"otp_hash" gorm:"type:varchar(255)"`
	IPAddress   string    `json:"ip_address" gorm:"type:varchar(45)"`
	UserAgent   string    `json:"user_agent" gorm:"type:text"`
	Attempts    int       `json:"attempts" gorm:"default:0"`
	Resends     int       `json:"resends" gorm:"default:0"`
	ErrorMsg    string    `json:"error_msg" gorm:"type:text"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (OTPEvent) TableName() string {
	return "otp_events"
}

// EventType constants
const (
	EventTypeGenerate = "GENERATE"
	EventTypeVerify   = "VERIFY"
	EventTypeResend   = "RESEND"
	EventTypeExpire   = "EXPIRE"
	EventTypeRateLimit = "RATE_LIMIT"
)

// EventStatus constants
const (
	EventStatusSuccess = "SUCCESS"
	EventStatusFailed  = "FAILED"
	EventStatusExpired = "EXPIRED"
	EventStatusBlocked = "BLOCKED"
) 