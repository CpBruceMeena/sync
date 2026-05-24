package auth

import (
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Handler handles authentication HTTP requests
type Handler struct {
	authService *Service
	repos       *repository.Repositories
}

// RegisterRequest represents a registration request body
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents a token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  UserResponse `json:"user"`
	Token *TokenPair   `json:"token"`
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
