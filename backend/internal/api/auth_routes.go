package api

import (
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/go-chi/chi/v5"
)

// registerPublicAuthRoutes sets up authentication routes accessible without a token
func registerPublicAuthRoutes(r chi.Router, h *auth.Handler) {
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Post("/refresh", h.Refresh)
	})
}

// registerProtectedAuthRoutes sets up authentication routes that require a valid token
func registerProtectedAuthRoutes(r chi.Router, h *auth.Handler) {
	r.Post("/api/auth/logout", h.Logout)
	r.Get("/api/auth/me", h.Me)
}
