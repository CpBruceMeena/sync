package tests

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/CpBruceMeena/sync/internal/middleware"
	"github.com/google/uuid"
)

func setupAuthMiddlewareTest(t *testing.T) (*auth.Service, string, uuid.UUID) {
	t.Helper()

	svc := auth.NewService("test-secret-key", 15, 7)
	userID := uuid.New()

	tokens, err := svc.GenerateTokens(userID, "testuser")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	return svc, tokens.AccessToken, userID
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	svc, token, userID := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID := r.Context().Value(middleware.UserIDKey).(uuid.UUID)
		if gotID != userID {
			t.Errorf("Expected userID %v, got %v", userID, gotID)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	svc, _, _ := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	svc, _, _ := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	svc, _, _ := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_WrongScheme(t *testing.T) {
	svc, token, _ := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Basic "+token)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestAuthMiddleware_EmptyToken(t *testing.T) {
	svc, _, _ := setupAuthMiddlewareTest(t)

	handler := middleware.AuthMiddleware(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/api/protected", nil)
	req.Header.Set("Authorization", "Bearer ")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}
