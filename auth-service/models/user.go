package models

import (
	"time"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Keys      *Keys     `gorm:"foreignKey:UserId;references:ID" json:"keys"`
}

// TableName returns the database table name for the User model
func (User) TableName() string {
	return "users"
}

type Keys struct {
	ID     int64  `gorm:"primaryKey; autoIncrement" json:"id"`
	UserId int64  `gorm:"uniqueIndex;not null" json:"user_id"`
	APIKey string `gorm:"unique;not null" json:"api_key"`
}

func (Keys) TableName() string {
	return "user_keys"
}
