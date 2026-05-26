package notifications

import (
	"log"
	"net/http"
	"strconv"

	"github.com/CpBruceMeena/sync/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// NewHandler creates a new notifications HTTP handler
func NewHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// ListNotifications returns all notifications for the authenticated user
//
// swagger:route GET /api/notifications notifications listNotifications
//
// Returns paginated notifications for the authenticated user.
//
// Responses:
//
//	200: []NotificationResponse
//	401: ErrorResponse
func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
			limit = v
		}
	}

	notifs, err := h.service.ListNotifications(r.Context(), userID, limit)
	if err != nil {
		log.Printf("Error listing notifications: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to list notifications")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, notifs)
}

// MarkRead marks a single notification as read
//
// swagger:route PUT /api/notifications/{id}/read notifications markNotificationRead
//
// Marks a specific notification as read.
//
// Responses:
//
//	200: map[string]string
//	400: ErrorResponse
//	401: ErrorResponse
func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	notifIDStr := chi.URLParam(r, "id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		httputil.RespondError(w, http.StatusBadRequest, "Invalid notification ID")
		return
	}

	if err := h.service.MarkRead(r.Context(), notifID, userID); err != nil {
		log.Printf("Error marking notification as read: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to mark notification as read")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "Notification marked as read"})
}

// MarkAllRead marks all notifications as read for the authenticated user
//
// swagger:route PUT /api/notifications/read-all notifications markAllNotificationsRead
//
// Marks all notifications as read.
//
// Responses:
//
//	200: map[string]string
//	401: ErrorResponse
func (h *Handler) MarkAllRead(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	if err := h.service.MarkAllRead(r.Context(), userID); err != nil {
		log.Printf("Error marking all notifications as read: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to mark all notifications as read")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]string{"message": "All notifications marked as read"})
}

// GetUnreadCount returns the unread notification count for the authenticated user
//
// swagger:route GET /api/notifications/unread-count notifications getUnreadCount
//
// Returns the number of unread notifications.
//
// Responses:
//
//	200: UnreadCountResponse
//	401: ErrorResponse
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("user_id").(uuid.UUID)

	count, err := h.service.GetUnreadCount(r.Context(), userID)
	if err != nil {
		log.Printf("Error getting unread count: %v", err)
		httputil.RespondError(w, http.StatusInternalServerError, "Failed to get unread count")
		return
	}

	httputil.RespondJSON(w, http.StatusOK, map[string]int64{"count": count})
}
