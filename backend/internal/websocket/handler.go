package websocket

import (
	"context"
	"log"
	"net/http"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/repository"
)

func NewWsHandler(hub *Hub, authService *auth.Service, repos *repository.Repositories) *WsHandler {
	return &WsHandler{
		hub:         hub,
		authService: authService,
		repos:       repos,
	}
}

func (h *WsHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, `{"error":"Missing token"}`, http.StatusUnauthorized)
		return
	}

	claims, err := h.authService.ValidateAccessToken(token)
	if err != nil {
		http.Error(w, `{"error":"Invalid token"}`, http.StatusUnauthorized)
		return
	}

	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection for user %s: %v", claims.Username, err)
		return
	}

	client := &Client{
		UserID:   claims.UserID,
		Username: claims.Username,
		conn:     conn,
		Send:     make(chan []byte, 256),
		Hub:      h.hub,
	}

	h.hub.RegisterClient(client)

	log.Printf("WebSocket connected: user=%s (%s)", claims.Username, claims.UserID)

	// Subscribe to user's conversations
	go h.subscribeToConversations(client)

	go client.WritePump()
	go client.ReadPump()
}

func (h *WsHandler) subscribeToConversations(client *Client) {
	convs, err := h.repos.Conversations.ListByUserID(context.Background(), client.UserID)
	if err != nil {
		log.Printf("Error fetching conversations for user %s: %v", client.Username, err)
		return
	}

	for _, conv := range convs {
		h.hub.JoinRoom(conv.ID, client)
	}
}
