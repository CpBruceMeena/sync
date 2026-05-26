package notifications

import "time"

// Notification type constants
const (
	TypeNewMessage  = "new_message"
	TypeReaction    = "reaction"
	TypeGroupInvite = "group_invite"
)

// Handler handles notification HTTP requests
type Handler struct {
	service *Service
}

// NotificationResponse represents a notification returned to the client
type NotificationResponse struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	ReferenceID *string   `json:"reference_id,omitempty"`
	Content     string    `json:"content"`
	IsRead      bool      `json:"is_read"`
	CreatedAt   time.Time `json:"created_at"`
}

// ListNotificationsRequest defines query parameters for listing notifications
type ListNotificationsRequest struct {
	Limit int `json:"limit"`
}

// MarkReadRequest defines the body for marking a notification as read
type MarkReadRequest struct {
	NotificationID string `json:"notification_id"`
	UserID         string `json:"user_id"`
}
