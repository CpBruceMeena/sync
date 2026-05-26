package messages

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NewHandler creates a new messages HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// ListMessages returns messages for a conversation with cursor-based pagination
// @Summary List messages
// @Description Get paginated messages for a conversation
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param limit query int false "Number of messages to return (max 100)" default(50)
// @Param cursor query string false "Cursor for pagination (message ID)"
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/messages [get]
func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	cursorStr := r.URL.Query().Get("cursor")
	var cursor uuid.UUID
	if cursorStr != "" {
		cursor, err = uuid.Parse(cursorStr)
		if err != nil {
			cursor = uuid.Nil
		}
	}

	msgs, err := h.service.ListMessages(r.Context(), convID, cursor, limit)
	if err != nil {
		log.Printf("Error listing messages: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list messages")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, msgs)
}

// SendMessage sends a new message to a conversation
// @Summary Send message
// @Description Send a message to a conversation
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/messages [post]
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	senderID := r.Context().Value("user_id").(uuid.UUID)

	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req SendMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	msg, err := h.service.SendMessage(r.Context(), senderID, convID, req.Content, req.Type, req.Attachment)
	if err != nil {
		log.Printf("Error creating message: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, msg)
}

// DeleteMessage deletes a message
// @Summary Delete message
// @Description Delete a message by its ID (only the sender can delete)
// @Tags messages
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Message ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/messages/{id} [delete]
func (h *Handler) DeleteMessage(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	msgIDStr := chi.URLParam(r, "id")
	msgID, err := uuid.Parse(msgIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.service.DeleteMessage(r.Context(), msgID, userID); err != nil {
		log.Printf("Error deleting message: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "Message deleted"})
}
