package users

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NewHandler creates a new users HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// ListUsers returns all registered users
// @Summary List users
// @Description Get a list of all registered users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserResponse
// @Failure 500 {object} map[string]string
// @Router /api/users [get]
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.service.ListUsers(r.Context())
	if err != nil {
		log.Printf("Error listing users: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list users")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, users)
}

// GetUser returns a specific user by ID
// @Summary Get user
// @Description Get a user by their ID
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/users/{id} [get]
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.service.GetUser(r.Context(), userID)
	if err != nil {
		httputil.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user)
}

// UpdateProfile updates the authenticated user's profile
// @Summary Update profile
// @Description Update the display name, avatar URL, or status of the authenticated user
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/users/me [put]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateProfile(r.Context(), userID, req.DisplayName, req.AvatarURL, req.Status, req.Bio)
	if err != nil {
		log.Printf("Error updating user: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, user)
}
