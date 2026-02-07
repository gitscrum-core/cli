package chat

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

func TestNewCmdChat(t *testing.T) {
	f := factory.New()
	cmd := NewCmdChat(f)

	if !strings.HasPrefix(cmd.Use, "chat") {
		t.Errorf("Use should start with 'chat', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestChatList_Success tests listing channels
func TestChatList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/discussions/channels" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":         "channel-1",
						"slug":         "general",
						"name":         "general",
						"description":  "General discussion",
						"unread_count": 5,
						"last_message": map[string]interface{}{
							"content":    "Hello team!",
							"created_at": "2026-02-07T10:00:00Z",
							"user":       map[string]interface{}{"name": "Alice"},
						},
					},
					{
						"uuid":         "channel-2",
						"slug":         "dev",
						"name":         "dev",
						"description":  "Development channel",
						"unread_count": 0,
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
	resp, err := client.Get("/discussions/channels")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Channel `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 channels, got %d", len(result.Data))
	}

	if result.Data[0].UnreadCount != 5 {
		t.Errorf("general unread count = %d, want 5", result.Data[0].UnreadCount)
	}
}

// TestChatView_Success tests viewing channel messages
func TestChatView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/messages") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":       "msg-1",
						"content":    "Hello everyone!",
						"created_at": "2026-02-07T10:00:00Z",
						"user":       map[string]interface{}{"name": "Alice"},
					},
					{
						"uuid":       "msg-2",
						"content":    "Hi Alice!",
						"created_at": "2026-02-07T10:05:00Z",
						"user":       map[string]interface{}{"name": "Bob"},
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
	resp, err := client.Get("/discussions/channels/general/messages")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Message `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(result.Data))
	}

	if result.Data[0].Content != "Hello everyone!" {
		t.Errorf("first message = %q, want 'Hello everyone!'", result.Data[0].Content)
	}
}

// TestChatSend_Success tests sending a message
func TestChatSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/messages") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":       "msg-new",
					"content":    "Test message",
					"created_at": "2026-02-07T11:00:00Z",
					"user":       map[string]interface{}{"name": "TestUser"},
				},
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/discussions/channels/general/messages", map[string]interface{}{
		"content": "Test message",
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

// TestChatUnread_Success tests getting unread messages
func TestChatUnread_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "unread=true") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"channel": map[string]interface{}{
							"name": "general",
						},
						"messages": []map[string]interface{}{
							{"content": "New message 1", "user": map[string]interface{}{"name": "Alice"}},
							{"content": "New message 2", "user": map[string]interface{}{"name": "Bob"}},
						},
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
	resp, err := client.Get("/discussions/channels?unread=true")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Channel  Channel   `json:"channel"`
			Messages []Message `json:"messages"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) == 0 {
		t.Fatal("expected unread messages")
	}

	if len(result.Data[0].Messages) != 2 {
		t.Errorf("messages = %d, want 2", len(result.Data[0].Messages))
	}
}
