package reactions

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NewHandler creates a new reactions HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// ToggleReaction adds a reaction if it doesn't exist, or removes it if it does
// @Summary Toggle reaction
// @Description Add or remove an emoji reaction on a message
// @Tags reactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Message ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/messages/{id}/reactions [post]
func (h *Handler) ToggleReaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)
	username := r.Context().Value("username").(string)

	msgIDStr := chi.URLParam(r, "id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	var req ReactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Emoji == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Emoji is required")
		return
	}

	reactions, err := h.service.ToggleReaction(r.Context(), msgID, userID, username, req.Emoji)
	if err != nil {
		log.Printf("Error toggling reaction: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to toggle reaction")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]interface{}{
		"reactions": reactions,
	})
}
