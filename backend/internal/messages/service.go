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

	// Fetch read receipts for the conversation
	reads, _ := s.repos.MessageRead.GetByConversation(ctx, convID)

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

		// Compute who has read this message
		readBy := make([]ReadReceiptInfo, 0)
		for _, read := range reads {
			if !read.LastReadAt.Before(msg.CreatedAt) {
				username := ""
				if user, err := s.repos.Users.GetByID(ctx, read.UserID); err == nil && user != nil {
					username = user.Username
				}
				readBy = append(readBy, ReadReceiptInfo{
					UserID:   read.UserID.String(),
					Username: username,
					ReadAt:   read.LastReadAt.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
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
			ReadBy:         readBy,
		})
	}

	return response, nil
}

// SendMessage creates a new message and sends notifications to conversation members
func (s *Service) SendMessage(ctx context.Context, senderID uuid.UUID, convID uuid.UUID, content string, msgType string, attachment *AttachmentUpload) (*MessageResponse, error) {
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

	// Get sender username for broadcast and response
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
		// Ensure all conversation members who are currently connected are subscribed to the room.
		// This handles the case where a conversation was created after the member's initial
		// WebSocket subscription (e.g., a new private conversation).
		members, err := s.repos.Conversations.GetMembers(ctx, convID)
		if err != nil {
			log.Printf("[WS] Error getting members for broadcast: %v", err)
		} else {
			for _, member := range members {
				client := s.hub.GetClient(member.UserID)
				if client != nil {
					log.Printf("[WS] Subscribing user %s (%s) to conv %s", member.UserID, member.UserID, convID)
					s.hub.JoinRoom(convID, client)
				} else {
					log.Printf("[WS] User %s is not connected via WebSocket, skipping room join", member.UserID)
				}
			}
		}

		log.Printf("[WS] Broadcasting message %s to conv %s (%d members)", msg.ID, convID, len(members))

		wsMsg := websocket.WSMessage{
			Type:           websocket.TypeNewMessage,
			ConversationID: convID,
			SenderID:       senderID,
			SenderUsername: senderUsername,
			Content:        content,
			MessageID:      msg.ID,
			Data:           msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Conversation: &websocket.ConversationInfo{
				ID:                 convID,
				LastMessageContent: content,
				LastMessageAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			},
		}
		data, err := json.Marshal(wsMsg)
		if err == nil {
			// Broadcast to ALL room members (including sender) so the sidebar updates for everyone
			s.hub.BroadcastToRoomAll(convID, data)
		}
	}

	// Notify conversation members (except sender)
	s.notifyMembers(ctx, senderID, convID, content)

	// Build a proper response with sender_username included
	resp := &MessageResponse{
		ID:             msg.ID,
		ConversationID: msg.ConversationID,
		SenderID:       msg.SenderID,
		SenderUsername: senderUsername,
		Content:        msg.Content,
		Type:           msg.Type,
		CreatedAt:      msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// If attachment was saved, add it to the response
	if attachment != nil {
		resp.Attachments = []AttachmentResponse{
			{
				ID:       uuid.Nil, // Will be the actual ID if saved
				FileURL:  attachment.FileURL,
				FileType: attachment.FileType,
				FileName: attachment.FileName,
				FileSize: attachment.FileSize,
			},
		}
	}

	return resp, nil
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

// SearchMessages searches messages within a conversation by content
func (s *Service) SearchMessages(ctx context.Context, convID uuid.UUID, query string, limit, offset int) ([]MessageResponse, error) {
	messages, err := s.repos.Messages.SearchByConversation(ctx, convID, query, limit, offset)
	if err != nil {
		return nil, err
	}

	// Fetch read receipts for the conversation
	reads, _ := s.repos.MessageRead.GetByConversation(ctx, convID)

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

		// Compute who has read this message
		readBy := make([]ReadReceiptInfo, 0)
		for _, read := range reads {
			if !read.LastReadAt.Before(msg.CreatedAt) {
				username := ""
				if user, err := s.repos.Users.GetByID(ctx, read.UserID); err == nil && user != nil {
					username = user.Username
				}
				readBy = append(readBy, ReadReceiptInfo{
					UserID:   read.UserID.String(),
					Username: username,
					ReadAt:   read.LastReadAt.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
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
			ReadBy:         readBy,
		})
	}

	return response, nil
}

// DeleteMessage deletes a message by ID
func (s *Service) DeleteMessage(ctx context.Context, msgID, userID uuid.UUID) error {
	return s.repos.Messages.Delete(ctx, msgID, userID)
}
