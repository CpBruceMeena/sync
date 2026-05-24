package conversations

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func NewHandler(queries database.Querier) *Handler {
	return &Handler{queries: queries}
}

// ListConversations returns all conversations for the authenticated user
// @Summary List conversations
// @Description Get all conversations (private and group) for the authenticated user
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/conversations [get]
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	convs, err := h.queries.ListUserConversations(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing conversations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	respondJSON(w, http.StatusOK, convs)
}

// CreateConversation creates a new conversation
// @Summary Create conversation
// @Description Create a new private or group conversation
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations [post]
func (h *Handler) CreateConversation(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	var req struct {
		Type    string   `json:"type"`
		Name    string   `json:"name"`
		Members []string `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Type != "private" && req.Type != "group" {
		respondError(w, http.StatusBadRequest, "Conversation type must be 'private' or 'group'")
		return
	}

	if req.Type == "private" {
		if len(req.Members) != 1 {
			respondError(w, http.StatusBadRequest, "Private conversation requires exactly one other user")
			return
		}

		otherUser, err := h.queries.GetUserByUsername(r.Context(), req.Members[0])
		if err != nil {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}

		// Check if private conversation already exists
		existing, err := h.queries.FindPrivateConversation(r.Context(), database.FindPrivateConversationParams{
			UserID:   userID,
			UserID_2: otherUser.ID,
		})
		if err == nil && existing.ID != uuid.Nil {
			respondJSON(w, http.StatusOK, existing)
			return
		}

		conv, err := h.queries.CreateConversation(r.Context(), database.CreateConversationParams{
			Type:    "private",
			Name:    "",
			AdminID: pgtype.UUID{Valid: false}, // null
		})
		if err != nil {
			log.Printf("Error creating conversation: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to create conversation")
			return
		}

		// Add both members
		h.queries.AddConversationMember(r.Context(), database.AddConversationMemberParams{
			ConversationID: conv.ID,
			UserID:         userID,
			Role:           "member",
		})
		h.queries.AddConversationMember(r.Context(), database.AddConversationMemberParams{
			ConversationID: conv.ID,
			UserID:         otherUser.ID,
			Role:           "member",
		})

		respondJSON(w, http.StatusCreated, conv)
		return
	}

	// Group conversation
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "Group name is required")
		return
	}

	if len(req.Members) == 0 {
		respondError(w, http.StatusBadRequest, "At least one member is required")
		return
	}

	conv, err := h.queries.CreateConversation(r.Context(), database.CreateConversationParams{
		Type:    "group",
		Name:    req.Name,
		AdminID: pgtype.UUID{Bytes: userID, Valid: true},
	})
	if err != nil {
		log.Printf("Error creating group: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Add admin
	h.queries.AddConversationMember(r.Context(), database.AddConversationMemberParams{
		ConversationID: conv.ID,
		UserID:         userID,
		Role:           "admin",
	})

	// Add members
	for _, memberUsername := range req.Members {
		memberUser, err := h.queries.GetUserByUsername(r.Context(), memberUsername)
		if err != nil {
			continue
		}
		h.queries.AddConversationMember(r.Context(), database.AddConversationMemberParams{
			ConversationID: conv.ID,
			UserID:         memberUser.ID,
			Role:           "member",
		})
	}

	respondJSON(w, http.StatusCreated, conv)
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
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/members [post]
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := r.PathValue("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.queries.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	_, err = h.queries.AddConversationMember(r.Context(), database.AddConversationMemberParams{
		ConversationID: convID,
		UserID:         user.ID,
		Role:           "member",
	})
	if err != nil {
		log.Printf("Error adding member: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to add member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member added"})
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
// @Failure 500 {object} map[string]string
// @Router /api/conversations/{id}/members/{userId} [delete]
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := r.PathValue("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	memberIDStr := r.PathValue("userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	if err := h.queries.RemoveConversationMember(r.Context(), database.RemoveConversationMemberParams{
		ConversationID: convID,
		UserID:         memberID,
	}); err != nil {
		log.Printf("Error removing member: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member removed"})
}

// GetConversation returns a specific conversation by ID
// @Summary Get conversation
// @Description Get a conversation by its ID
// @Tags conversations
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/conversations/{id} [get]
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	convIDStr := r.PathValue("id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	conv, err := h.queries.GetConversationByID(r.Context(), convID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	respondJSON(w, http.StatusOK, conv)
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
