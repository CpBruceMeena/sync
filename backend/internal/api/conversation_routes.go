package api

import (
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/reactions"
	"github.com/go-chi/chi/v5"
)

// registerConversationRoutes sets up conversation, message, and reaction routes
func registerConversationRoutes(r chi.Router, ch *conversations.Handler, mh *messages.Handler, rh *reactions.Handler) {
	r.Get("/api/conversations", ch.ListConversations)
	r.Post("/api/conversations", ch.CreateConversation)
	r.Get("/api/conversations/{id}", ch.GetConversation)
	r.Post("/api/conversations/{id}/members", ch.AddMember)
	r.Delete("/api/conversations/{id}/members/{userId}", ch.RemoveMember)
	r.Get("/api/conversations/{id}/messages", mh.ListMessages)
	r.Post("/api/conversations/{id}/messages", mh.SendMessage)
	r.Get("/api/conversations/{id}/search", mh.SearchMessages)
	r.Delete("/api/messages/{id}", mh.DeleteMessage)
	r.Post("/api/messages/{id}/reactions", rh.ToggleReaction)
}
