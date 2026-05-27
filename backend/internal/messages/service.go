package messages

import (
	"context"
	"encoding/json"
	"log"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/google/uuid"
)

// Service handles message business logic
type Service struct {
	repos        *repository.Repositories
	notifService *notifications.Service
	hub          *websocket.Hub
}

// NewService creates a new message service
func NewService(repos *repository.Repositories, notifService *notifications.Service, hub *websocket.Hub) *Service {
	return &Service{repos: repos, notifService: notifService, hub: hub}
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

		// Fetch attachments for this message
		attachments, _ := s.repos.Attachments.GetByMessageID(ctx, msg.ID)
		attachmentResponses := make([]AttachmentResponse, 0, len(attachments))
		for _, att := range attachments {
			attachmentResponses = append(attachmentResponses, AttachmentResponse{
				ID:       att.ID,
				FileURL:  att.FileUrl,
				FileType: att.FileType,
				FileName: att.FileName,
				FileSize: att.FileSize,
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
			Attachments:    attachmentResponses,
		})
	}

	return response, nil
}

// SendMessage creates a new message and sends notifications to conversation members
func (s *Service) SendMessage(ctx context.Context, senderID uuid.UUID, convID uuid.UUID, content string, msgType string, attachment *AttachmentUpload) (*models.Message, error) {
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

	// Get sender username for broadcast
	senderUsername := ""
	sender, err := s.repos.Users.GetByID(ctx, senderID)
	if err == nil && sender != nil {
		senderUsername = sender.Username
	}

	// Create attachment record if provided
	if attachment != nil {
		att := &models.Attachment{
			MessageID: msg.ID,
			FileUrl:   attachment.FileURL,
			FileType:  attachment.FileType,
			FileName:  attachment.FileName,
			FileSize:  attachment.FileSize,
		}
		if err := s.repos.Attachments.Create(ctx, att); err != nil {
			log.Printf("Failed to save attachment for message %s: %v", msg.ID, err)
		}
	}

	// Broadcast message to conversation room via WebSocket
	if s.hub != nil {
		wsMsg := websocket.WSMessage{
			Type:           websocket.TypeNewMessage,
			ConversationID: convID,
			SenderID:       senderID,
			SenderUsername: senderUsername,
			Content:        content,
			MessageID:      msg.ID,
		}
		data, err := json.Marshal(wsMsg)
		if err == nil {
			s.hub.BroadcastToRoom(convID, data, senderID) // Skip sender (they get the message via API response)
		}
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
