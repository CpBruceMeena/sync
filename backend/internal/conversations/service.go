package conversations

import (
	"context"
	"log"
	"time"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Service handles conversation business logic
type Service struct {
	repos        *repository.Repositories
	notifService *notifications.Service
}

// NewService creates a new conversation service
func NewService(repos *repository.Repositories, notifService *notifications.Service) *Service {
	return &Service{repos: repos, notifService: notifService}
}

// ListConversations returns all conversations for a user with member and message info
func (s *Service) ListConversations(ctx context.Context, userID uuid.UUID) ([]ConversationResponse, error) {
	convs, err := s.repos.Conversations.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	response := make([]ConversationResponse, 0, len(convs))
	for _, conv := range convs {
		resp := convToResponse(conv)
		resp.Members = s.getMembers(conv.ID)
		resp.LastMessageContent = s.getLastMessageContent(conv.ID)
		resp.LastMessageAt = s.getLastMessageAt(conv.ID)
		response = append(response, resp)
	}
	return response, nil
}

// GetConversation returns a specific conversation by ID
func (s *Service) GetConversation(ctx context.Context, convID uuid.UUID) (*ConversationResponse, error) {
	conv, err := s.repos.Conversations.GetByID(ctx, convID)
	if err != nil {
		return nil, err
	}

	resp := convToResponse(*conv)
	resp.Members = s.getMembers(conv.ID)
	return &resp, nil
}

// CreatePrivateConversation creates a private conversation between two users
func (s *Service) CreatePrivateConversation(ctx context.Context, userID uuid.UUID, otherUsername string) (*ConversationResponse, error) {
	otherUser, err := s.repos.Users.GetByUsername(ctx, otherUsername)
	if err != nil {
		return nil, err
	}

	// Check if private conversation already exists
	existing, err := s.repos.Conversations.FindPrivate(ctx, userID, otherUser.ID)
	if err == nil && existing != nil {
		resp := convToResponse(*existing)
		return &resp, nil
	}

	conv := &models.Conversation{
		Type: "private",
	}
	if err := s.repos.Conversations.Create(ctx, conv); err != nil {
		return nil, err
	}

	// Add both members
	s.repos.Conversations.AddMember(ctx, &models.ConversationMember{
		ConversationID: conv.ID,
		UserID:         userID,
		Role:           "member",
	})
	s.repos.Conversations.AddMember(ctx, &models.ConversationMember{
		ConversationID: conv.ID,
		UserID:         otherUser.ID,
		Role:           "member",
	})

	resp := convToResponse(*conv)
	return &resp, nil
}

// CreateGroupConversation creates a group conversation
func (s *Service) CreateGroupConversation(ctx context.Context, userID uuid.UUID, name string, memberUsernames []string, isPublic bool) (*ConversationResponse, error) {
	conv := &models.Conversation{
		Type:     "group",
		Name:     name,
		AdminID:  &userID,
		IsPublic: isPublic,
	}
	if err := s.repos.Conversations.Create(ctx, conv); err != nil {
		return nil, err
	}

	// Add admin
	s.repos.Conversations.AddMember(ctx, &models.ConversationMember{
		ConversationID: conv.ID,
		UserID:         userID,
		Role:           "admin",
	})

	// Add members
	for _, memberUsername := range memberUsernames {
		memberUser, err := s.repos.Users.GetByUsername(ctx, memberUsername)
		if err != nil {
			continue
		}
		s.repos.Conversations.AddMember(ctx, &models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         memberUser.ID,
			Role:           "member",
		})
	}

	resp := convToResponse(*conv)
	return &resp, nil
}

// AddMember adds a user to a conversation by username and sends a notification
func (s *Service) AddMember(ctx context.Context, convID uuid.UUID, username string) error {
	user, err := s.repos.Users.GetByUsername(ctx, username)
	if err != nil {
		return err
	}

	if err := s.repos.Conversations.AddMember(ctx, &models.ConversationMember{
		ConversationID: convID,
		UserID:         user.ID,
		Role:           "member",
	}); err != nil {
		return err
	}

	// Send notification to the new member
	if s.notifService != nil {
		conv, err := s.repos.Conversations.GetByID(ctx, convID)
		if err == nil && conv != nil {
			content := "You were added to " + conv.Name
			refID := convID
			if err := s.notifService.CreateNotification(ctx, user.ID, notifications.TypeGroupInvite, &refID, content); err != nil {
				log.Printf("Error creating group invite notification: %v", err)
			}
		}
	}

	return nil
}

// RemoveMember removes a member from a conversation
func (s *Service) RemoveMember(ctx context.Context, convID, memberID uuid.UUID) error {
	return s.repos.Conversations.RemoveMember(ctx, convID, memberID)
}

// --- Helper methods ---

func convToResponse(conv models.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:        conv.ID,
		Type:      conv.Type,
		Name:      conv.Name,
		AdminID:   conv.AdminID,
		IsPublic:  conv.IsPublic,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}
}

func (s *Service) getMembers(convID uuid.UUID) []MemberResponse {
	members, err := s.repos.Conversations.GetMembers(context.Background(), convID)
	if err != nil {
		return nil
	}

	result := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		user, err := s.repos.Users.GetByID(context.Background(), m.UserID)
		username := ""
		if err == nil && user != nil {
			username = user.Username
		}
		result = append(result, MemberResponse{
			UserID:   m.UserID,
			Username: username,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}
	return result
}

func (s *Service) getLastMessageContent(convID uuid.UUID) *string {
	msgs, err := s.repos.Messages.ListByConversation(context.Background(), convID, uuid.Nil, 1)
	if err != nil || len(msgs) == 0 {
		return nil
	}
	return &msgs[0].Content
}

func (s *Service) getLastMessageAt(convID uuid.UUID) *time.Time {
	msgs, err := s.repos.Messages.ListByConversation(context.Background(), convID, uuid.Nil, 1)
	if err != nil || len(msgs) == 0 {
		return nil
	}
	return &msgs[0].CreatedAt
}
