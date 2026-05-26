package api

import (
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/go-chi/chi/v5"
)

// registerUserRoutes sets up user management routes
func registerUserRoutes(r chi.Router, h *users.Handler) {
	r.Get("/api/users", h.ListUsers)
	r.Get("/api/users/{id}", h.GetUser)
	r.Put("/api/users/me", h.UpdateProfile)
}
