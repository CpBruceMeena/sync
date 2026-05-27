package discovery

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/google/uuid"
)

// SearchUsers searches for users by username, display name, or email
func (s *Service) SearchUsers(ctx context.Context, query string, limit int) ([]UserResult, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	users, err := s.repos.Users.Search(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	results := make([]UserResult, 0, len(users))
	for _, u := range users {
		results = append(results, UserResult{
			ID:          u.ID,
			Username:    u.Username,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarUrl,
			Bio:         u.Bio,
			Status:      u.Status,
		})
	}
	return results, nil
}

// ListPublicGroups lists public groups with pagination
func (s *Service) ListPublicGroups(ctx context.Context, limit, offset int) ([]GroupDetailResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	convs, err := s.repos.Conversations.ListPublic(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	return s.enrichGroups(ctx, convs), nil
}

// SearchPublicGroups searches public groups by name
func (s *Service) SearchPublicGroups(ctx context.Context, query string, limit int) ([]GroupDetailResponse, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	convs, err := s.repos.Conversations.SearchPublic(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	return s.enrichGroups(ctx, convs), nil
}

// GetGroupDetails returns detailed info about a public group
func (s *Service) GetGroupDetails(ctx context.Context, convID uuid.UUID) (*GroupDetailResponse, error) {
	conv, err := s.repos.Conversations.GetByID(ctx, convID)
	if err != nil {
		return nil, err
	}

	if conv.Type != "group" || !conv.IsPublic {
		return nil, nil
	}

	return s.enrichGroupDetail(ctx, conv), nil
}

func (s *Service) enrichGroups(ctx context.Context, convs []models.Conversation) []GroupDetailResponse {
	results := make([]GroupDetailResponse, 0, len(convs))
	for _, conv := range convs {
		detail := s.enrichGroupDetail(ctx, &conv)
		if detail != nil {
			detail.Members = nil // omit members in list results
			results = append(results, *detail)
		}
	}
	return results
}

func (s *Service) enrichGroupDetail(ctx context.Context, conv *models.Conversation) *GroupDetailResponse {
	members, err := s.repos.Conversations.GetMembers(ctx, conv.ID)
	if err != nil {
		return nil
	}

	detail := &GroupDetailResponse{
		ID:          conv.ID,
		Name:        conv.Name,
		AdminID:     conv.AdminID,
		CreatedAt:   conv.CreatedAt,
		UpdatedAt:   conv.UpdatedAt,
		MemberCount: len(members),
	}

	for _, m := range members {
		user, err := s.repos.Users.GetByID(ctx, m.UserID)
		username := ""
		if err == nil && user != nil {
			username = user.Username
		}
		detail.Members = append(detail.Members, GroupMemberResponse{
			UserID:   m.UserID,
			Username: username,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}

	return detail
}
