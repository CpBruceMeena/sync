package api

import (
	"net/http"
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/middleware"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/CpBruceMeena/sync/docs/swagger"
)

// SetupRoutes configures all API routes and returns a chi router.
// Business logic remains in handler files within their respective packages.
func SetupRoutes(
	authHandler *auth.Handler,
	usersHandler *users.Handler,
	conversationsHandler *conversations.Handler,
	messagesHandler *messages.Handler,
	wsHandler *websocket.WsHandler,
	authService *auth.Service,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Public auth routes
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/register", authHandler.Register)
		r.Post("/login", authHandler.Login)
		r.Post("/refresh", authHandler.Refresh)
	})

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authService))

		// Auth
		r.Post("/api/auth/logout", authHandler.Logout)
		r.Get("/api/auth/me", authHandler.Me)

		// Users
		r.Get("/api/users", usersHandler.ListUsers)
		r.Get("/api/users/{id}", usersHandler.GetUser)
		r.Put("/api/users/me", usersHandler.UpdateProfile)

		// Conversations
		r.Get("/api/conversations", conversationsHandler.ListConversations)
		r.Post("/api/conversations", conversationsHandler.CreateConversation)
		r.Get("/api/conversations/{id}", conversationsHandler.GetConversation)
		r.Post("/api/conversations/{id}/members", conversationsHandler.AddMember)
		r.Delete("/api/conversations/{id}/members/{userId}", conversationsHandler.RemoveMember)

		// Messages
		r.Get("/api/conversations/{id}/messages", messagesHandler.ListMessages)
		r.Post("/api/conversations/{id}/messages", messagesHandler.SendMessage)
		r.Delete("/api/messages/{id}", messagesHandler.DeleteMessage)
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// WebSocket endpoint
	r.Get("/ws", wsHandler.ServeWS)

	return r
}
