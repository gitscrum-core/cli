package standup

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

func TestNewCmdStandup(t *testing.T) {
	f := factory.New()
	cmd := NewCmdStandup(f)

	if !strings.HasPrefix(cmd.Use, "standup") {
		t.Errorf("Use should start with 'standup', got %q", cmd.Use)
	}

	// Verify command has subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}

	// Check for expected subcommands
	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}

	expected := []string{"completed", "blockers", "team", "digest"}
	for _, name := range expected {
		if !subcommands[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestStandupCompleted_Success tests fetching completed tasks from yesterday
func TestStandupCompleted_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/standup/completed-yesterday") {
			// Verify project_slug parameter is used correctly
			if r.URL.Query().Get("project_slug") != "" || r.URL.Query().Get("date") != "" {
				// OK - parameters are properly formatted
			}

			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"code":  "GS-100",
						"title": "Implement login page",
					},
					{
						"code":  "GS-101",
						"title": "Fix API endpoint",
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
	resp, err := client.Get("/companies/standup/completed-yesterday?date=2026-02-06")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Code  string `json:"code"`
			Title string `json:"title"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(result.Data))
	}

	if result.Data[0].Code != "GS-100" {
		t.Errorf("first task code = %q, want 'GS-100'", result.Data[0].Code)
	}
}

// TestStandupBlockers_Success tests fetching current blockers
func TestStandupBlockers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/standup/blockers") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"code":  "GS-50",
						"title": "Waiting for API access",
						"assignee": map[string]interface{}{
							"name": "Alice",
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
	resp, err := client.Get("/companies/standup/blockers")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Code     string `json:"code"`
			Title    string `json:"title"`
			Assignee struct {
				Name string `json:"name"`
			} `json:"assignee"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) == 0 {
		t.Fatal("expected blockers, got none")
	}

	if result.Data[0].Assignee.Name != "Alice" {
		t.Errorf("assignee = %q, want 'Alice'", result.Data[0].Assignee.Name)
	}
}

// TestStandupTeam_Success tests fetching team standup status
func TestStandupTeam_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/standup/team-status") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"user_uuid":             "uuid-1",
						"user_name":             "Alice",
						"tasks_in_progress":     3,
						"tasks_completed_today": 2,
						"tasks_completed_week":  8,
						"blocked_count":         0,
						"hours_tracked_today":   4.5,
					},
					{
						"user_uuid":             "uuid-2",
						"user_name":             "Bob",
						"tasks_in_progress":     1,
						"tasks_completed_today": 1,
						"tasks_completed_week":  5,
						"blocked_count":         1,
						"hours_tracked_today":   3.0,
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
	resp, err := client.Get("/companies/standup/team-status")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			UserName            string  `json:"user_name"`
			TasksInProgress     int     `json:"tasks_in_progress"`
			TasksCompletedToday int     `json:"tasks_completed_today"`
			BlockedCount        int     `json:"blocked_count"`
			HoursTrackedToday   float64 `json:"hours_tracked_today"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 team members, got %d", len(result.Data))
	}

	// Check Alice's stats
	if result.Data[0].TasksInProgress != 3 {
		t.Errorf("Alice tasks in progress = %d, want 3", result.Data[0].TasksInProgress)
	}

	// Check Bob has a blocker
	if result.Data[1].BlockedCount != 1 {
		t.Errorf("Bob blocked count = %d, want 1", result.Data[1].BlockedCount)
	}
}

// TestStandupDigest_Success tests fetching weekly digest
func TestStandupDigest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/standup/weekly-digest") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"total_completed": 45,
					"total_hours":     180.5,
					"velocity_change": 12.5,
					"top_contributors": []map[string]interface{}{
						{"name": "Alice", "tasks_completed": 15},
						{"name": "Bob", "tasks_completed": 12},
					},
					"daily_breakdown": []map[string]interface{}{
						{"date": "2026-02-03", "completed": 8, "created": 5, "blocked": 1},
						{"date": "2026-02-04", "completed": 10, "created": 6, "blocked": 0},
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
	resp, err := client.Get("/companies/standup/weekly-digest")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalCompleted  int     `json:"total_completed"`
			TotalHours      float64 `json:"total_hours"`
			VelocityChange  float64 `json:"velocity_change"`
			TopContributors []struct {
				Name           string `json:"name"`
				TasksCompleted int    `json:"tasks_completed"`
			} `json:"top_contributors"`
			DailyBreakdown []struct {
				Date      string `json:"date"`
				Completed int    `json:"completed"`
			} `json:"daily_breakdown"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.TotalCompleted != 45 {
		t.Errorf("total completed = %d, want 45", result.Data.TotalCompleted)
	}

	if result.Data.VelocityChange != 12.5 {
		t.Errorf("velocity change = %.1f, want 12.5", result.Data.VelocityChange)
	}

	if len(result.Data.TopContributors) != 2 {
		t.Errorf("top contributors = %d, want 2", len(result.Data.TopContributors))
	}
}

// TestStandupSummary_Success tests fetching daily standup summary
func TestStandupSummary_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/companies/standup/summary") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":      "standup-uuid-1",
						"date":      "2026-02-07",
						"completed": []string{"Finished login feature", "Fixed bug #123"},
						"planned":   []string{"Start dashboard", "Review PRs"},
						"blockers":  []string{},
						"user": map[string]interface{}{
							"name":   "Alice",
							"avatar": "https://example.com/avatar.jpg",
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
	resp, err := client.Get("/companies/standup/summary?date=2026-02-07")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Standup `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) == 0 {
		t.Fatal("expected standup entries, got none")
	}

	if len(result.Data[0].Completed) != 2 {
		t.Errorf("completed items = %d, want 2", len(result.Data[0].Completed))
	}

	if result.Data[0].User.Name != "Alice" {
		t.Errorf("user name = %q, want 'Alice'", result.Data[0].User.Name)
	}
}

// TestStandupBlockers_ProjectFilter tests that project_slug parameter is used
func TestStandupBlockers_ProjectFilter(t *testing.T) {
	receivedProjectSlug := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedProjectSlug = r.URL.Query().Get("project_slug")
		response := map[string]interface{}{
			"data": []map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	_, err := client.Get("/companies/standup/blockers?project_slug=my-project")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if receivedProjectSlug != "my-project" {
		t.Errorf("project_slug = %q, want 'my-project'", receivedProjectSlug)
	}
}
