package messages

import (
	"github.com/google/uuid"
)

// Handler handles message HTTP requests
type Handler struct {
	service *Service
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID             uuid.UUID          `json:"id"`
	ConversationID uuid.UUID          `json:"conversation_id"`
	SenderID       uuid.UUID          `json:"sender_id"`
	SenderUsername string             `json:"sender_username"`
	Content        string             `json:"content"`
	Type           string             `json:"type"`
	CreatedAt      string             `json:"created_at"`
	Reactions      []ReactionResponse `json:"reactions,omitempty"`
}

// ReactionResponse represents a reaction in API responses
type ReactionResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// SendMessageRequest represents a send message request body
type SendMessageRequest struct {
	Content string `json:"content"`
	Type    string `json:"type"`
}
