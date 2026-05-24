package models

import (
	"time"

	"github.com/google/uuid"
)

type Conversation struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Type      string     `gorm:"not null;size:20" json:"type"`
	Name      string     `gorm:"size:200" json:"name"`
	AdminID   *uuid.UUID `gorm:"type:uuid;column:admin_id" json:"admin_id"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (Conversation) TableName() string {
	return "conversations"
}

type ConversationMember struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user" json:"conversation_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_conv_user" json:"user_id"`
	Role           string    `gorm:"not null;size:20;default:'member'" json:"role"`
	JoinedAt       time.Time `json:"joined_at"`
}

func (ConversationMember) TableName() string {
	return "conversation_members"
}
