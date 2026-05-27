package users

import (
	"context"

	"github.com/CpBruceMeena/sync/internal/models"
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

func userToResponse(u models.User) UserResponse {
	return UserResponse{
		ID:          u.ID,
		Username:    u.Username,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarURL:   u.AvatarUrl,
		Status:      u.Status,
		Bio:         u.Bio,
	}
}

// ListUsers returns all registered users
func (s *Service) ListUsers(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repos.Users.List(ctx)
	if err != nil {
		return nil, err
	}

	response := make([]UserResponse, 0, len(users))
	for _, u := range users {
		response = append(response, userToResponse(u))
	}
	return response, nil
}

// GetUser returns a specific user by ID
func (s *Service) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repos.Users.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := userToResponse(*user)
	return &resp, nil
}

// UpdateProfile updates a user's profile fields
func (s *Service) UpdateProfile(ctx context.Context, userID uuid.UUID, displayName, avatarURL, status, bio string) (*UserResponse, error) {
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
	if bio != "" {
		user.Bio = bio
	}
	if err := s.repos.Users.Update(ctx, user); err != nil {
		return nil, err
	}

	resp := userToResponse(*user)
	return &resp, nil
}
