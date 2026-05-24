package tests

import (
	"testing"
	"time"

	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/google/uuid"
)

// mockConn wraps a nil conn for testing purposes
type mockConn struct{}

func TestHub_NewHub(t *testing.T) {
	hub := websocket.NewHub()
	if hub == nil {
		t.Fatal("NewHub returned nil")
	}

	go hub.Run()
}

func TestHub_RegisterAndUnregisterClient(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	client := &websocket.Client{
		UserID:   uuid.New(),
		Username: "testuser",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	// Register client
	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	// Check client is registered
	if !hub.IsUserOnline(client.UserID) {
		t.Error("Expected client to be online after register")
	}

	// Unregister client
	hub.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)

	if hub.IsUserOnline(client.UserID) {
		t.Error("Expected client to be offline after unregister")
	}
}

func TestHub_JoinRoom(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	client1 := &websocket.Client{
		UserID:   uuid.New(),
		Username: "user1",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	convID := uuid.New()

	hub.RegisterClient(client1)
	hub.JoinRoom(convID, client1)
}

func TestHub_LeaveRoom(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	client := &websocket.Client{
		UserID:   uuid.New(),
		Username: "user1",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	convID := uuid.New()

	hub.RegisterClient(client)
	hub.JoinRoom(convID, client)
	hub.LeaveRoom(convID, client.UserID)
}

func TestHub_IsUserOnline(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	userID := uuid.New()

	// Should not be online
	if hub.IsUserOnline(userID) {
		t.Error("Expected user to be offline")
	}

	client := &websocket.Client{
		UserID:   userID,
		Username: "testuser",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	if !hub.IsUserOnline(userID) {
		t.Error("Expected user to be online after register")
	}
}

func TestHub_GetClient(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	userID := uuid.New()
	client := &websocket.Client{
		UserID:   userID,
		Username: "testuser",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)

	got := hub.GetClient(userID)
	if got == nil {
		t.Fatal("GetClient returned nil")
	}
	if got.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, got.UserID)
	}
	if got.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", got.Username)
	}
}

func TestHub_DuplicateRegister(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	userID := uuid.New()
	client1 := &websocket.Client{
		UserID:   userID,
		Username: "user1",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	client2 := &websocket.Client{
		UserID:   userID,
		Username: "user2",
		Send:     make(chan []byte, 256),
		Hub:      hub,
	}

	hub.RegisterClient(client1)
	hub.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)

	// Should replace with latest client
	got := hub.GetClient(userID)
	if got == nil {
		t.Fatal("GetClient returned nil")
	}
	if got.Username != "user2" {
		t.Errorf("Expected latest client Username 'user2', got '%s'", got.Username)
	}
}

func TestHub_MultipleClientsAndRooms(t *testing.T) {
	hub := websocket.NewHub()
	go hub.Run()

	// Create multiple clients
	clients := make([]*websocket.Client, 5)
	for i := 0; i < 5; i++ {
		clients[i] = &websocket.Client{
			UserID:   uuid.New(),
			Username: "user",
			Send:     make(chan []byte, 256),
			Hub:      hub,
		}
		hub.RegisterClient(clients[i])
	}
	time.Sleep(10 * time.Millisecond)

	// Create multiple rooms
	convID1 := uuid.New()
	convID2 := uuid.New()

	// Join different rooms
	hub.JoinRoom(convID1, clients[0])
	hub.JoinRoom(convID1, clients[1])
	hub.JoinRoom(convID1, clients[2])
	hub.JoinRoom(convID2, clients[3])
	hub.JoinRoom(convID2, clients[4])

	// Verify all clients are online
	for i, c := range clients {
		if !hub.IsUserOnline(c.UserID) {
			t.Errorf("Expected client %d to be online", i)
		}
	}

	// Remove a client
	hub.LeaveRoom(convID1, clients[0].UserID)
	hub.UnregisterClient(clients[0])
	time.Sleep(10 * time.Millisecond)

	if hub.IsUserOnline(clients[0].UserID) {
		t.Error("Expected unregistered client to be offline")
	}
}
