package analytics

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

func TestNewCmdAnalytics(t *testing.T) {
	f := factory.New()
	cmd := NewCmdAnalytics(f)

	if !strings.HasPrefix(cmd.Use, "analytics") {
		t.Errorf("Use should start with 'analytics', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}

	// Check for expected subcommands
	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}

	expected := []string{"velocity", "workload", "blockers", "cycle-time", "throughput"}
	for _, name := range expected {
		if !subcommands[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestAnalyticsVelocity_Success tests velocity analytics
func TestAnalyticsVelocity_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cli/analytics/velocity") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"current_velocity":  45,
					"previous_velocity": 40,
					"change_percentage": 12.5,
					"trend":             "up",
					"sprints": []map[string]interface{}{
						{"name": "Sprint 10", "completed": 45, "planned": 50},
						{"name": "Sprint 9", "completed": 40, "planned": 48},
						{"name": "Sprint 8", "completed": 38, "planned": 45},
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
	resp, err := client.Get("/cli/analytics/velocity")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			CurrentVelocity  int     `json:"current_velocity"`
			PreviousVelocity int     `json:"previous_velocity"`
			ChangePercentage float64 `json:"change_percentage"`
			Trend            string  `json:"trend"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.CurrentVelocity != 45 {
		t.Errorf("current velocity = %d, want 45", result.Data.CurrentVelocity)
	}

	if result.Data.Trend != "up" {
		t.Errorf("trend = %q, want 'up'", result.Data.Trend)
	}
}

// TestAnalyticsWorkload_Success tests workload analytics
func TestAnalyticsWorkload_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cli/analytics/workload") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"team_members": []map[string]interface{}{
						{
							"name":         "Alice",
							"tasks":        8,
							"story_points": 21,
							"capacity":     24,
							"utilization":  87.5,
						},
						{
							"name":         "Bob",
							"tasks":        6,
							"story_points": 18,
							"capacity":     24,
							"utilization":  75.0,
						},
					},
					"total_tasks":  14,
					"total_points": 39,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/cli/analytics/workload")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TeamMembers []struct {
				Name        string  `json:"name"`
				Tasks       int     `json:"tasks"`
				Utilization float64 `json:"utilization"`
			} `json:"team_members"`
			TotalTasks int `json:"total_tasks"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data.TeamMembers) != 2 {
		t.Fatalf("team members = %d, want 2", len(result.Data.TeamMembers))
	}

	if result.Data.TotalTasks != 14 {
		t.Errorf("total tasks = %d, want 14", result.Data.TotalTasks)
	}
}

// TestAnalyticsBlockers_Success tests blockers analytics
func TestAnalyticsBlockers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cli/analytics/blockers") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"total_blockers": 5,
					"blockers": []map[string]interface{}{
						{
							"code":         "GS-123",
							"title":        "Waiting for API access",
							"blocked_days": 3,
							"assignee":     map[string]interface{}{"name": "Alice"},
						},
						{
							"code":         "GS-124",
							"title":        "Design review pending",
							"blocked_days": 2,
							"assignee":     map[string]interface{}{"name": "Bob"},
						},
					},
					"avg_blocked_days": 2.5,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/cli/analytics/blockers")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalBlockers int `json:"total_blockers"`
			Blockers      []struct {
				Code        string `json:"code"`
				BlockedDays int    `json:"blocked_days"`
			} `json:"blockers"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.TotalBlockers != 5 {
		t.Errorf("total blockers = %d, want 5", result.Data.TotalBlockers)
	}
}

// TestAnalyticsCycleTime_Success tests cycle time analytics
func TestAnalyticsCycleTime_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cli/analytics/cycle-time") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"avg_cycle_time": 4.5,
					"p50_cycle_time": 3.0,
					"p90_cycle_time": 8.0,
					"trend":          "improving",
					"by_type": map[string]interface{}{
						"bug":     2.5,
						"feature": 6.0,
						"task":    3.0,
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
	resp, err := client.Get("/cli/analytics/cycle-time")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			AvgCycleTime float64 `json:"avg_cycle_time"`
			P50CycleTime float64 `json:"p50_cycle_time"`
			Trend        string  `json:"trend"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.AvgCycleTime != 4.5 {
		t.Errorf("avg cycle time = %.1f, want 4.5", result.Data.AvgCycleTime)
	}

	if result.Data.Trend != "improving" {
		t.Errorf("trend = %q, want 'improving'", result.Data.Trend)
	}
}

// TestAnalyticsThroughput_Success tests throughput analytics
func TestAnalyticsThroughput_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/cli/analytics/throughput") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"daily_avg":    8.5,
					"weekly_total": 42,
					"monthly_avg":  175,
					"weekly_breakdown": []map[string]interface{}{
						{"day": "Mon", "completed": 10},
						{"day": "Tue", "completed": 8},
						{"day": "Wed", "completed": 9},
						{"day": "Thu", "completed": 7},
						{"day": "Fri", "completed": 8},
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
	resp, err := client.Get("/cli/analytics/throughput")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			DailyAvg    float64 `json:"daily_avg"`
			WeeklyTotal int     `json:"weekly_total"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.WeeklyTotal != 42 {
		t.Errorf("weekly total = %d, want 42", result.Data.WeeklyTotal)
	}
}
