package timer

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

func TestNewCmdTimer(t *testing.T) {
	f := factory.New()
	cmd := NewCmdTimer(f)

	if !strings.HasPrefix(cmd.Use, "timer") {
		t.Errorf("Use should start with 'timer', got %q", cmd.Use)
	}

	// Verify command has subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestTimerStart_Success tests starting a timer on a task
func TestTimerStart_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		// First call: get task UUID
		if strings.HasPrefix(r.URL.Path, "/tasks/by-code/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "task-uuid-123",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Second call: create time tracker
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/create-start") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"comment": "Working on task",
					"time": map[string]interface{}{
						"start": map[string]interface{}{
							"date":     "2024-02-07T10:00:00",
							"timezone": "UTC",
						},
						"end":              nil,
						"total":            "00:00:00",
						"duration_minutes": 0,
					},
					"task": map[string]interface{}{
						"uuid":  "task-uuid-123",
						"code":  "GS-123",
						"title": "Fix login issue",
					},
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

	// Get task UUID
	resp, err := client.Get("/tasks/by-code/GS-123")
	if err != nil {
		t.Fatalf("Get task failed: %v", err)
	}
	resp.Body.Close()

	// Start timer
	resp2, err := client.Post("/time-trackings/task-uuid-123/create-start", map[string]interface{}{
		"comment": "Working on task",
	})
	if err != nil {
		t.Fatalf("Post timer failed: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp2.StatusCode)
	}

	var result struct {
		Data struct {
			ID   int `json:"id"`
			Time struct {
				Start interface{} `json:"start"`
				End   interface{} `json:"end"`
			} `json:"time"`
		} `json:"data"`
	}
	api.DecodeResponse(resp2, &result)

	if result.Data.ID != 1 {
		t.Errorf("ID = %d, want 1", result.Data.ID)
	}
	if result.Data.Time.End != nil {
		t.Error("End should be nil for running timer")
	}
}

// TestTimerStop_Success tests stopping a running timer
func TestTimerStop_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get active timer
		if r.URL.Path == "/time-trackings/active" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":      1,
					"comment": "Working",
					"time": map[string]interface{}{
						"start": map[string]interface{}{
							"date": "2024-02-07T10:00:00",
						},
						"end":              nil,
						"duration_minutes": 45,
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Stop timer
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/stop") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id": 1,
					"time": map[string]interface{}{
						"start":            map[string]interface{}{"date": "2024-02-07T10:00:00"},
						"end":              map[string]interface{}{"date": "2024-02-07T10:45:00"},
						"total":            "00:45:00",
						"duration_minutes": 45,
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

	// Get active timer
	resp, err := client.Get("/time-trackings/active")
	if err != nil {
		t.Fatalf("Get active failed: %v", err)
	}

	var activeResult struct {
		Data struct {
			ID int `json:"id"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &activeResult)

	if activeResult.Data.ID != 1 {
		t.Errorf("active timer ID = %d, want 1", activeResult.Data.ID)
	}

	// Stop timer
	resp2, err := client.Post("/time-trackings/1/stop", nil)
	if err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
	defer resp2.Body.Close()

	var stopResult struct {
		Data struct {
			Time struct {
				DurationMinutes int `json:"duration_minutes"`
			} `json:"time"`
		} `json:"data"`
	}
	api.DecodeResponse(resp2, &stopResult)

	if stopResult.Data.Time.DurationMinutes != 45 {
		t.Errorf("duration = %d, want 45", stopResult.Data.Time.DurationMinutes)
	}
}

// TestTimerStatus_Running tests getting status of running timer
func TestTimerStatus_Running(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/time-trackings/active" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"id":      5,
					"comment": "Debugging login",
					"time": map[string]interface{}{
						"id":               5,
						"start":            map[string]interface{}{"date": "2024-02-07T14:00:00"},
						"end":              nil,
						"total":            "01:23:45",
						"duration_minutes": 83,
					},
					"task": map[string]interface{}{
						"uuid":  "task-abc",
						"code":  "GS-456",
						"title": "Debug login flow",
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
	resp, err := client.Get("/time-trackings/active")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Comment string `json:"comment"`
			Time    struct {
				Total           string `json:"total"`
				DurationMinutes int    `json:"duration_minutes"`
			} `json:"time"`
			Task struct {
				Code  string `json:"code"`
				Title string `json:"title"`
			} `json:"task"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Comment != "Debugging login" {
		t.Errorf("comment = %q, want 'Debugging login'", result.Data.Comment)
	}
	if result.Data.Task.Code != "GS-456" {
		t.Errorf("task code = %q, want 'GS-456'", result.Data.Task.Code)
	}
}

// TestTimerStatus_NoActive tests when no timer is running
func TestTimerStatus_NoActive(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Empty response or null
		response := map[string]interface{}{
			"data": nil,
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/time-trackings/active")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data interface{} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data != nil {
		t.Errorf("expected nil data for no active timer, got %v", result.Data)
	}
}

// TestTimerLog_Success tests getting timer log
func TestTimerLog_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/time-trackings" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":      1,
						"comment": "Morning work",
						"time": map[string]interface{}{
							"start":            map[string]interface{}{"date": "2024-02-07T09:00:00"},
							"end":              map[string]interface{}{"date": "2024-02-07T12:00:00"},
							"total":            "03:00:00",
							"duration_minutes": 180,
						},
						"task": map[string]interface{}{"code": "GS-100", "title": "Feature A"},
					},
					{
						"id":      2,
						"comment": "Afternoon work",
						"time": map[string]interface{}{
							"start":            map[string]interface{}{"date": "2024-02-07T13:00:00"},
							"end":              map[string]interface{}{"date": "2024-02-07T17:00:00"},
							"total":            "04:00:00",
							"duration_minutes": 240,
						},
						"task": map[string]interface{}{"code": "GS-101", "title": "Feature B"},
					},
				},
				"meta": map[string]interface{}{
					"current_page": 1,
					"per_page":     15,
					"total":        2,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/time-trackings")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			ID      int    `json:"id"`
			Comment string `json:"comment"`
			Time    struct {
				DurationMinutes int `json:"duration_minutes"`
			} `json:"time"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result.Data))
	}

	totalMinutes := 0
	for _, entry := range result.Data {
		totalMinutes += entry.Time.DurationMinutes
	}

	if totalMinutes != 420 { // 7 hours
		t.Errorf("total = %d minutes, want 420", totalMinutes)
	}
}

// TestTimerToday_Success tests getting today's timer summary
func TestTimerToday_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for date filter
		if !strings.Contains(r.URL.RawQuery, "date_from") {
			t.Error("expected date_from query parameter")
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id": 1,
					"time": map[string]interface{}{
						"duration_minutes": 120,
					},
					"task": map[string]interface{}{"code": "GS-50"},
				},
				{
					"id": 2,
					"time": map[string]interface{}{
						"duration_minutes": 90,
					},
					"task": map[string]interface{}{"code": "GS-51"},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/time-trackings?date_from=2024-02-07&date_to=2024-02-07")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Time struct {
				DurationMinutes int `json:"duration_minutes"`
			} `json:"time"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	total := 0
	for _, e := range result.Data {
		total += e.Time.DurationMinutes
	}

	if total != 210 {
		t.Errorf("total today = %d minutes, want 210", total)
	}
}
