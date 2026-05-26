package users

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Service handles user business logic
type Service struct {
	repos *repository.Repositories
}

// NewService creates a new user service
func NewService(repos *repository.Repositories) *Service {
	return &Service{repos: repos}
}

// ListUsers returns all registered users
func (s *Service) ListUsers(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repos.Users.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, UserResponse{
			ID:          u.ID,
			Username:    u.Username,
			Email:       u.Email,
			DisplayName: u.DisplayName,
			AvatarURL:   u.AvatarUrl,
			Status:      u.Status,
		})
	}
	return response, nil
}

// GetUser returns a specific user by ID
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	}, nil
}

// UpdateProfile updates a user's profile fields
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL, status string) (*UserResponse, error) {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if displayName != "" {
		user.DisplayName = displayName
	}
	if avatarURL != "" {
		user.AvatarUrl = avatarURL
	}
	if status != "" {
		user.Status = status
	}
	if err := s.repos.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	return &UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	}, nil
}
