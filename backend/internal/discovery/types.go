package discovery

import (
	"time"

	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

// Handler handles discovery HTTP requests
type Handler struct {
	service *Service
}

// Service handles discovery business logic
type Service struct {
	repos *repository.Repositories
}

// NewService creates a new discovery service
func NewService(repos *repository.Repositories) *Service {
	return &Service{repos: repos}
}

// NewHandler creates a new discovery HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// UserResult represents a user in search results
type UserResult struct {
	ID          uuid.UUID `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	AvatarURL   string    `json:"avatar_url"`
	Status      string    `json:"status"`
	Bio         string    `json:"bio"`
}

// GroupDetailResponse represents a public group in discovery results
type GroupDetailResponse struct {
	ID          uuid.UUID             `json:"id"`
	Name        string                `json:"name"`
	AdminID     *uuid.UUID            `json:"admin_id"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	MemberCount int                   `json:"member_count"`
	Members     []GroupMemberResponse `json:"members,omitempty"`
}

// GroupMemberResponse represents a member in group details
type GroupMemberResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}
