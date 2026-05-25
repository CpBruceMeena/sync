package conversations

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func NewHandler(repos *repository.Repositories) *Handler {
	return &Handler{repos: repos}
}

// ListConversations returns all conversations for the authenticated user
func (h *Handler) ListConversations(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	convs, err := h.repos.Conversations.ListByUserID(r.Context(), userID)
	if err != nil {
		log.Printf("Error listing conversations: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to list conversations")
		return
	}

	response := make([]ConversationResponse, 0, len(convs))
	for _, conv := range convs {
		resp := convToResponse(conv)
		resp.Members = getMembers(h, conv.ID)
		resp.LastMessageContent = getLastMessageContent(h, conv.ID)
		resp.LastMessageAt = getLastMessageAt(h, conv.ID)
		response = append(response, resp)
	}

	respondJSON(w, http.StatusOK, response)
}

// CreateConversation creates a new conversation
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

		otherUser, err := h.repos.Users.GetByUsername(r.Context(), req.Members[0])
		if err != nil {
			respondError(w, http.StatusNotFound, "User not found")
			return
		}

		// Check if private conversation already exists
		existing, err := h.repos.Conversations.FindPrivate(r.Context(), userID, otherUser.ID)
		if err == nil && existing != nil {
			respondJSON(w, http.StatusOK, convToResponse(*existing))
			return
		}

		conv := &models.Conversation{
			Type: "private",
		}
		if err := h.repos.Conversations.Create(r.Context(), conv); err != nil {
			log.Printf("Error creating conversation: %v", err)
			respondError(w, http.StatusInternalServerError, "Failed to create conversation")
			return
		}

		// Add both members
		h.repos.Conversations.AddMember(r.Context(), &models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         userID,
			Role:           "member",
		})
		h.repos.Conversations.AddMember(r.Context(), &models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         otherUser.ID,
			Role:           "member",
		})

		respondJSON(w, http.StatusCreated, convToResponse(*conv))
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

	conv := &models.Conversation{
		Type:    "group",
		Name:    req.Name,
		AdminID: &userID,
	}
	if err := h.repos.Conversations.Create(r.Context(), conv); err != nil {
		log.Printf("Error creating group: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to create group")
		return
	}

	// Add admin
	h.repos.Conversations.AddMember(r.Context(), &models.ConversationMember{
		ConversationID: conv.ID,
		UserID:         userID,
		Role:           "admin",
	})

	// Add members
	for _, memberUsername := range req.Members {
		memberUser, err := h.repos.Users.GetByUsername(r.Context(), memberUsername)
		if err != nil {
			continue
		}
		h.repos.Conversations.AddMember(r.Context(), &models.ConversationMember{
			ConversationID: conv.ID,
			UserID:         memberUser.ID,
			Role:           "member",
		})
	}

	respondJSON(w, http.StatusCreated, convToResponse(*conv))
}

// AddMember adds a member to a group conversation
func (h *Handler) AddMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
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

	user, err := h.repos.Users.GetByUsername(r.Context(), req.Username)
	if err != nil {
		respondError(w, http.StatusNotFound, "User not found")
		return
	}

	if err := h.repos.Conversations.AddMember(r.Context(), &models.ConversationMember{
		ConversationID: convID,
		UserID:         user.ID,
		Role:           "member",
	}); err != nil {
		log.Printf("Error adding member: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to add member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member added"})
}

// RemoveMember removes a member from a group conversation
func (h *Handler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	memberIDStr := chi.URLParam(r, "userId")
	memberID, err := uuid.Parse(memberIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid member ID")
		return
	}

	if err := h.repos.Conversations.RemoveMember(r.Context(), convID, memberID); err != nil {
		log.Printf("Error removing member: %v", err)
		respondError(w, http.StatusInternalServerError, "Failed to remove member")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Member removed"})
}

// GetConversation returns a specific conversation by ID
func (h *Handler) GetConversation(w http.ResponseWriter, r *http.Request) {
	convIDStr := chi.URLParam(r, "id")
	convID, err := uuid.Parse(convIDStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid conversation ID")
		return
	}

	conv, err := h.repos.Conversations.GetByID(r.Context(), convID)
	if err != nil {
		respondError(w, http.StatusNotFound, "Conversation not found")
		return
	}

	resp := convToResponse(*conv)
	resp.Members = getMembers(h, conv.ID)

	respondJSON(w, http.StatusOK, resp)
}

// Helper: convert models.Conversation to ConversationResponse
func convToResponse(conv models.Conversation) ConversationResponse {
	return ConversationResponse{
		ID:        conv.ID,
		Type:      conv.Type,
		Name:      conv.Name,
		AdminID:   conv.AdminID,
		CreatedAt: conv.CreatedAt,
		UpdatedAt: conv.UpdatedAt,
	}
}

// Helper: get members for a conversation
func getMembers(h *Handler, convID uuid.UUID) []MemberResponse {
	members, err := h.repos.Conversations.GetMembers(context.Background(), convID)
	if err != nil {
		return nil
	}

	result := make([]MemberResponse, 0, len(members))
	for _, m := range members {
		user, err := h.repos.Users.GetByID(context.Background(), m.UserID)
		username := ""
		if err == nil && user != nil {
			username = user.Username
		}
		result = append(result, MemberResponse{
			UserID:   m.UserID,
			Username: username,
			Role:     m.Role,
			JoinedAt: m.JoinedAt,
		})
	}
	return result
}

// Helper: get last message content for a conversation
func getLastMessageContent(h *Handler, convID uuid.UUID) *string {
	msgs, err := h.repos.Messages.ListByConversation(context.Background(), convID, uuid.Nil, 1)
	if err != nil || len(msgs) == 0 {
		return nil
	}
	return &msgs[0].Content
}

// Helper: get last message time for a conversation
func getLastMessageAt(h *Handler, convID uuid.UUID) *time.Time {
	msgs, err := h.repos.Messages.ListByConversation(context.Background(), convID, uuid.Nil, 1)
	if err != nil || len(msgs) == 0 {
		return nil
	}
	return &msgs[0].CreatedAt
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
