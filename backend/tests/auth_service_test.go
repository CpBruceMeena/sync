package tests

import (
	"testing"
	"time"

	"github.com/CpBruceMeena/sync/internal/auth"
	"github.com/google/uuid"
)

func TestAuthService_GenerateAndValidateTokens(t *testing.T) {
	s := auth.NewService("test-secret-key", 15, 7)

	userID := uuid.New()
	username := "testuser"

	tokens, err := s.GenerateTokens(userID, username)
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	if tokens.AccessToken == "" {
		t.Error("Expected access token to be non-empty")
	}
	if tokens.RefreshToken == "" {
		t.Error("Expected refresh token to be non-empty")
	}
	if tokens.ExpiresIn <= 0 {
		t.Errorf("Expected ExpiresIn > 0, got %d", tokens.ExpiresIn)
	}

	claims, err := s.ValidateAccessToken(tokens.AccessToken)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Expected UserID %v, got %v", userID, claims.UserID)
	}
	if claims.Username != username {
		t.Errorf("Expected Username %s, got %s", username, claims.Username)
	}
	if claims.Issuer != "sync" {
		t.Errorf("Expected Issuer 'go-chatsync', got '%s'", claims.Issuer)
	}
}

func TestAuthService_InvalidToken(t *testing.T) {
	s := auth.NewService("test-secret-key", 15, 7)

	_, err := s.ValidateAccessToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestAuthService_ExpiredToken(t *testing.T) {
	s := auth.NewService("test-secret-key", -1, 7) // negative TTL = already expired

	userID := uuid.New()
	tokens, err := s.GenerateTokens(userID, "testuser")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	// Wait a moment to ensure expiration
	time.Sleep(10 * time.Millisecond)

	_, err = s.ValidateAccessToken(tokens.AccessToken)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestAuthService_WrongSecret(t *testing.T) {
	s1 := auth.NewService("secret-1", 15, 7)
	s2 := auth.NewService("secret-2", 15, 7)

	userID := uuid.New()
	tokens, err := s1.GenerateTokens(userID, "testuser")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	_, err = s2.ValidateAccessToken(tokens.AccessToken)
	if err == nil {
		t.Error("Expected error for token signed with different secret, got nil")
	}
}

func TestAuthService_TokenDuration(t *testing.T) {
	s := auth.NewService("test-secret-key", 60, 7) // 60 min access TTL

	userID := uuid.New()
	tokens, err := s.GenerateTokens(userID, "testuser")
	if err != nil {
		t.Fatalf("GenerateTokens failed: %v", err)
	}

	expectedExpiry := 60 * 60 // 60 min in seconds
	if tokens.ExpiresIn != expectedExpiry {
		t.Errorf("Expected ExpiresIn %d, got %d", expectedExpiry, tokens.ExpiresIn)
	}
}

func TestAuthService_ConcurrentTokenGeneration(t *testing.T) {
	s := auth.NewService("test-secret-key", 15, 7)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			userID := uuid.New()
			_, err := s.GenerateTokens(userID, "user")
			if err != nil {
				t.Errorf("Concurrent GenerateTokens failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
