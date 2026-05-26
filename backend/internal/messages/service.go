package messages

import (
	"context"
	"log"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Service handles message business logic
type Service struct {
	repos        *repository.Repositories
	notifService *notifications.Service
}

// NewService creates a new message service
func NewService(repos *repository.Repositories, notifService *notifications.Service) *Service {
	return &Service{repos: repos, notifService: notifService}
}

// ListMessages returns paginated messages with sender info and reactions
func (s *Service) ListMessages(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]MessageResponse, error) {
	messages, err := s.repos.Messages.ListByConversation(ctx, convID, cursor, limit)
	if err != nil {
		return nil, err
	}

	response := make([]MessageResponse, 0, len(messages))
	for _, msg := range messages {
		sender, err := s.repos.Users.GetByID(ctx, msg.SenderID)
		senderUsername := ""
		if err == nil && sender != nil {
			senderUsername = sender.Username
		}

		// Fetch reactions for this message
		reactions, _ := s.repos.Messages.GetReactionsByMessage(ctx, msg.ID)
		reactionResponses := make([]ReactionResponse, 0, len(reactions))
		for _, rxn := range reactions {
			user, err := s.repos.Users.GetByID(ctx, rxn.UserID)
			rxnUsername := ""
			if err == nil && user != nil {
				rxnUsername = user.Username
			}
			reactionResponses = append(reactionResponses, ReactionResponse{
				UserID:   rxn.UserID.String(),
				Username: rxnUsername,
				Emoji:    rxn.Emoji,
			})
		}

		response = append(response, MessageResponse{
			ID:             msg.ID,
			ConversationID: msg.ConversationID,
			SenderID:       msg.SenderID,
			SenderUsername: senderUsername,
			Content:        msg.Content,
			Type:           msg.Type,
			CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Reactions:      reactionResponses,
		})
	}

	return response, nil
}

// SendMessage creates a new message and sends notifications to conversation members
func (s *Service) SendMessage(ctx context.Context, senderID uuid.UUID, convID uuid.UUID, content string, msgType string) (*models.Message, error) {
	if msgType == "" {
		msgType = "text"
	}

	msg := &models.Message{
		ConversationID: convID,
		SenderID:       senderID,
		Content:        content,
		Type:           msgType,
	}
	if err := s.repos.Messages.Create(ctx, msg); err != nil {
		return nil, err
	}

	// Notify conversation members (except sender)
	s.notifyMembers(ctx, senderID, convID, content)

	return msg, nil
}

// notifyMembers sends notifications to all conversation members except the sender
func (s *Service) notifyMembers(ctx context.Context, senderID, convID uuid.UUID, content string) {
	if s.notifService == nil {
		return
	}

	members, err := s.repos.Conversations.GetMembers(ctx, convID)
	if err != nil {
		log.Printf("Error fetching members for notification: %v", err)
		return
	}

	for _, member := range members {
		if member.UserID == senderID {
			continue
		}

		refID := convID
		if err := s.notifService.CreateNotification(ctx, member.UserID, notifications.TypeNewMessage, &refID, content); err != nil {
			log.Printf("Error creating notification for user %s: %v", member.UserID, err)
		}
	}
}

// DeleteMessage deletes a message by ID
func (s *Service) DeleteMessage(ctx context.Context, msgID, userID uuid.UUID) error {
	return s.repos.Messages.Delete(ctx, msgID, userID)
}
