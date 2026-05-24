package websocket

import (
	"encoding/json"

	"github.com/google/uuid"
)

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		rooms:      make(map[uuid.UUID]map[uuid.UUID]*Client),
		register:   make(chan *Client, 256),
		unregister: make(chan *Client, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			h.BroadcastOnlineUsers()

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
			h.BroadcastOnlineUsers()
		}
	}
}

func (h *Hub) BroadcastOnlineUsers() {
	h.mu.RLock()
	onlineUsers := make([]uuid.UUID, 0, len(h.clients))
	for userID := range h.clients {
		onlineUsers = append(onlineUsers, userID)
	}
	h.mu.RUnlock()

	msg := WSMessage{
		Type: TypeOnlineUsers,
		Data: onlineUsers,
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

func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
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
	h.mu.RUnlock()

	for _, client := range room {
		if client.UserID != senderID {
			select {
			case client.Send <- message:
			default:
			}
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
