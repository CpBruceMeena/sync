package models

import (
	"time"

	"github.com/google/uuid"
)

type Presence struct {
	UserID     uuid.UUID `gorm:"type:uuid;primaryKey" json:"user_id"`
	Status     string    `gorm:"not null;size:20;default:'offline'" json:"status"`
	LastSeenAt time.Time `gorm:"column:last_seen_at;not null;default:now()" json:"last_seen_at"`
}

func (Presence) TableName() string {
	return "presence"
}
