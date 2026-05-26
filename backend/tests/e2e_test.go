package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/CpBruceMeena/sync/internal/api"
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/CpBruceMeena/sync/internal/reactions"
	"github.com/CpBruceMeena/sync/internal/repository"
	"github.com/CpBruceMeena/sync/internal/users"
	"github.com/CpBruceMeena/sync/internal/websocket"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// e2eTestSuite holds all the state needed for end-to-end tests
type e2eTestSuite struct {
	db      *database.DB
	repos   *repository.Repositories
	authSvc *auth.Service
	mux     http.Handler
	baseURL string
}

// Test response types for verifying DB state

type ConversationTestResponse struct {
	ID                 uuid.UUID            `json:"id"`
	Type               string               `json:"type"`
	Name               string               `json:"name"`
	AdminID            *uuid.UUID           `json:"admin_id"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
	Members            []MemberTestResponse `json:"members,omitempty"`
	LastMessageContent *string              `json:"last_message_content,omitempty"`
	LastMessageAt      *time.Time           `json:"last_message_at,omitempty"`
}

type MemberTestResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type MessageTestResponse struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
	return cmd.Run() == nil
}

// setupE2E creates a test PostgreSQL container and wires up the application
func setupE2E(t *testing.T) *e2eTestSuite {
	t.Helper()

	ctx := context.Background()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		if !isDockerAvailable() {
			t.Skip("TEST_DATABASE_URL not set and Docker not available; skipping E2E tests")
		}

		pgContainer, err := postgres.Run(ctx,
			"postgres:16-alpine",
			postgres.WithDatabase("sync_test"),
			postgres.WithUsername("postgres"),
			postgres.WithPassword("postgres"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*1000000000)),
		)
		require.NoError(t, err, "failed to start postgres container")
		t.Cleanup(func() { pgContainer.Terminate(ctx) })

		dsn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		require.NoError(t, err, "failed to get connection string")
	}

	s := &e2eTestSuite{}

	var err error
	s.db, err = database.NewDB(dsn)
	require.NoError(t, err, "failed to connect to database")
	t.Cleanup(func() { s.db.Close() })

	// Run migrations
	err = s.db.DB.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Conversation{},
		&models.ConversationMember{},
		&models.Message{},
		&models.Reaction{},
		&models.Attachment{},
		&models.Notification{},
		&models.Presence{},
		&models.TypingEvent{},
	)
	require.NoError(t, err, "failed to run migrations")

	// Clean database before each test run
	s.db.DB.Exec("TRUNCATE TABLE users, sessions, conversations, conversation_members, messages, reactions, attachments, notifications, presence, typing_events CASCADE")

	// Wire up the app
	s.repos = repository.NewRepositories(s.db.DB)
	s.authSvc = auth.NewService("test-jwt-secret-for-e2e", 60, 7)

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

	s.mux = api.SetupRoutes(authHandler, usersHandler, convsHandler, msgsHandler, reactionsHandler, notifHandler, wsHandler, s.authSvc)
	s.baseURL = "http://test"

	return s
}

// doRequest performs an HTTP request against the test router
func (s *e2eTestSuite) doRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, s.baseURL+path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)
	return rec
}

// parseResponse decodes JSON response body
func parseResponse(t *testing.T, rec *httptest.ResponseRecorder, dest interface{}) {
	t.Helper()
	err := json.Unmarshal(rec.Body.Bytes(), dest)
	require.NoError(t, err, "failed to parse response body: %s", rec.Body.String())
}

// --- E2E Tests ---

func TestE2E_AuthRegisterLogin(t *testing.T) {
	s := setupE2E(t)

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
	s := setupE2E(t)

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
	s := setupE2E(t)

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
	s := setupE2E(t)

	// Register users
	regAlice := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "alice_e2e", "email": "alice_e2e@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, regAlice.Code)
	var aliceAuth auth.AuthResponse
	parseResponse(t, regAlice, &aliceAuth)
	aliceToken := aliceAuth.Token.AccessToken

	regBob := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "bob_e2e", "email": "bob_e2e@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, regBob.Code)
	var bobAuth auth.AuthResponse
	parseResponse(t, regBob, &bobAuth)
	bobToken := bobAuth.Token.AccessToken

	// Register a third user for group tests
	regCharlie := s.doRequest("POST", "/api/auth/register", map[string]string{
		"username": "charlie_e2e", "email": "charlie_e2e@test.com", "password": "password123",
	}, "")
	require.Equal(t, http.StatusCreated, regCharlie.Code)
	var charlieAuth auth.AuthResponse
	parseResponse(t, regCharlie, &charlieAuth)
	charlieToken := charlieAuth.Token.AccessToken

	t.Run("create private conversation and verify DB state", func(t *testing.T) {
		// Alice creates a private conversation with Bob
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create private conv failed: %s", createResp.Body.String())

		var conv ConversationTestResponse
		parseResponse(t, createResp, &conv)
		require.Equal(t, "private", conv.Type)

		// Verify it appears in Alice's conversation list
		listResp := s.doRequest("GET", "/api/conversations", nil, aliceToken)
		require.Equal(t, http.StatusOK, listResp.Code)
		var convs []ConversationTestResponse
		parseResponse(t, listResp, &convs)
		require.GreaterOrEqual(t, len(convs), 1, "expected at least 1 conversation")

		// Verify it also appears in Bob's conversation list
		bobListResp := s.doRequest("GET", "/api/conversations", nil, bobToken)
		require.Equal(t, http.StatusOK, bobListResp.Code)
		var bobConvs []ConversationTestResponse
		parseResponse(t, bobListResp, &bobConvs)
		require.GreaterOrEqual(t, len(bobConvs), 1, "expected Bob to see the conversation")
	})

	var existingConvID string
	t.Run("send and list messages with DB verification", func(t *testing.T) {
		// Use unique users for this subtest to avoid the private-conversation-already-exists case
		regUnique := s.doRequest("POST", "/api/auth/register", map[string]string{
			"username": "unique_sender", "email": "unique_sender@test.com", "password": "password123",
		}, "")
		require.Equal(t, http.StatusCreated, regUnique.Code)
		var uniqueAuth auth.AuthResponse
		parseResponse(t, regUnique, &uniqueAuth)
		uniqueToken := uniqueAuth.Token.AccessToken

		// Create a new unique private conversation
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, uniqueToken)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var conv ConversationTestResponse
		parseResponse(t, createResp, &conv)
		convID := conv.ID.String()
		existingConvID = convID

		// Alice (unique_sender) sends a message
		sendResp := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "Hello Bob!",
		}, uniqueToken)
		require.Equal(t, http.StatusCreated, sendResp.Code, "send message failed: %s", sendResp.Body.String())

		var msg MessageTestResponse
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

		var msgs []MessageTestResponse
		parseResponse(t, listMsgResp, &msgs)
		require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages")
		// Messages are returned newest-first, so msgs[1] is the first message
		require.Equal(t, "Hello Bob!", msgs[1].Content)
		require.Equal(t, "Hey Alice!", msgs[0].Content)
	})

	t.Run("create group conversation with DB verification", func(t *testing.T) {
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "group",
			"name":    "Test Group",
			"members": []string{"bob_e2e", "charlie_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create group failed: %s", createResp.Body.String())

		var conv ConversationTestResponse
		parseResponse(t, createResp, &conv)
		require.Equal(t, "group", conv.Type)
		require.Equal(t, "Test Group", conv.Name)

		// Verify group appears in all members' conversation lists
		for name, tok := range map[string]string{"alice": aliceToken, "bob": bobToken, "charlie": charlieToken} {
			listResp := s.doRequest("GET", "/api/conversations", nil, tok)
			require.Equal(t, http.StatusOK, listResp.Code)
			var convs []ConversationTestResponse
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
		var msgs []MessageTestResponse
		parseResponse(t, listMsgResp, &msgs)
		require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages in group")
	})

	t.Run("list messages validation", func(t *testing.T) {
		// Send empty content should fail
		sendResp := s.doRequest("POST", "/api/conversations/"+existingConvID+"/messages", map[string]interface{}{
			"content": "",
		}, aliceToken)
		require.Equal(t, http.StatusBadRequest, sendResp.Code)
	})
}

func TestE2E_UnauthenticatedAccess(t *testing.T) {
	s := setupE2E(t)

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
	s := setupE2E(t)

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
	s := setupE2E(t)

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
