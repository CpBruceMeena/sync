package users

import (
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Handler handles user HTTP requests
type Handler struct {
	repos *repository.Repositories
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
}
