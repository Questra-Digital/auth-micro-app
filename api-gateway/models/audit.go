package models

import (
	"time"
	"github.com/google/uuid"
)

// Custom types
type EventGroup string
type EventAction string

// Constants for EventGroup
const (
	EventGroupAuth    EventGroup = "AUTH"
	EventGroupSession EventGroup = "SESSION"
	EventGroupAPI     EventGroup = "API"
	EventGroupError   EventGroup = "ERROR"
)

// Constants for EventAction
const (
	// AUTH group
	ActionSignup       EventAction = "SIGNUP"
	ActionLogin        EventAction = "LOGIN"
	ActionOTPGenerated EventAction = "OTP_GENERATED"
	ActionOTPVerified  EventAction = "OTP_VERIFIED"
	ActionAuthFailed   EventAction = "AUTH_FAILED"

	// SESSION group
	ActionSessionCreated EventAction = "SESSION_CREATED"
	ActionTokenStored    EventAction = "TOKEN_STORED"
	ActionSessionDeleted EventAction = "SESSION_DELETED"

	// API / ERROR
	ActionAccess EventAction = "ACCESS"
	ActionError  EventAction = "ERROR"
)

// AuditLog model
type AuditLog struct {
	EventID      uuid.UUID   `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // UUID primary key
	EventGroup   EventGroup  `gorm:"type:text;not null"`                             // Custom EventGroup type
	EventAction  EventAction `gorm:"type:text;not null"`                             // Custom EventAction type
	UserID       *string     `gorm:"type:text"`                                      // Nullable user/client ID
	SessionID    *string     `gorm:"type:text"`                                      // Nullable session ID
	SourceIP     string      `gorm:"type:text;not null"`
	Endpoint     string      `gorm:"type:text;not null"`
	HTTPMethod   string      `gorm:"type:text;not null"`
	StatusCode   int         `gorm:"not null"`                                       // HTTP status code returned
	Message      *string     `gorm:"type:text"`                                      // Optional message
	ServiceName  string      `gorm:"type:text;not null"`                             // e.g. API-GATEWAY
	Timestamp    time.Time   `gorm:"autoCreateTime"`                                 // Set on insert
}

// TableName overrides the table name used by GORM
func (AuditLog) TableName() string {
	return "audit_logs"
}

type RequestContext struct {
	IP     string
	Path   string
	Method string
}