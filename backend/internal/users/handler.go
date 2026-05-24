package users

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/google/uuid"
)

func NewHandler(queries database.Querier) *Handler {
	return &Handler{queries: queries}
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
	users, err := h.queries.ListUsers(r.Context())
	if err != nil {
		log.Printf("Error listing users: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list users")
		return
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

	respondJSON(w, http.StatusOK, response)
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
	userIDStr := r.PathValue("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	user, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	})
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

	var req struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Status      string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.queries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:          userID,
		DisplayName: req.DisplayName,
		AvatarUrl:   req.AvatarURL,
		Status:      req.Status,
	})
	if err != nil {
		log.Printf("Error updating user: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	respondJSON(w, http.StatusOK, UserResponse{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarUrl,
		Status:      user.Status,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
