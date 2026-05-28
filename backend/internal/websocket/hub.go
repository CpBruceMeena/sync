package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/google/uuid"
)

func NewHub(presenceRepo repository.PresenceRepository, messageReadRepo repository.MessageReadRepository) *Hub {
	return &Hub{
		clients:         make(map[uuid.UUID]*Client),
		rooms:           make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:        make(chan *Client, 256),
		unregister:      make(chan *Client, 256),
		presenceRepo:    presenceRepo,
		messageReadRepo: messageReadRepo,
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			client.Status = "online"
			h.clients[client.UserID] = client
			h.mu.Unlock()
			// Persist presence
			if err := h.presenceRepo.Upsert(context.Background(), &models.Presence{
				UserID:     client.UserID,
				Status:     "online",
				LastSeenAt: time.Now(),
			}); err != nil {
				// Log but don't block connection
			}
			h.BroadcastOnlineUsers()
			h.BroadcastPresence(client.UserID, client.Username, "online")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				// Remove from all rooms
				for _, room := range h.rooms {
					delete(room, client.UserID)
				}
			}
			h.mu.Unlock()
			// Persist offline status
			if err := h.presenceRepo.Upsert(context.Background(), &models.Presence{
				UserID:     client.UserID,
				Status:     "offline",
				LastSeenAt: time.Now(),
			}); err != nil {
				// Log but don't block
			}
			h.BroadcastOnlineUsers()
			h.BroadcastPresence(client.UserID, client.Username, "offline")
		}
	}
}

func (h *Hub) BroadcastOnlineUsers() {
	h.mu.RLock()
	onlineInfo := make([]PresenceInfo, 0, len(h.clients))
	for _, client := range h.clients {
		onlineInfo = append(onlineInfo, PresenceInfo{
			UserID:   client.UserID,
			Username: client.Username,
			Status:   client.Status,
		})
	}
	h.mu.RUnlock()

	msg := WSMessage{
		Type: TypeOnlineUsers,
		Data: onlineInfo,
	}
	data, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) BroadcastPresence(userID uuid.UUID, username, status string) {
	msg := WSMessage{
		Type:     TypePresence,
		UserID:   userID,
		Username: username,
		Status:   status,
	}
	data, _ := json.Marshal(msg)

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, client := range h.clients {
		if client.UserID != userID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

func (h *Hub) SetUserStatus(client *Client, status string) {
	h.mu.Lock()
	client.Status = status
	h.mu.Unlock()

	if err := h.presenceRepo.Upsert(context.Background(), &models.Presence{
		UserID:     client.UserID,
		Status:     status,
		LastSeenAt: time.Now(),
	}); err != nil {
		// Log but don't block
	}

	h.BroadcastPresence(client.UserID, client.Username, status)
	h.BroadcastOnlineUsers()
}

func (h *Hub) GetOnlineUserStatuses() []PresenceInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]PresenceInfo, 0, len(h.clients))
	for _, client := range h.clients {
		result = append(result, PresenceInfo{
			UserID:   client.UserID,
			Username: client.Username,
			Status:   client.Status,
		})
	}
	return result
}

func (h *Hub) BroadcastTyping(conversationID uuid.UUID, senderID uuid.UUID, senderUsername string, isTyping bool) {
	msg := WSMessage{
		Type:           TypeTyping,
		ConversationID: conversationID,
		SenderID:       senderID,
		SenderUsername: senderUsername,
		IsTyping:       isTyping,
	}
	data, _ := json.Marshal(msg)
	h.BroadcastToRoom(conversationID, data, senderID)
}

func (h *Hub) RegisterClient(client *Client) {
	// Add to clients map synchronously so GetClient immediately returns the client.
	// This avoids a race between registration and message delivery.
	h.mu.Lock()
	client.Status = "online"
	h.clients[client.UserID] = client
	h.mu.Unlock()

	// Also send to register channel for async presence broadcasts (Run goroutine)
	select {
	case h.register <- client:
	default:
	}
}

func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

func (h *Hub) JoinRoom(conversationID uuid.UUID, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[conversationID] == nil {
		h.rooms[conversationID] = make(map[uuid.UUID]*Client)
	}
	h.rooms[conversationID][client.UserID] = client
}

// SubscribeUserToConversation subscribes a user's WebSocket client to a conversation room.
// Returns true if the user was found and subscribed, false otherwise.
func (h *Hub) SubscribeUserToConversation(convID uuid.UUID, userID uuid.UUID) bool {
	client := h.GetClient(userID)
	if client == nil {
		return false
	}
	h.JoinRoom(convID, client)
	return true
}

func (h *Hub) LeaveRoom(conversationID uuid.UUID, userID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[conversationID]; ok {
		delete(room, userID)
		if len(room) == 0 {
			delete(h.rooms, conversationID)
		}
	}
}

func (h *Hub) BroadcastToRoom(conversationID uuid.UUID, message []byte, senderID uuid.UUID) {
	h.mu.RLock()
	room := h.rooms[conversationID]
	// Clone the room under the lock to avoid concurrent map iteration panics
	clients := make([]*Client, 0, len(room))
	for _, client := range room {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		if client.UserID != senderID {
			select {
			case client.Send <- message:
			default:
			}
		}
	}
}

func (h *Hub) BroadcastToRoomAll(conversationID uuid.UUID, message []byte) {
	h.mu.RLock()
	room := h.rooms[conversationID]
	// Clone the room under the lock to avoid concurrent map iteration panics
	clients := make([]*Client, 0, len(room))
	for _, client := range room {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	if len(clients) == 0 {
		log.Printf("[WS] BroadcastToRoomAll: room %s has no subscribers — message was not delivered to anyone", conversationID)
		return
	}

	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
		}
	}
}

func (h *Hub) IsUserOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

func (h *Hub) GetClient(userID uuid.UUID) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

// SendMessageToUser sends a message to a specific user if they're online
func (h *Hub) SendMessageToUser(userID uuid.UUID, message []byte) {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if ok {
		select {
		case client.Send <- message:
		default:
		}
	}
}
