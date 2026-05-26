package conversations

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NewHandler creates a new conversations HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// ListConversations returns all conversations for the authenticated user
// @Summary List conversations
// @Description Get all conversations (private and group) for the authenticated user
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ConversationResponse
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations [get]
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	convs, err := h.service.ListConversations(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing conversations: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, convs)
}

// CreateConversation creates a new conversation
// @Summary Create conversation
// @Description Create a new private or group conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} ConversationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations [post]
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req CreateConversationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Type != "private" && req.Type != "group" {
		httputil.RespondError(w, http.StatusBadRequest, "Conversation type must be 'private' or 'group'")
		return
	}

	var conv *ConversationResponse
	var err error

	if req.Type == "private" {
		if len(req.Members) != 1 {
			httputil.RespondError(w, http.StatusBadRequest, "Private conversation requires exactly one other user")
			return
		}

		conv, err = h.service.CreatePrivateConversation(r.Context(), userID, req.Members[0])
		if err != nil {
			log.Printf("Error creating private conversation: %v", err)
			httputil.RespondError(w, http.StatusNotFound, "User not found")
			return
		}

		httputil.RespondJSON(w, http.StatusCreated, conv)
		return
	}

	// Group conversation
	if req.Name == "" {
		httputil.RespondError(w, http.StatusBadRequest, "Group name is required")
		return
	}
	if len(req.Members) == 0 {
		httputil.RespondError(w, http.StatusBadRequest, "At least one member is required")
		return
	}

	conv, err = h.service.CreateGroupConversation(r.Context(), userID, req.Name, req.Members, req.IsPublic)
	if err != nil {
		log.Printf("Error creating group: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, conv)
}

// AddMember adds a member to a group conversation
// @Summary Add member
// @Description Add a user to a group conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/members [post]
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddMember(r.Context(), convID, req.Username); err != nil {
		log.Printf("Error adding member: %v", err)
		httputil.RespondError(w, http.StatusNotFound, "User not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "Member added"})
}

// RemoveMember removes a member from a group conversation
// @Summary Remove member
// @Description Remove a user from a group conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Param userId path string true "User ID to remove"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/members/{userId} [delete]
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	memberIDStr := chi.URLParam(r, "userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	if err := h.service.RemoveMember(r.Context(), convID, memberID); err != nil {
		log.Printf("Error removing member: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "Member removed"})
}

// GetConversation returns a specific conversation by ID
// @Summary Get conversation
// @Description Get a conversation by its ID
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} ConversationResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/conversations/{id} [get]
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	conv, err := h.service.GetConversation(r.Context(), convID)
	if err != nil {
		httputil.RespondError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, conv)
}
