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
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/google/uuid"
)

// ---- Mock repository implementations ----

type mockUserRepo struct {
	getByEmailFn             func(ctx context.Context, email string) (*models.User, error)
	getByUsernameFn          func(ctx context.Context, username string) (*models.User, error)
	listFn                   func(ctx context.Context) ([]models.User, error)
	createFn                 func(ctx context.Context, user *models.User) error
	getByIDFn                func(ctx context.Context, id uuid.UUID) (*models.User, error)
	getByEmailWithPasswordFn func(ctx context.Context, email string) (*models.User, error)
	updateFn                 func(ctx context.Context, user *models.User) error
	updatePasswordFn         func(ctx context.Context, id uuid.UUID, passwordHash string) error
	updateStatusFn           func(ctx context.Context, id uuid.UUID, status string) error
	deleteFn                 func(ctx context.Context, id uuid.UUID) error
}

func (m *mockUserRepo) Create(ctx context.Context, user *models.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByEmailWithPassword(ctx context.Context, email string) (*models.User, error) {
	if m.getByEmailWithPasswordFn != nil {
		return m.getByEmailWithPasswordFn(ctx, email)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, nil
}
func (m *mockUserRepo) List(ctx context.Context) ([]models.User, error) {
	if m.listFn != nil {
		return m.listFn(ctx)
	}
	return nil, nil
}
func (m *mockUserRepo) Update(ctx context.Context, user *models.User) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, user)
	}
	return nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, id, passwordHash)
	}
	return nil
}
func (m *mockUserRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}
func (m *mockUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockConvRepo struct {
	createFn       func(ctx context.Context, conv *models.Conversation) error
	getByIDFn      func(ctx context.Context, id uuid.UUID) (*models.Conversation, error)
	listByUserIDFn func(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error)
	findPrivateFn  func(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error)
	addMemberFn    func(ctx context.Context, member *models.ConversationMember) error
	removeMemberFn func(ctx context.Context, convID, userID uuid.UUID) error
	getMembersFn   func(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error)
	isMemberFn     func(ctx context.Context, convID, userID uuid.UUID) (bool, error)
	deleteFn       func(ctx context.Context, id uuid.UUID) error
}

func (m *mockConvRepo) Create(ctx context.Context, conv *models.Conversation) error {
	if m.createFn != nil {
		return m.createFn(ctx, conv)
	}
	return nil
}
func (m *mockConvRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockConvRepo) ListByUserID(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
	if m.listByUserIDFn != nil {
		return m.listByUserIDFn(ctx, userID)
	}
	return nil, nil
}
func (m *mockConvRepo) FindPrivate(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
	if m.findPrivateFn != nil {
		return m.findPrivateFn(ctx, userID1, userID2)
	}
	return nil, nil
}
func (m *mockConvRepo) AddMember(ctx context.Context, member *models.ConversationMember) error {
	if m.addMemberFn != nil {
		return m.addMemberFn(ctx, member)
	}
	return nil
}
func (m *mockConvRepo) RemoveMember(ctx context.Context, convID, userID uuid.UUID) error {
	if m.removeMemberFn != nil {
		return m.removeMemberFn(ctx, convID, userID)
	}
	return nil
}
func (m *mockConvRepo) GetMembers(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error) {
	if m.getMembersFn != nil {
		return m.getMembersFn(ctx, convID)
	}
	return nil, nil
}
func (m *mockConvRepo) IsMember(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
	if m.isMemberFn != nil {
		return m.isMemberFn(ctx, convID, userID)
	}
	return false, nil
}
func (m *mockConvRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}

type mockMsgRepo struct {
	createFn         func(ctx context.Context, msg *models.Message) error
	getByIDFn        func(ctx context.Context, id uuid.UUID) (*models.Message, error)
	listByConvFn     func(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error)
	deleteFn         func(ctx context.Context, id, senderID uuid.UUID) error
	addReactionFn    func(ctx context.Context, reaction *models.Reaction) error
	removeReactionFn func(ctx context.Context, messageID, userID uuid.UUID, emoji string) error
}

func (m *mockMsgRepo) Create(ctx context.Context, msg *models.Message) error {
	if m.createFn != nil {
		return m.createFn(ctx, msg)
	}
	return nil
}
func (m *mockMsgRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Message, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}
func (m *mockMsgRepo) ListByConversation(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error) {
	if m.listByConvFn != nil {
		return m.listByConvFn(ctx, convID, cursor, limit)
	}
	return nil, nil
}
func (m *mockMsgRepo) Delete(ctx context.Context, id, senderID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, senderID)
	}
	return nil
}
func (m *mockMsgRepo) AddReaction(ctx context.Context, reaction *models.Reaction) error {
	if m.addReactionFn != nil {
		return m.addReactionFn(ctx, reaction)
	}
	return nil
}
func (m *mockMsgRepo) RemoveReaction(ctx context.Context, messageID, userID uuid.UUID, emoji string) error {
	if m.removeReactionFn != nil {
		return m.removeReactionFn(ctx, messageID, userID, emoji)
	}
	return nil
}

type mockSessionRepo struct {
	createFn         func(ctx context.Context, session *models.Session) error
	getByTokenFn     func(ctx context.Context, refreshToken string) (*models.Session, error)
	deleteFn         func(ctx context.Context, id uuid.UUID) error
	deleteByUserIDFn func(ctx context.Context, userID uuid.UUID) error
	cleanExpiredFn   func(ctx context.Context) error
}

func (m *mockSessionRepo) Create(ctx context.Context, session *models.Session) error {
	if m.createFn != nil {
		return m.createFn(ctx, session)
	}
	return nil
}
func (m *mockSessionRepo) GetByToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	if m.getByTokenFn != nil {
		return m.getByTokenFn(ctx, refreshToken)
	}
	return nil, nil
}
func (m *mockSessionRepo) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockSessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if m.deleteByUserIDFn != nil {
		return m.deleteByUserIDFn(ctx, userID)
	}
	return nil
}
func (m *mockSessionRepo) CleanExpired(ctx context.Context) error {
	if m.cleanExpiredFn != nil {
		return m.cleanExpiredFn(ctx)
	}
	return nil
}

type mockNotifRepo struct {
	createFn         func(ctx context.Context, notification *models.Notification) error
	listFn           func(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error)
	markReadFn       func(ctx context.Context, id, userID uuid.UUID) error
	markAllReadFn    func(ctx context.Context, userID uuid.UUID) error
	getUnreadCountFn func(ctx context.Context, userID uuid.UUID) (int64, error)
	deleteFn         func(ctx context.Context, id, userID uuid.UUID) error
}

func (m *mockNotifRepo) Create(ctx context.Context, notification *models.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, notification)
	}
	return nil
}
func (m *mockNotifRepo) List(ctx context.Context, userID uuid.UUID, limit int) ([]models.Notification, error) {
	if m.listFn != nil {
		return m.listFn(ctx, userID, limit)
	}
	return nil, nil
}
func (m *mockNotifRepo) MarkRead(ctx context.Context, id, userID uuid.UUID) error {
	if m.markReadFn != nil {
		return m.markReadFn(ctx, id, userID)
	}
	return nil
}
func (m *mockNotifRepo) MarkAllRead(ctx context.Context, userID uuid.UUID) error {
	if m.markAllReadFn != nil {
		return m.markAllReadFn(ctx, userID)
	}
	return nil
}
func (m *mockNotifRepo) GetUnreadCount(ctx context.Context, userID uuid.UUID) (int64, error) {
	if m.getUnreadCountFn != nil {
		return m.getUnreadCountFn(ctx, userID)
	}
	return 0, nil
}
func (m *mockNotifRepo) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id, userID)
	}
	return nil
}

// newMockRepos creates a *repository.Repositories with all mock repositories
func newMockRepos() *repository.Repositories {
	return &repository.Repositories{
		Users:         &mockUserRepo{},
		Conversations: &mockConvRepo{},
		Messages:      &mockMsgRepo{},
		Sessions:      &mockSessionRepo{},
		Notifications: &mockNotifRepo{},
	}
}

// Helper to create a test context with user_id
func authContext(userID uuid.UUID) context.Context {
	ctx := context.WithValue(context.Background(), "user_id", userID)
	return context.WithValue(ctx, "username", "testuser")
}

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
	h := users.NewHandler(repos)

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
	h := messages.NewHandler(newMockRepos())

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
	h := conversations.NewHandler(newMockRepos())

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
