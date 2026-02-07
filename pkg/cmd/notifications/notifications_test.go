package notifications

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

func TestNewCmdNotifications(t *testing.T) {
	f := factory.New()
	cmd := NewCmdNotifications(f)

	if !strings.HasPrefix(cmd.Use, "notifications") {
		t.Errorf("Use should start with 'notifications', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestNotificationsList_Success tests listing notifications
func TestNotificationsList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/notifications" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":       "notif-1",
						"type":       "task_assigned",
						"title":      "Task assigned to you",
						"body":       "GS-123: Fix login bug",
						"read_at":    nil,
						"created_at": "2026-02-07T10:00:00Z",
						"data": map[string]interface{}{
							"task_code":    "GS-123",
							"project_slug": "web-app",
						},
					},
					{
						"uuid":       "notif-2",
						"type":       "comment",
						"title":      "New comment",
						"body":       "Alice commented on your task",
						"read_at":    "2026-02-07T09:00:00Z",
						"created_at": "2026-02-06T15:00:00Z",
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/notifications")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Notification `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(result.Data))
	}

	if result.Data[0].ReadAt != "" {
		t.Error("first notification should be unread")
	}

	if result.Data[1].ReadAt == "" {
		t.Error("second notification should be read")
	}
}

// TestNotificationsUnread_Success tests listing unread notifications
func TestNotificationsUnread_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "unread=true") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":    "notif-1",
						"title":   "Unread notification",
						"read_at": nil,
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/notifications?unread=true")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Notification `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 1 {
		t.Fatalf("expected 1 unread notification, got %d", len(result.Data))
	}
}

// TestNotificationRead_Success tests marking a notification as read
func TestNotificationRead_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/read") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":    "notif-1",
					"read_at": "2026-02-07T11:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/notifications/notif-1/read", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// TestNotificationReadAll_Success tests marking all notifications as read
func TestNotificationReadAll_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/notifications/read-all" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"updated": 5,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/notifications/read-all", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// TestNotificationsClear_Success tests clearing all notifications
func TestNotificationsClear_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete && r.URL.Path == "/notifications" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Delete("/notifications")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("status = %d, want 204", resp.StatusCode)
	}
}

// TestSearch_Success tests global search
func TestSearch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"tasks": []map[string]interface{}{
						{"code": "GS-123", "title": "Fix login"},
					},
					"projects": []map[string]interface{}{
						{"slug": "web-app", "name": "Web Application"},
					},
					"wiki": []map[string]interface{}{
						{"slug": "getting-started", "title": "Getting Started"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/search?q=login")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Tasks    []struct{ Code string } `json:"tasks"`
			Projects []struct{ Slug string } `json:"projects"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data.Tasks) == 0 {
		t.Error("expected tasks in search results")
	}
}
