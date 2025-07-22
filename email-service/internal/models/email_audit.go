package models

import (
	"time"
	"gorm.io/gorm"
)

type EmailAudit struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	From   string         `gorm:"index;not null" json:"email_id"`
	Recipient string         `gorm:"not null" json:"recipient"`
	Status    string         `gorm:"not null" json:"status"`
	Timestamp time.Time       `gorm:"not null" json:"timestamp"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"index" json:"-"`
} 