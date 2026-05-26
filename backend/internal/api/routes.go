package api

import (
	"net/http"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/files"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/middleware"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/reactions"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/CpBruceMeena/sync/docs/swagger"
)

// SetupRoutes configures all API routes and returns a chi router.
// Business logic resides in service files within their respective packages.
func SetupRoutes(
	authHandler *auth.Handler,
	usersHandler *users.Handler,
	conversationsHandler *conversations.Handler,
	messagesHandler *messages.Handler,
	reactionsHandler *reactions.Handler,
	notificationsHandler *notifications.Handler,
	fileHandler *files.Handler,
	wsHandler *websocket.WsHandler,
	authService *auth.Service,
) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://127.0.0.1:3000"},
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

	// Public routes (no auth required)
	registerPublicAuthRoutes(r, authHandler)

	// Protected routes (auth required)
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(authService))

		registerProtectedAuthRoutes(r, authHandler)
		registerUserRoutes(r, usersHandler)
		registerConversationRoutes(r, conversationsHandler, messagesHandler, reactionsHandler)
		registerNotificationRoutes(r, notificationsHandler)
		registerFileRoutes(r, fileHandler)
	})

	// Swagger documentation
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	// WebSocket endpoint
	r.Get("/ws", wsHandler.ServeWS)

	return r
}
