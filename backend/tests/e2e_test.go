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

	"github.com/CpBruceMeena/sync/internal/api"
	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/conversations"
	"github.com/CpBruceMeena/sync/internal/database"
	"github.com/CpBruceMeena/sync/internal/messages"
	"github.com/CpBruceMeena/sync/internal/models"
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

	authHandler := auth.NewHandler(s.authSvc, s.repos)
	usersHandler := users.NewHandler(s.repos)
	convsHandler := conversations.NewHandler(s.repos)
	msgsHandler := messages.NewHandler(s.repos)
	wsHub := websocket.NewHub()
	go wsHub.Run()
	wsHandler := websocket.NewWsHandler(wsHub, s.authSvc, s.repos)

	s.mux = api.SetupRoutes(authHandler, usersHandler, convsHandler, msgsHandler, wsHandler, s.authSvc)
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

	t.Run("create private conversation", func(t *testing.T) {
		// Alice creates a private conversation with Bob
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create private conv failed: %s", createResp.Body.String())

		var conv map[string]interface{}
		parseResponse(t, createResp, &conv)
		require.Equal(t, "private", conv["type"])

		// Verify it appears in Alice's conversation list
		listResp := s.doRequest("GET", "/api/conversations", nil, aliceToken)
		require.Equal(t, http.StatusOK, listResp.Code)
		var convs []map[string]interface{}
		parseResponse(t, listResp, &convs)
		require.GreaterOrEqual(t, len(convs), 1, "expected at least 1 conversation")
	})

	t.Run("send and list messages", func(t *testing.T) {
		// First create a conversation
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "private",
			"members": []string{"bob_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code)
		var conv map[string]interface{}
		parseResponse(t, createResp, &conv)
		convID := conv["id"].(string)

		// Alice sends a message
		sendResp := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "Hello Bob!",
		}, aliceToken)
		require.Equal(t, http.StatusCreated, sendResp.Code, "send message failed: %s", sendResp.Body.String())

		var msg map[string]interface{}
		parseResponse(t, sendResp, &msg)
		require.Equal(t, "Hello Bob!", msg["content"])

		// Bob sends a reply
		sendResp2 := s.doRequest("POST", "/api/conversations/"+convID+"/messages", map[string]interface{}{
			"content": "Hey Alice!",
		}, bobToken)
		require.Equal(t, http.StatusCreated, sendResp2.Code, "bob send failed: %s", sendResp2.Body.String())

		// List messages
		listMsgResp := s.doRequest("GET", "/api/conversations/"+convID+"/messages", nil, aliceToken)
		require.Equal(t, http.StatusOK, listMsgResp.Code, "list messages failed: %s", listMsgResp.Body.String())

		var msgs []map[string]interface{}
		parseResponse(t, listMsgResp, &msgs)
		require.GreaterOrEqual(t, len(msgs), 2, "expected at least 2 messages")
	})

	t.Run("create group conversation", func(t *testing.T) {
		createResp := s.doRequest("POST", "/api/conversations", map[string]interface{}{
			"type":    "group",
			"name":    "Test Group",
			"members": []string{"bob_e2e"},
		}, aliceToken)
		require.Equal(t, http.StatusCreated, createResp.Code, "create group failed: %s", createResp.Body.String())

		var conv map[string]interface{}
		parseResponse(t, createResp, &conv)
		require.Equal(t, "group", conv["type"])
		require.Equal(t, "Test Group", conv["name"])
	})

	t.Run("list messages validation", func(t *testing.T) {
		// Send empty content should fail
		sendResp := s.doRequest("POST", "/api/conversations/"+uuid.New().String()+"/messages", map[string]interface{}{
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
