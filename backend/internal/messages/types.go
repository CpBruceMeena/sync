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
	ID             uuid.UUID            `json:"id"`
	ConversationID uuid.UUID            `json:"conversation_id"`
	SenderID       uuid.UUID            `json:"sender_id"`
	SenderUsername string               `json:"sender_username"`
	Content        string               `json:"content"`
	Type           string               `json:"type"`
	CreatedAt      string               `json:"created_at"`
	Reactions      []ReactionResponse   `json:"reactions,omitempty"`
	Attachments    []AttachmentResponse `json:"attachments,omitempty"`
}

// ReactionResponse represents a reaction in API responses
type ReactionResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Emoji    string `json:"emoji"`
}

// AttachmentResponse represents an attachment in API responses
type AttachmentResponse struct {
	ID       uuid.UUID `json:"id"`
	FileURL  string    `json:"file_url"`
	FileType string    `json:"file_type"`
	FileName string    `json:"file_name"`
	FileSize int64     `json:"file_size"`
}

// AttachmentUpload represents attachment data sent with a message
type AttachmentUpload struct {
	FileURL  string `json:"file_url"`
	FileType string `json:"file_type"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
}

// SendMessageRequest represents a send message request body
type SendMessageRequest struct {
	Content    string            `json:"content"`
	Type       string            `json:"type"`
	Attachment *AttachmentUpload `json:"attachment,omitempty"`
}
