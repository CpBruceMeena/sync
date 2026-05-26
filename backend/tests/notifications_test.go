package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CpBruceMeena/sync/internal/notifications"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// --- Notification Handler Tests ---

func TestNotificationsHandler_ListNotifications(t *testing.T) {
	svc := notifications.NewService(newMockRepos())
	h := notifications.NewHandler(svc)

	req := httptest.NewRequest("GET", "/api/notifications", nil)
	req = req.WithContext(authContext(uuid.New()))
	rec := httptest.NewRecorder()

	h.ListNotifications(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var notifs []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &notifs); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
}

func TestNotificationsHandler_GetUnreadCount(t *testing.T) {
	svc := notifications.NewService(newMockRepos())
	h := notifications.NewHandler(svc)

	req := httptest.NewRequest("GET", "/api/notifications/unread-count", nil)
	req = req.WithContext(authContext(uuid.New()))
	rec := httptest.NewRecorder()

	h.GetUnreadCount(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if count, ok := resp["count"]; !ok {
		t.Errorf("Expected count in response, got %v", resp)
	} else if count.(float64) != 0 {
		t.Errorf("Expected count 0, got %v", count)
	}
}

func TestNotificationsHandler_MarkRead(t *testing.T) {
	svc := notifications.NewService(newMockRepos())
	h := notifications.NewHandler(svc)

	req := httptest.NewRequest("PUT", "/api/notifications/"+uuid.New().String()+"/read", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", uuid.New().String())
	req = req.WithContext(context.WithValue(authContext(uuid.New()), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestNotificationsHandler_MarkAllRead(t *testing.T) {
	svc := notifications.NewService(newMockRepos())
	h := notifications.NewHandler(svc)

	req := httptest.NewRequest("PUT", "/api/notifications/read-all", nil)
	req = req.WithContext(authContext(uuid.New()))
	rec := httptest.NewRecorder()

	h.MarkAllRead(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestNotificationsHandler_MarkReadInvalidID(t *testing.T) {
	svc := notifications.NewService(newMockRepos())
	h := notifications.NewHandler(svc)

	req := httptest.NewRequest("PUT", "/api/notifications/invalid/read", nil)
	req = req.WithContext(authContext(uuid.New()))
	rec := httptest.NewRecorder()

	h.MarkRead(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
