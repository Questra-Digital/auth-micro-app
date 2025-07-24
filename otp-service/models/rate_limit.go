package models

import (
	"time"
)

// RateLimitEvent represents a rate limiting audit log entry
type RateLimitEvent struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	IPAddress   string    `json:"ip_address" gorm:"type:varchar(45);not null;index"`
	UserAgent   string    `json:"user_agent" gorm:"type:text"`
	Endpoint    string    `json:"endpoint" gorm:"type:varchar(255);not null"`
	Method      string    `json:"method" gorm:"type:varchar(10);not null"`
	RateLimit   string    `json:"rate_limit" gorm:"type:varchar(50);not null"`
	RequestCount int      `json:"request_count" gorm:"default:0"`
	Limit       int       `json:"limit" gorm:"default:0"`
	Window      string    `json:"window" gorm:"type:varchar(20);not null"`
	Blocked     bool      `json:"blocked" gorm:"default:false"`
	ExpiresAt   time.Time `json:"expires_at" gorm:"index"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for GORM
func (RateLimitEvent) TableName() string {
	return "rate_limit_events"
}

// RateLimitType constants
const (
	RateLimitTypeIP     = "IP_BASED"
	RateLimitTypeUser   = "USER_BASED"
	RateLimitTypeGlobal = "GLOBAL"
)

// RateLimitWindow constants
const (
	RateLimitWindowMinute = "1m"
	RateLimitWindowHour   = "1h"
	RateLimitWindowDay    = "1d"
) 