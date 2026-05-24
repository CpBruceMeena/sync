package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CpBruceMeena/sync/internal/api"
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/config"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/CpBruceMeena/sync/internal/websocket"
)

func main() {
	cfg := config.Load()

	db, err := database.NewDB(cfg.DSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Auto-migrate schema
	if err := db.DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Conversation{},
		&models.ConversationMember{},
		&models.Message{},
		&models.Reaction{},
		&models.Attachment{},
		&models.Notification{},
		&models.Presence{},
		&models.TypingEvent{},
	); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	repos := repository.NewRepositories(db.DB)
	authService := auth.NewService(cfg.JWTSecret, cfg.AccessTTL, cfg.RefreshTTL)

	authHandler := auth.NewHandler(authService, repos)
	usersHandler := users.NewHandler(repos)
	conversationsHandler := conversations.NewHandler(repos)
	messagesHandler := messages.NewHandler(repos)

	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, authService, repos)

	// All routes are defined in internal/api/routes.go
	r := api.SetupRoutes(
		authHandler,
		usersHandler,
		conversationsHandler,
		messagesHandler,
		wsHandler,
		authService,
	)

	addr := fmt.Sprintf(":%d", cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
