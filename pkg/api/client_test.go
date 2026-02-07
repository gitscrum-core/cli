package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitscrum-core/cli/pkg/auth"
)

func TestNewClient(t *testing.T) {
	token := &auth.Token{AccessToken: "test-token"}
	client := NewClient("https://api.example.com", token)

	if client.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL = %q, want %q", client.BaseURL, "https://api.example.com")
	}
	if client.Token != token {
		t.Error("Token not set correctly")
	}
	if client.HTTPClient == nil {
		t.Error("HTTPClient should not be nil")
	}
}

func TestClient_Request_WithAuth(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	token := &auth.Token{AccessToken: "my-secret-token"}
	client := NewClient(server.URL, token)

	_, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedAuth := "Bearer my-secret-token"
	if capturedAuth != expectedAuth {
		t.Errorf("Authorization header = %q, want %q", capturedAuth, expectedAuth)
	}
}

func TestClient_Request_WithoutAuth(t *testing.T) {
	var capturedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedAuth != "" {
		t.Errorf("expected no Authorization header, got %q", capturedAuth)
	}
}

func TestClient_Request_Headers(t *testing.T) {
	var headers http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.Get("/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		header   string
		expected string
	}{
		{"Content-Type", "application/json"},
		{"Accept", "application/json"},
		{"User-Agent", "GitScrum-CLI/1.0"},
	}

	for _, tt := range tests {
		if got := headers.Get(tt.header); got != tt.expected {
			t.Errorf("%s = %q, want %q", tt.header, got, tt.expected)
		}
	}
}

func TestClient_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/tasks" {
			t.Errorf("Path = %s, want /tasks", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"data": "test"})
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	resp, err := client.Get("/tasks")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want 200", resp.StatusCode)
	}
}

func TestClient_Post(t *testing.T) {
	var receivedBody map[string]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &receivedBody)
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.Post("/tasks", map[string]string{"title": "Test Task"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody["title"] != "Test Task" {
		t.Errorf("body title = %q, want %q", receivedBody["title"], "Test Task")
	}
}

func TestClient_Put(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Method = %s, want PUT", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.Put("/tasks/123", map[string]string{"title": "Updated"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Patch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("Method = %s, want PATCH", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	_, err := client.Patch("/tasks/123", map[string]string{"status": "done"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	resp, err := client.Delete("/tasks/123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("StatusCode = %d, want 204", resp.StatusCode)
	}
}

func TestDecodeResponse_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]string{
				"id":    "123",
				"title": "Test Task",
			},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, nil)
	resp, _ := client.Get("/test")

	var result struct {
		Data struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"data"`
	}

	err := DecodeResponse(resp, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Data.ID != "123" {
		t.Errorf("data.id = %q, want %q", result.Data.ID, "123")
	}
	if result.Data.Title != "Test Task" {
		t.Errorf("data.title = %q, want %q", result.Data.Title, "Test Task")
	}
}

func TestDecodeResponse_Error(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
	}{
		{"bad request", http.StatusBadRequest, `{"message": "invalid request"}`},
		{"unauthorized", http.StatusUnauthorized, `{"message": "unauthorized"}`},
		{"not found", http.StatusNotFound, `{"message": "not found"}`},
		{"server error", http.StatusInternalServerError, `{"message": "server error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.body))
			}))
			defer server.Close()

			client := NewClient(server.URL, nil)
			resp, _ := client.Get("/test")

			var result map[string]interface{}
			err := DecodeResponse(resp, &result)

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), "API error") {
				t.Errorf("error = %q, should contain 'API error'", err.Error())
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{
		StatusCode: 400,
		Message:    "validation failed",
	}

	expected := "API error 400: validation failed"
	if err.Error() != expected {
		t.Errorf("Error() = %q, want %q", err.Error(), expected)
	}
}
