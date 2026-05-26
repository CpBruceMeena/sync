package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	messages "github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/google/uuid"
)

// --- Auth Handler Tests ---

func TestAuthHandler_RegisterValidation(t *testing.T) {
	authSvc := auth.NewService("test-secret", 15, 7)
	h := auth.NewHandler(authSvc, newMockRepos())

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"missing all fields", map[string]string{}, http.StatusBadRequest},
		{"missing password", map[string]string{"username": "u", "email": "u@u.com"}, http.StatusBadRequest},
		{"short password", map[string]string{"username": "u", "email": "u@u.com", "password": "123"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.Register(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestAuthHandler_RegisterDuplicateEmail(t *testing.T) {
	authSvc := auth.NewService("test-secret", 15, 7)
	repos := newMockRepos()
	repos.Users = &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{Email: email}, nil
		},
	}
	h := auth.NewHandler(authSvc, repos)

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"email":    "existing@test.com",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	h.Register(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", rec.Code)
	}
}

// --- User Handler Tests ---

func TestUsersHandler_ListUsers(t *testing.T) {
	repos := newMockRepos()
	repos.Users = &mockUserRepo{
		listFn: func(ctx context.Context) ([]models.User, error) {
			return []models.User{
				{ID: uuid.New(), Username: "user1", Email: "user1@test.com", DisplayName: "User 1", AvatarUrl: "", Status: "online"},
				{ID: uuid.New(), Username: "user2", Email: "user2@test.com", DisplayName: "User 2", AvatarUrl: "", Status: "offline"},
			}, nil
		},
	}
	h := users.NewHandler(users.NewService(repos))

	req := httptest.NewRequest("GET", "/api/users", nil)
	req = req.WithContext(authContext(uuid.New()))
	rec := httptest.NewRecorder()

	h.ListUsers(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &users); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
}

// --- Message Handler Tests ---

func TestMessagesHandler_SendMessageValidation(t *testing.T) {
	h := messages.NewHandler(messages.NewService(newMockRepos(), nil))

	tests := []struct {
		name       string
		body       map[string]string
		wantStatus int
	}{
		{"empty content", map[string]string{"content": ""}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/conversations/"+uuid.New().String()+"/messages", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(authContext(uuid.New()))
			rec := httptest.NewRecorder()

			h.SendMessage(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

// --- Conversation Handler Tests ---

func TestConversationsHandler_CreateConversationValidation(t *testing.T) {
	h := conversations.NewHandler(conversations.NewService(newMockRepos(), nil))

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
	}{
		{"missing type", map[string]interface{}{}, http.StatusBadRequest},
		{"invalid type", map[string]interface{}{"type": "invalid"}, http.StatusBadRequest},
		{"group without name", map[string]interface{}{"type": "group", "members": []string{"user1"}}, http.StatusBadRequest},
		{"private without members", map[string]interface{}{"type": "private"}, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/conversations", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(authContext(uuid.New()))
			rec := httptest.NewRecorder()

			h.CreateConversation(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("Expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}
