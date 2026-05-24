package conversations

import (
	"time"

	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Handler handles conversation HTTP requests
type Handler struct {
	repos *repository.Repositories
}

// ConversationResponse represents a conversation in API responses
type ConversationResponse struct {
	ID                 uuid.UUID        `json:"id"`
	Type               string           `json:"type"`
	Name               string           `json:"name"`
	AdminID            *uuid.UUID       `json:"admin_id"`
	CreatedAt          time.Time        `json:"created_at"`
	UpdatedAt          time.Time        `json:"updated_at"`
	Members            []MemberResponse `json:"members,omitempty"`
	LastMessageContent *string          `json:"last_message_content,omitempty"`
	LastMessageAt      *time.Time       `json:"last_message_at,omitempty"`
}

// MemberResponse represents a conversation member in API responses
type MemberResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
