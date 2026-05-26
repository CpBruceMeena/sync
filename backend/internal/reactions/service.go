package reactions

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

// Service handles reaction business logic
type Service struct {
	repos        *repository.Repositories
	hub          *websocket.Hub
	notifService *notifications.Service
}

// NewService creates a new reaction service
func NewService(repos *repository.Repositories, hub *websocket.Hub, notifService *notifications.Service) *Service {
	return &Service{repos: repos, hub: hub, notifService: notifService}
}

// ToggleReaction adds a reaction if not present, removes if present, fetches updated
// reactions with usernames, and broadcasts via WebSocket.
func (s *Service) ToggleReaction(ctx context.Context, msgID, userID uuid.UUID, username, emoji string) ([]ReactionResponse, error) {
	// Get the message to verify it exists and get conversation ID
	msg, err := s.repos.Messages.GetByID(ctx, msgID)
	if err != nil {
		return nil, err
	}

	// Check if reaction already exists (for toggle behavior)
	existingReactions, err := s.repos.Messages.GetReactionsByMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}

	reactionExists := false
	for _, rxn := range existingReactions {
		if rxn.UserID == userID && rxn.Emoji == emoji {
			reactionExists = true
			break
		}
	}

	var wsEventType string
	if reactionExists {
		if err := s.repos.Messages.RemoveReaction(ctx, msgID, userID, emoji); err != nil {
			return nil, err
		}
		wsEventType = websocket.TypeReactionRemoved
	} else {
		reaction := &models.Reaction{
			MessageID: msgID,
			UserID:    userID,
			Emoji:     emoji,
		}
		if err := s.repos.Messages.AddReaction(ctx, reaction); err != nil {
			return nil, err
		}
		wsEventType = websocket.TypeReactionAdded

		// Notify the message author about the reaction (if not reacting to own message)
		if s.notifService != nil && msg.SenderID != userID {
			content := username + " reacted with " + emoji + " to your message"
			refID := msgID
			if err := s.notifService.CreateNotification(ctx, msg.SenderID, notifications.TypeReaction, &refID, content); err != nil {
				log.Printf("Error creating reaction notification: %v", err)
			}
		}
	}

	// Fetch updated reactions with usernames
	updatedReactions, err := s.repos.Messages.GetReactionsByMessage(ctx, msgID)
	if err != nil {
		return nil, err
	}

	reactionResponses := make([]ReactionResponse, 0, len(updatedReactions))
	for _, rxn := range updatedReactions {
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

	// Broadcast via WebSocket
	s.broadcastReactionEvent(msg.ConversationID, msgID, userID, username, emoji, wsEventType, reactionResponses)

	return reactionResponses, nil
}

// broadcastReactionEvent sends a WebSocket message for a reaction event
func (s *Service) broadcastReactionEvent(convID, msgID, userID uuid.UUID, username, emoji, eventType string, reactions []ReactionResponse) {
	wsMsg := websocket.WSMessage{
		Type:           eventType,
		ConversationID: convID,
		MessageID:      msgID,
		UserID:         userID,
		Username:       username,
		Emoji:          emoji,
		Data:           reactions,
	}

	data, err := json.Marshal(wsMsg)
	if err != nil {
		log.Printf("Error marshaling reaction WS message: %v", err)
		return
	}

	s.hub.BroadcastToRoom(convID, data, userID)
}
