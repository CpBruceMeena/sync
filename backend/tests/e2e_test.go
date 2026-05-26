package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CpBruceMeena/sync/internal/api"
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/files"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/reactions"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

// mockE2ESuite holds all state for mock-based end-to-end tests.
// It uses in-memory maps to simulate database storage.
type mockE2ESuite struct {
	repos   *repository.Repositories
	authSvc *auth.Service
	mux     http.Handler

	// In-memory storage
	usersByID       map[uuid.UUID]*models.User
	usersByEmail    map[string]*models.User
	usersByUsername map[string]*models.User
	sessionsByToken map[string]*models.Session
	convsByID       map[uuid.UUID]*models.Conversation
	convMembers     map[string]*models.ConversationMember // key: "convID:userID"
	messagesByConv  map[string][]*models.Message          // key: convID.String()
}

// newMockE2ESuite wires up the entire application with stateful mock repositories.
func newMockE2ESuite() *mockE2ESuite {
	s := &mockE2ESuite{
		usersByID:       make(map[uuid.UUID]*models.User),
		usersByEmail:    make(map[string]*models.User),
		usersByUsername: make(map[string]*models.User),
		sessionsByToken: make(map[string]*models.Session),
		convsByID:       make(map[uuid.UUID]*models.Conversation),
		convMembers:     make(map[string]*models.ConversationMember),
		messagesByConv:  make(map[string][]*models.Message),
	}

	s.repos = newMockRepos()
	s.authSvc = auth.NewService("test-jwt-secret-for-e2e", 60, 7)

	// ---- Configure User Mocks ----
	mu := s.repos.Users.(*mockUserRepo)
	mu.createFn = func(ctx context.Context, user *models.User) error {
		if user.ID == uuid.Nil {
			user.ID = uuid.New()
		}
		now := time.Now()
		user.CreatedAt = now
		user.UpdatedAt = now
		s.usersByID[user.ID] = user
		s.usersByEmail[user.Email] = user
		s.usersByUsername[user.Username] = user
		return nil
	}
	mu.getByEmailFn = func(ctx context.Context, email string) (*models.User, error) {
		if u, ok := s.usersByEmail[email]; ok {
			return u, nil
		}
		return nil, errors.New("user not found")
	}
	mu.getByEmailWithPasswordFn = func(ctx context.Context, email string) (*models.User, error) {
		if u, ok := s.usersByEmail[email]; ok {
			return u, nil
		}
		return nil, errors.New("user not found")
	}
	mu.getByUsernameFn = func(ctx context.Context, username string) (*models.User, error) {
		if u, ok := s.usersByUsername[username]; ok {
			return u, nil
		}
		return nil, errors.New("user not found")
	}
	mu.getByIDFn = func(ctx context.Context, id uuid.UUID) (*models.User, error) {
		if u, ok := s.usersByID[id]; ok {
			return u, nil
		}
		return nil, errors.New("user not found")
	}
	mu.listFn = func(ctx context.Context) ([]models.User, error) {
		result := make([]models.User, 0, len(s.usersByID))
		for _, u := range s.usersByID {
			result = append(result, *u)
		}
		return result, nil
	}

	// ---- Configure Session Mocks ----
	ms := s.repos.Sessions.(*mockSessionRepo)
	ms.createFn = func(ctx context.Context, session *models.Session) error {
		if session.ID == uuid.Nil {
			session.ID = uuid.New()
		}
		session.CreatedAt = time.Now()
		s.sessionsByToken[session.RefreshToken] = session
		return nil
	}
	ms.getByTokenFn = func(ctx context.Context, refreshToken string) (*models.Session, error) {
		if sess, ok := s.sessionsByToken[refreshToken]; ok {
			return sess, nil
		}
		return nil, errors.New("session not found")
	}
	ms.deleteFn = func(ctx context.Context, id uuid.UUID) error {
		for token, sess := range s.sessionsByToken {
			if sess.ID == id {
				delete(s.sessionsByToken, token)
				return nil
			}
		}
		return nil
	}
	ms.deleteByUserIDFn = func(ctx context.Context, userID uuid.UUID) error {
		for token, sess := range s.sessionsByToken {
			if sess.UserID == userID {
				delete(s.sessionsByToken, token)
			}
		}
		return nil
	}

	// ---- Configure Conversation Mocks ----
	mc := s.repos.Conversations.(*mockConvRepo)
	mc.createFn = func(ctx context.Context, conv *models.Conversation) error {
		if conv.ID == uuid.Nil {
			conv.ID = uuid.New()
		}
		now := time.Now()
		conv.CreatedAt = now
		conv.UpdatedAt = now
		s.convsByID[conv.ID] = conv
		return nil
	}
	mc.getByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Conversation, error) {
		if c, ok := s.convsByID[id]; ok {
			return c, nil
		}
		return nil, errors.New("conversation not found")
	}
	mc.listByUserIDFn = func(ctx context.Context, userID uuid.UUID) ([]models.Conversation, error) {
		var result []models.Conversation
		for _, c := range s.convsByID {
			// Check if user is a member of this conversation
			key := c.ID.String() + ":" + userID.String()
			if _, ok := s.convMembers[key]; ok {
				result = append(result, *c)
			}
		}
		return result, nil
	}
	mc.findPrivateFn = func(ctx context.Context, userID1, userID2 uuid.UUID) (*models.Conversation, error) {
		for _, c := range s.convsByID {
			if c.Type != "private" {
				continue
			}
			key1 := c.ID.String() + ":" + userID1.String()
			key2 := c.ID.String() + ":" + userID2.String()
			if _, ok1 := s.convMembers[key1]; ok1 {
				if _, ok2 := s.convMembers[key2]; ok2 {
					return c, nil
				}
			}
		}
		return nil, errors.New("private conversation not found")
	}
	mc.addMemberFn = func(ctx context.Context, member *models.ConversationMember) error {
		if member.ID == uuid.Nil {
			member.ID = uuid.New()
		}
		member.JoinedAt = time.Now()
		key := member.ConversationID.String() + ":" + member.UserID.String()
		s.convMembers[key] = member
		return nil
	}
	mc.removeMemberFn = func(ctx context.Context, convID, userID uuid.UUID) error {
		key := convID.String() + ":" + userID.String()
		delete(s.convMembers, key)
		return nil
	}
	mc.getMembersFn = func(ctx context.Context, convID uuid.UUID) ([]models.ConversationMember, error) {
		var result []models.ConversationMember
		for key, member := range s.convMembers {
			// Parse convID from key "convID:userID"
			var cID uuid.UUID
			if err := cID.UnmarshalText([]byte(key[:36])); err == nil && cID == convID {
				result = append(result, *member)
			}
		}
		return result, nil
	}
	mc.isMemberFn = func(ctx context.Context, convID, userID uuid.UUID) (bool, error) {
		key := convID.String() + ":" + userID.String()
		_, ok := s.convMembers[key]
		return ok, nil
	}

	// ---- Configure Message Mocks ----
	mm := s.repos.Messages.(*mockMsgRepo)
	mm.createFn = func(ctx context.Context, msg *models.Message) error {
		if msg.ID == uuid.Nil {
			msg.ID = uuid.New()
		}
		msg.CreatedAt = time.Now()
		convKey := msg.ConversationID.String()
		s.messagesByConv[convKey] = append(s.messagesByConv[convKey], msg)
		return nil
	}
	mm.listByConvFn = func(ctx context.Context, convID uuid.UUID, cursor uuid.UUID, limit int) ([]models.Message, error) {
		convKey := convID.String()
		msgs := s.messagesByConv[convKey]
		if msgs == nil {
			return []models.Message{}, nil
		}
		// Return newest first (reverse order)
		result := make([]models.Message, 0, len(msgs))
		for i := len(msgs) - 1; i >= 0; i-- {
			result = append(result, *msgs[i])
		}
		// Handle cursor (skip messages after cursor)
		if cursor != uuid.Nil {
			cut := -1
			for i, m := range result {
				if m.ID == cursor {
					cut = i + 1
					break
				}
			}
			if cut >= 0 && cut < len(result) {
				result = result[cut:]
			}
		}
		if len(result) > limit {
			result = result[:limit]
		}
		return result, nil
	}
	mm.getByIDFn = func(ctx context.Context, id uuid.UUID) (*models.Message, error) {
		for _, msgs := range s.messagesByConv {
			for _, m := range msgs {
				if m.ID == id {
					return m, nil
				}
			}
		}
		return nil, errors.New("message not found")
	}

	// ---- Configure Notification Mocks (default no-op) ----

	// Wire up the app
	userSvc := users.NewService(s.repos)
	notifSvc := notifications.NewService(s.repos)
	messageSvc := messages.NewService(s.repos, notifSvc)
	conversationSvc := conversations.NewService(s.repos, notifSvc)

	authHandler := auth.NewHandler(s.authSvc, s.repos)
	usersHandler := users.NewHandler(userSvc)
	convsHandler := conversations.NewHandler(conversationSvc)
	msgsHandler := messages.NewHandler(messageSvc)
	notifHandler := notifications.NewHandler(notifSvc)
	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, s.authSvc, s.repos)

	reactionSvc := reactions.NewService(s.repos, wsHub, notifSvc)
	reactionsHandler := reactions.NewHandler(reactionSvc)

	// File uploads
	fileSvc := files.NewService(s.repos, "./uploads")
	fileHandler := files.NewHandler(fileSvc, "./uploads")

	s.mux = api.SetupRoutes(authHandler, usersHandler, convsHandler, msgsHandler, reactionsHandler, notifHandler, fileHandler, wsHandler, s.authSvc)

	return s
}

// registerUser helper: register, return token and user ID
func (s *mockE2ESuite) registerUser(t *testing.T, username, email, password string) (string, uuid.UUID) {
	t.Helper()
	resp := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": username,
		"email":    email,
		"password": password,
	}, "")
	require.Equal(t, http.StatusCreated, resp.Code, "register failed: %s", resp.Body.String())

	var authRes auth.AuthResponse
	err := json.Unmarshal(resp.Body.Bytes(), &authRes)
	require.NoError(t, err)
	return authRes.Token.AccessToken, authRes.User.ID
}

// doRequest performs an HTTP request against the test router
func (s *mockE2ESuite) doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, "http://test"+path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

// --- Test types for response parsing ---

type convTestResponse struct {
	ID                 uuid.UUID            `json:"id"`
	Type               string               `json:"type"`
	Name               string               `json:"name"`
	AdminID            *uuid.UUID           `json:"admin_id"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	Members            []memberTestResponse `json:"members,omitempty"`
	LastMessageContent *string              `json:"last_message_content,omitempty"`
	LastMessageAt      *time.Time           `json:"last_message_at,omitempty"`
}

type memberTestResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type msgTestResponse struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      string    `json:"created_at"`
	UpdatedAt      string    `json:"updated_at"`
}

// --- E2E Tests ---

func TestE2E_AuthRegisterLogin(t *testing.T) {
	s := newMockE2ESuite()

	// Register a new user
	registerResp := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "alice",
		"email":    "alice@test.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, registerResp.Code, "register failed: %s", registerResp.Body.String())

	var authRes auth.AuthResponse
	parseResponse(t, registerResp, &authRes)
	require.Equal(t, "alice", authRes.User.Username)
	require.Equal(t, "alice@test.com", authRes.User.Email)
	require.NotEmpty(t, authRes.Token.AccessToken)
	require.NotEmpty(t, authRes.Token.RefreshToken)

	aliceToken := authRes.Token.AccessToken
	aliceID := authRes.User.ID

	// Register another user
	registerResp2 := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "bob",
		"email":    "bob@test.com",
		"password": "password456",
	}, "")
	require.Equal(t, http.StatusCreated, registerResp2.Code, "register bob failed: %s", registerResp2.Body.String())

	var bobAuthRes auth.AuthResponse
	parseResponse(t, registerResp2, &bobAuthRes)

	// Login with alice
	loginResp := s.doRequest("POST", "/api/auth/login", map[string]string{
		"email":    "alice@test.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusOK, loginResp.Code, "login failed: %s", loginResp.Body.String())

	var loginRes auth.AuthResponse
	parseResponse(t, loginResp, &loginRes)
	require.Equal(t, aliceID, loginRes.User.ID)

	// Get current user (Me endpoint)
	meResp := s.doRequest("GET", "/api/auth/me", nil, aliceToken)
	require.Equal(t, http.StatusOK, meResp.Code, "me endpoint failed: %s", meResp.Body.String())

	var meRes auth.UserResponse
	parseResponse(t, meResp, &meRes)
	require.Equal(t, "alice", meRes.Username)
}

func TestE2E_AuthValidation(t *testing.T) {
	s := newMockE2ESuite()

	// Register with missing fields
	resp := s.doRequest("POST", "/api/auth/register", map[string]string{}, "")
	require.Equal(t, http.StatusBadRequest, resp.Code)

	// Register with short password
	resp = s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "u", "email": "u@u.com", "password": "123",
	}, "")
	require.Equal(t, http.StatusBadRequest, resp.Code)

	// Register duplicate email
	s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "alice", "email": "alice@test.com", "password": "password123",
	}, "")
	resp = s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "alice2", "email": "alice@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusConflict, resp.Code, "expected 409 for duplicate email")
}

func TestE2E_UsersList(t *testing.T) {
	s := newMockE2ESuite()

	// Register test users
	reg1 := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "user1", "email": "user1@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, reg1.Code)

	reg2 := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "user2", "email": "user2@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, reg2.Code)

	var authRes auth.AuthResponse
	parseResponse(t, reg1, &authRes)
	token := authRes.Token.AccessToken

	// List users
	listResp := s.doRequest("GET", "/api/users", nil, token)
	require.Equal(t, http.StatusOK, listResp.Code)

	var usersList []map[string]interface{}
	parseResponse(t, listResp, &usersList)
	require.GreaterOrEqual(t, len(usersList), 2, "expected at least 2 users")
}

func TestE2E_ConversationsAndMessages(t *testing.T) {
	s := newMockE2ESuite()

	// Register users
	aliceToken, _ := s.registerUser(t, "alice_e2e", "alice_e2e@test.com", "password123")
	bobToken, _ := s.registerUser(t, "bob_e2e", "bob_e2e@test.com", "password123")
	charlieToken, _ := s.registerUser(t, "charlie_e2e", "charlie_e2e@test.com", "password123")

	t.Run("create private conversation", func(t *testing.T) {
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create private conv failed: %s", createResp.Body.String())

		var conv convTestResponse
		parseResponse(t, createResp, &conv)
		require.Equal(t, "private", conv.Type)

		// Verify it appears in Alice's conversation list
		listResp := s.doRequest("GET", "/api/conversations", nil, aliceToken)
		require.Equal(t, http.StatusOK, listResp.Code)
		var convs []convTestResponse
		parseResponse(t, listResp, &convs)
		require.GreaterOrEqual(t, len(convs), 1, "expected at least 1 conversation")

		// Verify it also appears in Bob's conversation list
		bobListResp := s.doRequest("GET", "/api/conversations", nil, bobToken)
		require.Equal(t, http.StatusOK, bobListResp.Code)
		var bobConvs []convTestResponse
		parseResponse(t, bobListResp, &bobConvs)
		require.GreaterOrEqual(t, len(bobConvs), 1, "expected Bob to see the conversation")
	})

	t.Run("send and list messages", func(t *testing.T) {
		// Register a unique user for this subtest
		uniqueToken, _ := s.registerUser(t, "unique_sender", "unique_sender@test.com", "password123")

		// Create a new unique private conversation
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, uniqueToken)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var conv convTestResponse
		parseResponse(t, createResp, &conv)
		convID := conv.ID.String()

		// Alice sends a message
		sendResp := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "Hello Bob!",
		}, uniqueToken)
		require.Equal(t, http.StatusCreated, sendResp.Code, "send message failed: %s", sendResp.Body.String())

		var msg msgTestResponse
		parseResponse(t, sendResp, &msg)
		require.Equal(t, "Hello Bob!", msg.Content)

		// Bob sends a reply
		sendResp2 := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "Hey Alice!",
		}, bobToken)
		require.Equal(t, http.StatusCreated, sendResp2.Code, "bob send failed: %s", sendResp2.Body.String())

		// List messages (verify both appear)
		listMsgResp := s.doRequest("GET", "/api/conversations/"+convID+"/messages", nil, uniqueToken)
		require.Equal(t, http.StatusOK, listMsgResp.Code, "list messages failed: %s", listMsgResp.Body.String())

		var msgs []msgTestResponse
		parseResponse(t, listMsgResp, &msgs)
		require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages")
		// Messages are returned newest-first, so msgs[1] is the first message
		require.Equal(t, "Hello Bob!", msgs[1].Content)
		require.Equal(t, "Hey Alice!", msgs[0].Content)
	})

	t.Run("create group conversation", func(t *testing.T) {
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "group",
			"name":    "Test Group",
			"members": []string{"bob_e2e", "charlie_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create group failed: %s", createResp.Body.String())

		var conv convTestResponse
		parseResponse(t, createResp, &conv)
		require.Equal(t, "group", conv.Type)
		require.Equal(t, "Test Group", conv.Name)

		// Verify group appears in all members' conversation lists
		for name, tok := range map[string]string{"alice": aliceToken, "bob": bobToken, "charlie": charlieToken} {
			listResp := s.doRequest("GET", "/api/conversations", nil, tok)
			require.Equal(t, http.StatusOK, listResp.Code)
			var convs []convTestResponse
			parseResponse(t, listResp, &convs)
			groupFound := false
			for _, c := range convs {
				if c.Name == "Test Group" && c.Type == "group" {
					groupFound = true
					break
				}
			}
			require.True(t, groupFound, "%s should see the group in their conversations", name)
		}

		// Send a group message
		sendResp := s.doRequest("POST", "/api/conversations/"+conv.ID.String()+"/messages", map[string]interface{}{
			"content": "Welcome to the group!",
		}, aliceToken)
		require.Equal(t, http.StatusCreated, sendResp.Code, "send group message failed: %s", sendResp.Body.String())

		// Bob sends a message in the group
		sendResp2 := s.doRequest("POST", "/api/conversations/"+conv.ID.String()+"/messages", map[string]interface{}{
			"content": "Thanks Alice!",
		}, bobToken)
		require.Equal(t, http.StatusCreated, sendResp2.Code)

		// List all messages in the group
		listMsgResp := s.doRequest("GET", "/api/conversations/"+conv.ID.String()+"/messages", nil, aliceToken)
		require.Equal(t, http.StatusOK, listMsgResp.Code)
		var msgs []msgTestResponse
		parseResponse(t, listMsgResp, &msgs)
		require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages in group")
	})

	t.Run("list messages validation", func(t *testing.T) {
		// Need a conversation to test against - use existing by creating one
		listToken, _ := s.registerUser(t, "list_validate", "list_validate@test.com", "password123")
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, listToken)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var conv convTestResponse
		parseResponse(t, createResp, &conv)
		convID := conv.ID.String()

		// Send empty content should fail
		sendResp := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "",
		}, listToken)
		require.Equal(t, http.StatusBadRequest, sendResp.Code)
	})
}

func TestE2E_UnauthenticatedAccess(t *testing.T) {
	s := newMockE2ESuite()

	// Access without token should fail
	endpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/users"},
		{"GET", "/api/conversations"},
		{"POST", "/api/conversations"},
		{"POST", "/api/auth/logout"},
		{"GET", "/api/auth/me"},
	}

	for _, ep := range endpoints {
		t.Run(ep.method+" "+ep.path, func(t *testing.T) {
			resp := s.doRequest(ep.method, ep.path, nil, "")
			require.Equal(t, http.StatusUnauthorized, resp.Code,
				"expected 401 for %s %s, got %d", ep.method, ep.path, resp.Code)
		})
	}
}

func TestE2E_RefreshToken(t *testing.T) {
	s := newMockE2ESuite()

	// Register
	regResp := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "refresh_user",
		"email":    "refresh@test.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, regResp.Code)

	var authRes auth.AuthResponse
	parseResponse(t, regResp, &authRes)
	refreshToken := authRes.Token.RefreshToken

	// Refresh token
	refreshResp := s.doRequest("POST", "/api/auth/refresh", map[string]string{
		"refresh_token": refreshToken,
	}, "")
	require.Equal(t, http.StatusOK, refreshResp.Code, "refresh failed: %s", refreshResp.Body.String())

	var refreshRes map[string]interface{}
	parseResponse(t, refreshResp, &refreshRes)
	_, ok := refreshRes["token"]
	require.True(t, ok, "expected token in refresh response")
}

func TestE2E_Logout(t *testing.T) {
	s := newMockE2ESuite()

	regResp := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "logout_user",
		"email":    "logout@test.com",
		"password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, regResp.Code)

	var authRes auth.AuthResponse
	parseResponse(t, regResp, &authRes)
	token := authRes.Token.AccessToken

	// Logout
	logoutResp := s.doRequest("POST", "/api/auth/logout", nil, token)
	require.Equal(t, http.StatusOK, logoutResp.Code)

	// After logout, me endpoint should still work (access token is still valid)
	// since logout only invalidates refresh tokens
	meResp := s.doRequest("GET", "/api/auth/me", nil, token)
	require.Equal(t, http.StatusOK, meResp.Code)
}

// parseResponse helper
func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()
	err := json.Unmarshal(rec.Body.Bytes(), dest)
	require.NoError(t, err, "failed to parse response body: %s", rec.Body.String())
}
