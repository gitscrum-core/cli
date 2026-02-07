package workspaces

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

func TestNewCmdWorkspaces(t *testing.T) {
	f := factory.New()
	cmd := NewCmdWorkspaces(f)

	if !strings.HasPrefix(cmd.Use, "workspaces") {
		t.Errorf("Use should start with 'workspaces', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestWorkspacesList_Success tests listing workspaces
func TestWorkspacesList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/companies" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":           "ws-1",
						"slug":           "acme-inc",
						"name":           "ACME Inc",
						"description":    "Main workspace",
						"members_count":  5,
						"projects_count": 12,
					},
					{
						"uuid":           "ws-2",
						"slug":           "personal",
						"name":           "Personal",
						"description":    "Personal projects",
						"members_count":  1,
						"projects_count": 3,
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
	resp, err := client.Get("/companies")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Workspace `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(result.Data))
	}

	if result.Data[0].Name != "ACME Inc" {
		t.Errorf("first workspace name = %q, want 'ACME Inc'", result.Data[0].Name)
	}
}

// TestWorkspaceView_Success tests viewing a workspace
func TestWorkspaceView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":           "ws-1",
					"slug":           "acme-inc",
					"name":           "ACME Inc",
					"description":    "Main workspace for all projects",
					"members_count":  15,
					"projects_count": 25,
					"created_at":     "2025-01-01T00:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/companies/acme-inc")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data Workspace `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.MembersCount != 15 {
		t.Errorf("members = %d, want 15", result.Data.MembersCount)
	}
}

// TestWorkspaceStats_Success tests getting workspace statistics
func TestWorkspaceStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/stats") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"total_projects":      25,
					"active_projects":     18,
					"total_tasks":         450,
					"completed_tasks":     380,
					"total_members":       15,
					"active_members":      12,
					"total_hours_tracked": 1250.5,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/companies/acme-inc/stats")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalProjects     int     `json:"total_projects"`
			ActiveProjects    int     `json:"active_projects"`
			TotalTasks        int     `json:"total_tasks"`
			TotalHoursTracked float64 `json:"total_hours_tracked"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.TotalProjects != 25 {
		t.Errorf("total projects = %d, want 25", result.Data.TotalProjects)
	}

	if result.Data.TotalHoursTracked != 1250.5 {
		t.Errorf("hours tracked = %.1f, want 1250.5", result.Data.TotalHoursTracked)
	}
}

// TestWorkspaceMembers_Success tests listing workspace members
func TestWorkspaceMembers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/members") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":   "user-1",
						"name":   "Alice Johnson",
						"email":  "alice@acme.com",
						"role":   "admin",
						"avatar": "https://example.com/alice.jpg",
					},
					{
						"uuid":   "user-2",
						"name":   "Bob Smith",
						"email":  "bob@acme.com",
						"role":   "member",
						"avatar": "https://example.com/bob.jpg",
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
	resp, err := client.Get("/companies/acme-inc/members")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("members = %d, want 2", len(result.Data))
	}

	if result.Data[0].Role != "admin" {
		t.Errorf("first member role = %q, want 'admin'", result.Data[0].Role)
	}
}

// TestWorkspaceSwitch_Success tests switching workspace context
func TestWorkspaceSwitch_Success(t *testing.T) {
	f := factory.New()
	cmd := NewCmdWorkspaces(f)

	// Verify switch subcommand exists
	found := false
	for _, sub := range cmd.Commands() {
		if strings.HasPrefix(sub.Use, "switch") {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected 'switch' subcommand")
	}
}
