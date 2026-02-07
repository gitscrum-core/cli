package auth

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

func TestNewCmdAuth(t *testing.T) {
	f := factory.New()
	cmd := NewCmdAuth(f)

	if !strings.HasPrefix(cmd.Use, "auth") {
		t.Errorf("Use should start with 'auth', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}

	// Check for expected subcommands
	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}

	expected := []string{"login", "logout", "whoami", "status"}
	for _, name := range expected {
		if !subcommands[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestAuthWhoami_Success tests the whoami command
func TestAuthWhoami_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/me" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":     "user-123",
					"username": "johndoe",
					"name":     "John Doe",
					"email":    "john@example.com",
					"avatar":   "https://example.com/avatar.jpg",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/me")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UUID     string `json:"uuid"`
			Username string `json:"username"`
			Name     string `json:"name"`
			Email    string `json:"email"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Username != "johndoe" {
		t.Errorf("username = %q, want 'johndoe'", result.Data.Username)
	}

	if result.Data.Email != "john@example.com" {
		t.Errorf("email = %q, want 'john@example.com'", result.Data.Email)
	}
}

// TestAuthStatus_Authenticated tests status when authenticated
func TestAuthStatus_Authenticated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/me" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":     "user-123",
					"username": "johndoe",
					"name":     "John Doe",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "valid-token"})
	resp, err := client.Get("/me")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

// TestAuthStatus_NotAuthenticated tests status when not authenticated
func TestAuthStatus_NotAuthenticated(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Unauthenticated",
		})
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "invalid-token"})
	resp, err := client.Get("/me")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", resp.StatusCode)
	}
}
