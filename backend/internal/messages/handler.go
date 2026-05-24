package messages

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewHandler(queries database.Querier) *Handler {
	return &Handler{queries: queries}
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
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
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

	messages, err := h.queries.ListMessagesByConversation(r.Context(), database.ListMessagesByConversationParams{
		ConversationID: convID,
		Column2:        cursor,
		Limit:          int32(limit),
	})
	if err != nil {
		log.Printf("Error listing messages: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list messages")
		return
	}

	respondJSON(w, http.StatusOK, messages)
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
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req struct {
		Content string `json:"content"`
		Type    string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Content == "" {
		respondError(w, http.StatusBadRequest, "Message content is required")
		return
	}

	msgType := req.Type
	if msgType == "" {
		msgType = "text"
	}

	msg, err := h.queries.CreateMessage(r.Context(), database.CreateMessageParams{
		ConversationID: convID,
		SenderID:       senderID,
		Content:        req.Content,
		Type:           msgType,
	})
	if err != nil {
		log.Printf("Error creating message: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to send message")
		return
	}

	respondJSON(w, http.StatusCreated, msg)
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
		respondError(w, http.StatusBadRequest, "Invalid message ID")
		return
	}

	if err := h.queries.DeleteMessage(r.Context(), database.DeleteMessageParams{
		ID:       msgID,
		SenderID: userID,
	}); err != nil {
		log.Printf("Error deleting message: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to delete message")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Message deleted"})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
