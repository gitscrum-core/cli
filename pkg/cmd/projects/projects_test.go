package projects

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

func TestNewCmdProjects(t *testing.T) {
	f := factory.New()
	cmd := NewCmdProjects(f)

	if !strings.HasPrefix(cmd.Use, "projects") {
		t.Errorf("Use should start with 'projects', got %q", cmd.Use)
	}

	// Verify command has subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestProjectsList_Success tests listing projects
func TestProjectsList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/projects" {
			t.Errorf("Path = %q, want /projects", r.URL.Path)
		}

		// Response matching actual ProjectResource format
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          1,
					"uuid":        "proj-uuid-1",
					"slug":        "mobile-app",
					"name":        "Mobile App",
					"code":        "MA",
					"description": "Mobile application project",
					"stats": map[string]interface{}{
						"total":    50,
						"open":     20,
						"progress": 15,
						"closed":   15,
					},
					"company": map[string]interface{}{
						"uuid": "comp-uuid-1",
						"slug": "acme-corp",
						"name": "ACME Corporation",
					},
					"status": map[string]interface{}{
						"code":  "active",
						"title": "Active",
					},
					"settings": map[string]interface{}{
						"use_timer":    true,
						"has_sprints":  true,
						"has_wiki":     true,
						"show_number":  true,
					},
				},
				{
					"id":          2,
					"uuid":        "proj-uuid-2",
					"slug":        "web-platform",
					"name":        "Web Platform",
					"code":        "WP",
					"description": "Web platform project",
					"stats": map[string]interface{}{
						"total":    120,
						"open":     45,
						"progress": 30,
						"closed":   45,
					},
					"company": map[string]interface{}{
						"uuid": "comp-uuid-1",
						"slug": "acme-corp",
						"name": "ACME Corporation",
					},
					"status": map[string]interface{}{
						"code":  "active",
						"title": "Active",
					},
				},
			},
			"meta": map[string]interface{}{
				"current_page": 1,
				"per_page":     15,
				"total":        2,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/projects")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID    int    `json:"id"`
			UUID  string `json:"uuid"`
			Slug  string `json:"slug"`
			Name  string `json:"name"`
			Code  string `json:"code"`
			Stats struct {
				Total  int `json:"total"`
				Open   int `json:"open"`
				Closed int `json:"closed"`
			} `json:"stats"`
			Company struct {
				Slug string `json:"slug"`
				Name string `json:"name"`
			} `json:"company"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(result.Data))
	}

	proj := result.Data[0]
	if proj.Slug != "mobile-app" {
		t.Errorf("slug = %q, want 'mobile-app'", proj.Slug)
	}
	if proj.Name != "Mobile App" {
		t.Errorf("name = %q, want 'Mobile App'", proj.Name)
	}
	if proj.Stats.Total != 50 {
		t.Errorf("stats.total = %d, want 50", proj.Stats.Total)
	}
	if proj.Company.Slug != "acme-corp" {
		t.Errorf("company.slug = %q, want 'acme-corp'", proj.Company.Slug)
	}
}

// TestProjectsView_Success tests viewing a single project
func TestProjectsView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/projects/") {
			t.Errorf("Path should start with /projects/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":          1,
				"uuid":        "proj-uuid-1",
				"slug":        "mobile-app",
				"name":        "Mobile App",
				"code":        "MA",
				"description": "Our flagship mobile application for iOS and Android",
				"logo":        "https://example.com/logo.png",
				"budget": map[string]interface{}{
					"raw":       1000000,
					"formatted": 10000.00,
				},
				"stats": map[string]interface{}{
					"total":    50,
					"open":     20,
					"progress": 15,
					"closed":   15,
				},
				"company": map[string]interface{}{
					"uuid": "comp-uuid-1",
					"slug": "acme-corp",
					"name": "ACME Corporation",
				},
				"dates": map[string]interface{}{
					"start": map[string]interface{}{"date": "2024-01-01"},
					"due":   map[string]interface{}{"date": "2024-06-30"},
				},
				"settings": map[string]interface{}{
					"use_timer":       true,
					"has_sprints":     true,
					"has_wiki":        true,
					"has_discussions": true,
					"show_number":     true,
				},
				"users": []map[string]interface{}{
					{"uuid": "u1", "username": "johndoe", "name": "John Doe"},
					{"uuid": "u2", "username": "janedoe", "name": "Jane Doe"},
				},
				"labels": []map[string]interface{}{
					{"id": 1, "slug": "bug", "title": "Bug", "color": "#e74c3c"},
					{"id": 2, "slug": "feature", "title": "Feature", "color": "#2ecc71"},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/projects/mobile-app")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Stats       struct {
				Total  int `json:"total"`
				Open   int `json:"open"`
				Closed int `json:"closed"`
			} `json:"stats"`
			Settings struct {
				UseTimer   bool `json:"use_timer"`
				HasSprints bool `json:"has_sprints"`
			} `json:"settings"`
			Users []struct {
				Username string `json:"username"`
			} `json:"users"`
			Labels []struct {
				Slug  string `json:"slug"`
				Title string `json:"title"`
			} `json:"labels"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if result.Data.Name != "Mobile App" {
		t.Errorf("name = %q, want 'Mobile App'", result.Data.Name)
	}
	if !result.Data.Settings.UseTimer {
		t.Error("expected use_timer to be true")
	}
	if len(result.Data.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(result.Data.Users))
	}
	if len(result.Data.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(result.Data.Labels))
	}
}

// TestProjectsWorkflows_Success tests getting project workflows
func TestProjectsWorkflows_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/workflows") {
			t.Errorf("Path should contain /workflows, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          1,
					"slug":        "backlog",
					"title":       "Backlog",
					"color":       "#95a5a6",
					"description": "Tasks waiting to be started",
					"position":    0,
					"is_default":  true,
					"is_done":     false,
				},
				{
					"id":          2,
					"slug":        "in-progress",
					"title":       "In Progress",
					"color":       "#3498db",
					"description": "Tasks actively being worked on",
					"position":    1,
					"is_default":  false,
					"is_done":     false,
				},
				{
					"id":          3,
					"slug":        "review",
					"title":       "Review",
					"color":       "#f39c12",
					"description": "Tasks under review",
					"position":    2,
					"is_default":  false,
					"is_done":     false,
				},
				{
					"id":          4,
					"slug":        "done",
					"title":       "Done",
					"color":       "#27ae60",
					"description": "Completed tasks",
					"position":    3,
					"is_default":  false,
					"is_done":     true,
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/projects/mobile-app/workflows")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID        int    `json:"id"`
			Slug      string `json:"slug"`
			Title     string `json:"title"`
			Color     string `json:"color"`
			IsDone    bool   `json:"is_done"`
			IsDefault bool   `json:"is_default"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if len(result.Data) != 4 {
		t.Fatalf("expected 4 workflows, got %d", len(result.Data))
	}

	// Check "done" workflow
	doneWf := result.Data[3]
	if doneWf.Slug != "done" {
		t.Errorf("last workflow slug = %q, want 'done'", doneWf.Slug)
	}
	if !doneWf.IsDone {
		t.Error("expected 'done' workflow to have is_done=true")
	}

	// Check default workflow
	backlog := result.Data[0]
	if !backlog.IsDefault {
		t.Error("expected 'backlog' to be default")
	}
}

// TestProjectsCreate_Success tests creating a project
func TestProjectsCreate_Success(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   3,
				"uuid": "new-proj-uuid",
				"slug": receivedBody["slug"],
				"name": receivedBody["name"],
				"code": "NP",
			},
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/projects", map[string]interface{}{
		"name":        "New Project",
		"slug":        "new-project",
		"description": "A brand new project",
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}

	if receivedBody["name"] != "New Project" {
		t.Errorf("name = %v, want 'New Project'", receivedBody["name"])
	}
}

// TestProjectsListEmpty tests empty projects list
func TestProjectsListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []interface{}{},
			"meta": map[string]interface{}{
				"total": 0,
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/projects")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []interface{} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 0 {
		t.Errorf("expected empty data, got %d items", len(result.Data))
	}
}
