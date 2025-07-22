package models

import (
	"time"
	"strings"
)

// ActionType defines types of actions to be logged in audit
type ActionType string

const (
	LoginSuccess     ActionType = "LOGIN_SUCCESS"
	LoginFailure     ActionType = "LOGIN_FAILURE"
	Logout           ActionType = "LOGOUT"
	TokenIssued      ActionType = "TOKEN_ISSUED"
	TokenRevoked     ActionType = "TOKEN_REVOKED"
	PermissionCheck  ActionType = "PERMISSION_CHECK"
)

// StatusType defines the outcome of an action
type StatusType string

const (
	StatusSuccess StatusType = "SUCCESS"
	StatusFailure StatusType = "FAILURE"
)

// ScopeType defines possible token scopes
type ScopeType string

const (
	ScopeRead  ScopeType = "read"
	ScopeWrite ScopeType = "write"
	ScopeAdmin ScopeType = "admin"
)

// AuditRecord represents an audit trail entry
type AuditRecord struct {
	ID          int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	UserID      string     `json:"user_id"`
	Action      ActionType `json:"action"`
	Status      StatusType `json:"status"`
	Timestamp   time.Time  `json:"timestamp"`
	ClientIP    string     `json:"client_ip"`
	UserAgent   string     `json:"user_agent"`
	Description string     `json:"description"`
	Scopes      string     `json:"scopes"` // comma-separated string e.g. "read,write"
}

// TableName overrides the default table name
func (AuditRecord) TableName() string {
	return "audit_records"
}

// Utility function : convert slice of scopes to string
func ScopesToString(scopes []ScopeType) string {
	strs := make([]string, len(scopes))
	for i, s := range scopes {
		strs[i] = string(s)
	}
	return strings.Join(strs, ",")
}

// Utility function : convert comma-separated scopes to slice
func StringToScopes(s string) []ScopeType {
	parts := strings.Split(s, ",")
	scopes := make([]ScopeType, len(parts))
	for i, part := range parts {
		scopes[i] = ScopeType(strings.TrimSpace(part))
	}
	return scopes
}
