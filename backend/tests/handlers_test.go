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
	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/google/uuid"
)

// mockQueries implements database.Querier interface for testing
type mockQueries struct {
	usersFn          func() ([]database.ListUsersRow, error)
	getUserByIDFn    func(id uuid.UUID) (database.GetUserByIDRow, error)
	getUserByEmailFn func(email string) (database.GetUserByEmailRow, error)
	createUserFn     func(params database.CreateUserParams) (database.CreateUserRow, error)
}

func (m *mockQueries) AddConversationMember(ctx context.Context, arg database.AddConversationMemberParams) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (m *mockQueries) AddReaction(ctx context.Context, arg database.AddReactionParams) (uuid.UUID, error) {
	return uuid.New(), nil
}
func (m *mockQueries) CleanExpiredSessions(ctx context.Context) error { return nil }
func (m *mockQueries) CreateConversation(ctx context.Context, arg database.CreateConversationParams) (database.Conversation, error) {
	return database.Conversation{ID: uuid.New(), Type: arg.Type, Name: arg.Name, AdminID: arg.AdminID}, nil
}
func (m *mockQueries) CreateMessage(ctx context.Context, arg database.CreateMessageParams) (database.Message, error) {
	return database.Message{ID: uuid.New(), ConversationID: arg.ConversationID, SenderID: arg.SenderID, Content: arg.Content, Type: arg.Type}, nil
}
func (m *mockQueries) CreateNotification(ctx context.Context, arg database.CreateNotificationParams) (database.Notification, error) {
	return database.Notification{}, nil
}
func (m *mockQueries) CreateSession(ctx context.Context, arg database.CreateSessionParams) (database.Session, error) {
	return database.Session{ID: uuid.New(), UserID: arg.UserID, RefreshToken: arg.RefreshToken, ExpiresAt: arg.ExpiresAt}, nil
}
func (m *mockQueries) CreateUser(ctx context.Context, arg database.CreateUserParams) (database.CreateUserRow, error) {
	if m.createUserFn != nil {
		return m.createUserFn(arg)
	}
	return database.CreateUserRow{}, nil
}
func (m *mockQueries) DeleteConversation(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockQueries) DeleteMessage(ctx context.Context, arg database.DeleteMessageParams) error {
	return nil
}
func (m *mockQueries) DeleteNotification(ctx context.Context, arg database.DeleteNotificationParams) error {
	return nil
}
func (m *mockQueries) DeleteSession(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockQueries) DeleteUser(ctx context.Context, id uuid.UUID) error    { return nil }
func (m *mockQueries) DeleteUserSessions(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockQueries) FindPrivateConversation(ctx context.Context, arg database.FindPrivateConversationParams) (database.Conversation, error) {
	return database.Conversation{}, nil
}
func (m *mockQueries) GetConversationByID(ctx context.Context, id uuid.UUID) (database.GetConversationByIDRow, error) {
	return database.GetConversationByIDRow{}, nil
}
func (m *mockQueries) GetConversationMembers(ctx context.Context, conversationID uuid.UUID) ([]database.GetConversationMembersRow, error) {
	return nil, nil
}
func (m *mockQueries) GetMessageByID(ctx context.Context, id uuid.UUID) (database.GetMessageByIDRow, error) {
	return database.GetMessageByIDRow{}, nil
}
func (m *mockQueries) GetSessionByToken(ctx context.Context, refreshToken string) (database.GetSessionByTokenRow, error) {
	return database.GetSessionByTokenRow{}, nil
}
func (m *mockQueries) GetUnreadNotificationCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	return 0, nil
}
func (m *mockQueries) GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error) {
	if m.getUserByEmailFn != nil {
		return m.getUserByEmailFn(email)
	}
	return database.GetUserByEmailRow{}, nil
}
func (m *mockQueries) GetUserByEmailWithPassword(ctx context.Context, email string) (database.User, error) {
	return database.User{}, nil
}
func (m *mockQueries) GetUserByID(ctx context.Context, id uuid.UUID) (database.GetUserByIDRow, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(id)
	}
	return database.GetUserByIDRow{}, nil
}
func (m *mockQueries) GetUserByUsername(ctx context.Context, username string) (database.GetUserByUsernameRow, error) {
	return database.GetUserByUsernameRow{}, nil
}
func (m *mockQueries) IsConversationMember(ctx context.Context, arg database.IsConversationMemberParams) (bool, error) {
	return false, nil
}
func (m *mockQueries) ListMessagesByConversation(ctx context.Context, arg database.ListMessagesByConversationParams) ([]database.ListMessagesByConversationRow, error) {
	return nil, nil
}
func (m *mockQueries) ListNotifications(ctx context.Context, arg database.ListNotificationsParams) ([]database.Notification, error) {
	return nil, nil
}
func (m *mockQueries) ListUserConversations(ctx context.Context, userID uuid.UUID) ([]database.ListUserConversationsRow, error) {
	return nil, nil
}
func (m *mockQueries) ListUsers(ctx context.Context) ([]database.ListUsersRow, error) {
	if m.usersFn != nil {
		return m.usersFn()
	}
	return nil, nil
}
func (m *mockQueries) MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	return nil
}
func (m *mockQueries) MarkNotificationRead(ctx context.Context, arg database.MarkNotificationReadParams) error {
	return nil
}
func (m *mockQueries) RemoveConversationMember(ctx context.Context, arg database.RemoveConversationMemberParams) error {
	return nil
}
func (m *mockQueries) RemoveReaction(ctx context.Context, arg database.RemoveReactionParams) error {
	return nil
}
func (m *mockQueries) UpdateUser(ctx context.Context, arg database.UpdateUserParams) (database.UpdateUserRow, error) {
	return database.UpdateUserRow{}, nil
}
func (m *mockQueries) UpdateUserPassword(ctx context.Context, arg database.UpdateUserPasswordParams) error {
	return nil
}
func (m *mockQueries) UpdateUserStatus(ctx context.Context, arg database.UpdateUserStatusParams) error {
	return nil
}

// Helper to create a test context with user_id
func authContext(userID uuid.UUID) context.Context {
	ctx := context.WithValue(context.Background(), "user_id", userID)
	return context.WithValue(ctx, "username", "testuser")
}

// --- Auth Handler Tests ---

type testDB struct {
	Pool    interface{}
	Queries *mockQueries
}

func TestAuthHandler_RegisterValidation(t *testing.T) {
	authSvc := auth.NewService("test-secret", 15, 7)
	h := auth.NewHandler(authSvc, nil, &mockQueries{})

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
	mq := &mockQueries{
		getUserByEmailFn: func(email string) (database.GetUserByEmailRow, error) {
			return database.GetUserByEmailRow{Email: email}, nil
		},
	}
	h := auth.NewHandler(authSvc, nil, mq)

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
	mq := &mockQueries{
		usersFn: func() ([]database.ListUsersRow, error) {
			return []database.ListUsersRow{
				{ID: uuid.New(), Username: "user1", Email: "user1@test.com", DisplayName: "User 1", AvatarUrl: "", Status: "online"},
				{ID: uuid.New(), Username: "user2", Email: "user2@test.com", DisplayName: "User 2", AvatarUrl: "", Status: "offline"},
			}, nil
		},
	}
	h := users.NewHandler(mq)

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
	mq := &mockQueries{}
	h := messages.NewHandler(mq)

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
	mq := &mockQueries{}
	h := conversations.NewHandler(mq)

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
