package api

import (
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/go-chi/chi/v5"
)

// registerNotificationRoutes sets up notification routes
func registerNotificationRoutes(r chi.Router, h *notifications.Handler) {
	r.Get("/api/notifications", h.ListNotifications)
	r.Get("/api/notifications/unread-count", h.GetUnreadCount)
	r.Put("/api/notifications/{id}/read", h.MarkRead)
	r.Put("/api/notifications/read-all", h.MarkAllRead)
}
