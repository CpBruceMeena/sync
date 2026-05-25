package models

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;index" json:"conversation_id"`
	SenderID       uuid.UUID `gorm:"type:uuid;not null" json:"sender_id"`
	Content        string    `gorm:"not null" json:"content"`
	Type           string    `gorm:"not null;size:20;default:'text'" json:"type"`
	CreatedAt      time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}

type Reaction struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_msg_user_emoji" json:"message_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_msg_user_emoji" json:"user_id"`
	Emoji     string    `gorm:"not null;uniqueIndex:idx_msg_user_emoji" json:"emoji"`
	CreatedAt time.Time `json:"created_at"`
}

func (Reaction) TableName() string {
	return "reactions"
}

type Attachment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	MessageID uuid.UUID `gorm:"type:uuid;not null;index" json:"message_id"`
	FileUrl   string    `gorm:"column:file_url;not null;size:500" json:"file_url"`
	FileType  string    `gorm:"column:file_type;size:50" json:"file_type"`
	FileName  string    `gorm:"column:file_name;size:255" json:"file_name"`
	FileSize  int64     `gorm:"column:file_size" json:"file_size"`
	CreatedAt time.Time `json:"created_at"`
}

func (Attachment) TableName() string {
	return "attachments"
}

type TypingEvent struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConversationID uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_typing_conv_user" json:"conversation_id"`
	UserID         uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_typing_conv_user" json:"user_id"`
	IsTyping       bool      `gorm:"column:is_typing;not null;default:false" json:"is_typing"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (TypingEvent) TableName() string {
	return "typing_events"
}
