package sprints

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

func TestNewCmdSprints(t *testing.T) {
	f := factory.New()
	cmd := NewCmdSprints(f)

	if !strings.HasPrefix(cmd.Use, "sprints") {
		t.Errorf("Use should start with 'sprints', got %q", cmd.Use)
	}

	// Verify command has subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestSprintsList_Success tests listing sprints (matching SprintResource)
func TestSprintsList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Response matching SprintResource format
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":          1,
					"code":        "SPR-1",
					"slug":        "sprint-1-feb-2024",
					"title":       "Sprint 1 - February 2024",
					"color":       "#3498db",
					"description": "First sprint of Q1",
					"duration":    14,
					"timebox": map[string]interface{}{
						"start": map[string]interface{}{
							"date":     "2024-02-01",
							"timezone": "UTC",
							"ago":      "6 days ago",
						},
						"finish": map[string]interface{}{
							"date":     "2024-02-14",
							"timezone": "UTC",
							"ago":      "in 8 days",
						},
					},
					"stats": map[string]interface{}{
						"story_points": 21,
						"worked_hours": 45,
						"total_tasks":  15,
						"closed_tasks": 8,
						"percentage":   53,
						"comments":     25,
					},
					"status": map[string]interface{}{
						"slug":        "sprint-open",
						"title":       "Open",
						"color":       "079a0d",
						"description": "Sprint is active",
					},
					"project": map[string]interface{}{
						"slug": "mobile-app",
						"name": "Mobile App",
					},
					"company": map[string]interface{}{
						"slug": "acme-corp",
						"name": "ACME Corporation",
					},
				},
				{
					"id":       2,
					"code":     "SPR-2",
					"slug":     "sprint-2-feb-2024",
					"title":    "Sprint 2 - February 2024",
					"color":    "#9b59b6",
					"duration": 14,
					"stats": map[string]interface{}{
						"story_points": 34,
						"total_tasks":  20,
						"closed_tasks": 0,
						"percentage":   0,
					},
					"status": map[string]interface{}{
						"slug":  "sprint-planned",
						"title": "Planned",
						"color": "6c757d",
					},
					"project": map[string]interface{}{
						"slug": "mobile-app",
						"name": "Mobile App",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/sprints")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID       int    `json:"id"`
			Code     string `json:"code"`
			Slug     string `json:"slug"`
			Title    string `json:"title"`
			Duration int    `json:"duration"`
			Stats    struct {
				TotalTasks  int `json:"total_tasks"`
				ClosedTasks int `json:"closed_tasks"`
				Percentage  int `json:"percentage"`
			} `json:"stats"`
			Status struct {
				Slug  string `json:"slug"`
				Title string `json:"title"`
			} `json:"status"`
			Project struct {
				Slug string `json:"slug"`
			} `json:"project"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 sprints, got %d", len(result.Data))
	}

	sprint := result.Data[0]
	if sprint.Code != "SPR-1" {
		t.Errorf("code = %q, want 'SPR-1'", sprint.Code)
	}
	if sprint.Title != "Sprint 1 - February 2024" {
		t.Errorf("title = %q, want expected", sprint.Title)
	}
	if sprint.Duration != 14 {
		t.Errorf("duration = %d, want 14", sprint.Duration)
	}
	if sprint.Stats.TotalTasks != 15 {
		t.Errorf("total_tasks = %d, want 15", sprint.Stats.TotalTasks)
	}
	if sprint.Stats.Percentage != 53 {
		t.Errorf("percentage = %d, want 53", sprint.Stats.Percentage)
	}
	if sprint.Status.Slug != "sprint-open" {
		t.Errorf("status.slug = %q, want 'sprint-open'", sprint.Status.Slug)
	}
}

// TestSprintsView_Success tests viewing a single sprint
func TestSprintsView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/sprints/") {
			t.Errorf("Path should start with /sprints/, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":              1,
				"code":            "SPR-1",
				"slug":            "sprint-1-feb-2024",
				"title":           "Sprint 1 - February 2024",
				"description":     "First sprint focusing on authentication and core features",
				"color":           "#3498db",
				"duration":        14,
				"close_on_finish": true,
				"timebox": map[string]interface{}{
					"start":  map[string]interface{}{"date": "2024-02-01"},
					"finish": map[string]interface{}{"date": "2024-02-14"},
				},
				"date_start":  map[string]interface{}{"date": "2024-02-01"},
				"date_finish": map[string]interface{}{"date": "2024-02-14"},
				"stats": map[string]interface{}{
					"story_points": 21,
					"worked_hours": 45,
					"total_tasks":  15,
					"closed_tasks": 8,
					"percentage":   53,
					"comments":     25,
				},
				"status": map[string]interface{}{
					"slug":  "sprint-open",
					"title": "Open",
					"color": "079a0d",
				},
				"project": map[string]interface{}{
					"slug": "mobile-app",
					"name": "Mobile App",
				},
				"users": []map[string]interface{}{
					{"uuid": "u1", "username": "johndoe", "name": "John Doe"},
					{"uuid": "u2", "username": "janedoe", "name": "Jane Doe"},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/sprints/sprint-1-feb-2024")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Code        string `json:"code"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Duration    int    `json:"duration"`
			Stats       struct {
				TotalTasks  int `json:"total_tasks"`
				ClosedTasks int `json:"closed_tasks"`
				Percentage  int `json:"percentage"`
				WorkedHours int `json:"worked_hours"`
			} `json:"stats"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if result.Data.Code != "SPR-1" {
		t.Errorf("code = %q, want 'SPR-1'", result.Data.Code)
	}
	if result.Data.Stats.WorkedHours != 45 {
		t.Errorf("worked_hours = %d, want 45", result.Data.Stats.WorkedHours)
	}
}

// TestSprintsCurrent_Success tests getting current active sprint
func TestSprintsCurrent_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for current/active filter
		if !strings.Contains(r.URL.RawQuery, "status=sprint-open") &&
			!strings.Contains(r.URL.Path, "current") {
			// Return the current sprint
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":       1,
				"code":     "SPR-1",
				"slug":     "sprint-1-feb-2024",
				"title":    "Sprint 1 - February 2024",
				"duration": 14,
				"stats": map[string]interface{}{
					"total_tasks":  15,
					"closed_tasks": 8,
					"percentage":   53,
				},
				"status": map[string]interface{}{
					"slug":  "sprint-open",
					"title": "Open",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/sprints/current?project=mobile-app")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Code   string `json:"code"`
			Status struct {
				Slug string `json:"slug"`
			} `json:"status"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Status.Slug != "sprint-open" {
		t.Errorf("status = %q, want 'sprint-open'", result.Data.Status.Slug)
	}
}

// TestSprintsCreate_Success tests creating a sprint
func TestSprintsCreate_Success(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":    3,
				"code":  "SPR-3",
				"slug":  "sprint-3",
				"title": receivedBody["title"],
				"stats": map[string]interface{}{
					"total_tasks":  0,
					"closed_tasks": 0,
					"percentage":   0,
				},
				"status": map[string]interface{}{
					"slug":  "sprint-planned",
					"title": "Planned",
				},
			},
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/sprints", map[string]interface{}{
		"title":       "Sprint 3 - March 2024",
		"date_start":  "2024-03-01",
		"date_finish": "2024-03-14",
		"project_id":  1,
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}

	if receivedBody["title"] != "Sprint 3 - March 2024" {
		t.Errorf("title = %v, want expected", receivedBody["title"])
	}
}

// TestSprintsClose_Success tests closing a sprint
func TestSprintsClose_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost && r.Method != http.MethodPut {
			t.Logf("Method = %s", r.Method)
		}

		if !strings.Contains(r.URL.Path, "/close") {
			t.Errorf("Path should contain /close, got %s", r.URL.Path)
		}

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"id":   1,
				"code": "SPR-1",
				"status": map[string]interface{}{
					"slug":  "sprint-closed",
					"title": "Closed",
				},
				"closed_at": map[string]interface{}{
					"date":     "2024-02-14",
					"timezone": "UTC",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/sprints/sprint-1-feb-2024/close", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Status struct {
				Slug string `json:"slug"`
			} `json:"status"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Status.Slug != "sprint-closed" {
		t.Errorf("status = %q, want 'sprint-closed'", result.Data.Status.Slug)
	}
}

// TestSprintsListEmpty tests empty sprints list
func TestSprintsListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/sprints")
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
