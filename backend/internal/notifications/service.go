package notifications

import (
	"context"
	"log"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Service handles notification business logic
type Service struct {
	repos *repository.Repositories
}

// NewService creates a new notification service
func NewService(repos *repository.Repositories) *Service {
	return &Service{repos: repos}
}

// CreateNotification creates a new notification for a user
func (s *Service) CreateNotification(ctx context.Context, userID uuid.UUID, notifType string, referenceID *uuid.UUID, content string) error {
	notif := &models.Notification{
		UserID:      userID,
		Type:        notifType,
		ReferenceID: referenceID,
		Content:     content,
	}
	return s.repos.Notifications.Create(ctx, notif)
}

// ListNotifications returns notifications for a user with an optional limit
func (s *Service) ListNotifications(ctx context.Context, userID uuid.UUID, limit int) ([]NotificationResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	notifs, err := s.repos.Notifications.List(ctx, userID, limit)
	if err != nil {
		log.Printf("Error listing notifications: %v", err)
		return nil, err
	}

	resp := make([]NotificationResponse, len(notifs))
	for i, n := range notifs {
		resp[i] = NotificationResponse{
			ID:          n.ID.String(),
			Type:        n.Type,
			ReferenceID: refIDToStringPtr(n.ReferenceID),
			Content:     n.Content,
			IsRead:      n.IsRead,
			CreatedAt:   n.CreatedAt,
		}
	}
	return resp, nil
}

// MarkRead marks a single notification as read
func (s *Service) MarkRead(ctx context.Context, notifID, userID uuid.UUID) error {
	return s.repos.Notifications.MarkRead(ctx, notifID, userID)
}

// MarkAllRead marks all notifications as read for a user
func (s *Service) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	return s.repos.Notifications.MarkAllRead(ctx, userID)
}

// GetUnreadCount returns the unread notification count for a user
func (s *Service) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repos.Notifications.GetUnreadCount(ctx, userID)
}

// helper to convert *uuid.UUID to *string
func refIDToStringPtr(refID *uuid.UUID) *string {
	if refID == nil {
		return nil
	}
	s := refID.String()
	return &s
}
