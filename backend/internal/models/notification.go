package models

import (
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID          uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Type        string     `gorm:"not null;size:50" json:"type"`
	ReferenceID *uuid.UUID `gorm:"type:uuid;column:reference_id" json:"reference_id"`
	Content     string     `gorm:"not null" json:"content"`
	IsRead      bool       `gorm:"column:is_read;not null;default:false" json:"is_read"`
	CreatedAt   time.Time  `json:"created_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
