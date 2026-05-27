package websocket

import (
	"sync"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// PresenceInfo holds user presence data for broadcasting
type PresenceInfo struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Status   string    `json:"status"`
}

// Message types sent between client and server
const (
	TypeNewMessage      = "new_message"
	TypeTyping          = "typing"
	TypeStopTyping      = "stop_typing"
	TypePresence        = "presence"
	TypeReadReceipt     = "read_receipt"
	TypeReactionAdded   = "reaction_added"
	TypeReactionRemoved = "reaction_removed"
	TypeError           = "error"
	TypeOnlineUsers     = "online_users"
	TypeNotification    = "notification"
)

// WsHandler handles WebSocket upgrade requests
type WsHandler struct {
	hub         *Hub
	authService *auth.Service
	repos       *repository.Repositories
}

// WSMessage is the WebSocket message format exchanged between clients and server
type WSMessage struct {
	Type           string      `json:"type"`
	ConversationID uuid.UUID   `json:"conversation_id,omitempty"`
	SenderID       uuid.UUID   `json:"sender_id,omitempty"`
	SenderUsername string      `json:"sender_username,omitempty"`
	Content        string      `json:"content,omitempty"`
	MessageID      uuid.UUID   `json:"message_id,omitempty"`
	UserID         uuid.UUID   `json:"user_id,omitempty"`
	Username       string      `json:"username,omitempty"`
	Status         string      `json:"status,omitempty"`
	IsTyping       bool        `json:"is_typing,omitempty"`
	Emoji          string      `json:"emoji,omitempty"`
	Error          string      `json:"error,omitempty"`
	Data           interface{} `json:"data,omitempty"`
}

// Client represents a connected WebSocket client
type Client struct {
	UserID   uuid.UUID
	Username string
	Status   string
	conn     *websocket.Conn
	Send     chan []byte
	Hub      *Hub
}

// Hub manages all connected WebSocket clients and room subscriptions
type Hub struct {
	clients         map[uuid.UUID]*Client
	rooms           map[uuid.UUID]map[uuid.UUID]*Client // conversationID -> clients
	register        chan *Client
	unregister      chan *Client
	mu              sync.RWMutex
	presenceRepo    repository.PresenceRepository
	messageReadRepo repository.MessageReadRepository
}
