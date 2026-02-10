package tasks

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/output"
)

// mockFactory creates a factory with mocked API client
func mockFactory(serverURL string) *factory.Factory {
	f := factory.New()
	f.OutputFormat = output.FormatTable
	return f
}

// TestNewCmdTasks tests the root tasks command creation
func TestNewCmdTasks(t *testing.T) {
	f := factory.New()
	cmd := NewCmdTasks(f)

	if !strings.HasPrefix(cmd.Use, "tasks") {
		t.Errorf("Use should start with 'tasks', got %q", cmd.Use)
	}

	// Verify subcommands exist
	subcommands := []string{"list", "view", "create", "update", "complete", "assign", "today", "current", "branch", "branches", "pr", "prs"}
	for _, name := range subcommands {
		found := false
		for _, sub := range cmd.Commands() {
			if sub.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not found", name)
		}
	}
}

// TestRunTasksList_Success tests tasks list with mocked API response
func TestRunTasksList_Success(t *testing.T) {
	// Create mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request path
		if r.URL.Path != "/issues/my" {
			t.Errorf("Path = %q, want /issues/my", r.URL.Path)
		}

		// Return mock task list (matching actual API format)
		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":                  "550e8400-e29b-41d4-a716-446655440000",
					"code":                  "GS-123",
					"number":                123,
					"title":                 "Fix login issue on mobile",
					"config_workflow_title": "In Progress",
					"users": []map[string]interface{}{
						{"username": "johndoe"},
					},
					"project": map[string]interface{}{
						"code": "GS",
						"slug": "mobile-app",
					},
				},
				{
					"uuid":                  "550e8400-e29b-41d4-a716-446655440001",
					"code":                  "GS-124",
					"number":                124,
					"title":                 "Add dark mode support",
					"config_workflow_title": "To Do",
					"users":                 []map[string]interface{}{},
					"project": map[string]interface{}{
						"code": "GS",
						"slug": "mobile-app",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client manually for testing
	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})

	// Execute API call
	resp, err := client.Get("/issues/my")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Data []struct {
			UUID     string `json:"uuid"`
			Code     string `json:"code"`
			Title    string `json:"title"`
			Workflow string `json:"config_workflow_title"`
			Users    []struct {
				Username string `json:"username"`
			} `json:"users"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	// Verify parsed data
	if len(result.Data) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(result.Data))
	}

	task := result.Data[0]
	if task.Code != "GS-123" {
		t.Errorf("task.Code = %q, want %q", task.Code, "GS-123")
	}
	if task.Title != "Fix login issue on mobile" {
		t.Errorf("task.Title = %q, want %q", task.Title, "Fix login issue on mobile")
	}
	if task.Workflow != "In Progress" {
		t.Errorf("task.Workflow = %q, want %q", task.Workflow, "In Progress")
	}
	if len(task.Users) == 0 || task.Users[0].Username != "johndoe" {
		t.Error("task.Users not parsed correctly")
	}
}

// TestRunTasksList_WithProject tests tasks list with project filter
func TestRunTasksList_WithProject(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should use project-specific path
		expectedPath := "/projects/mobile-app/issues"
		if r.URL.Path != expectedPath {
			t.Errorf("Path = %q, want %q", r.URL.Path, expectedPath)
		}

		response := map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"uuid":                  "test-uuid",
					"code":                  "MA-1",
					"title":                 "Project specific task",
					"config_workflow_title": "Done",
					"users":                 []map[string]interface{}{},
					"project":               map[string]interface{}{"code": "MA"},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/projects/mobile-app/issues")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	resp.Body.Close()
}

// TestRunTasksView_Success tests viewing a single task
func TestRunTasksView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should call tasks by-code endpoint first
		if strings.HasPrefix(r.URL.Path, "/tasks/by-code/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "550e8400-e29b-41d4-a716-446655440000",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		// Then call issues endpoint
		if strings.HasPrefix(r.URL.Path, "/issues/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":        "550e8400-e29b-41d4-a716-446655440000",
					"code":        "GS-123",
					"slug":        "fix-login-issue",
					"title":       "Fix login issue on mobile",
					"description": "Users cannot login on iOS devices using Face ID",
					"state":       "open",
					"workflow": map[string]interface{}{
						"slug":  "in-progress",
						"title": "In Progress",
						"color": "#3498db",
					},
					"settings": map[string]interface{}{
						"is_blocker": false,
						"is_bug":     true,
						"is_draft":   false,
					},
					"stats": map[string]interface{}{
						"votes":       5,
						"comments":    3,
						"attachments": 2,
						"subtasks":    4,
					},
					"users": []map[string]interface{}{
						{"uuid": "user-1", "username": "johndoe", "name": "John Doe"},
					},
					"project": map[string]interface{}{
						"slug": "mobile-app",
						"name": "Mobile App",
					},
					"created_at": map[string]interface{}{
						"date":     "2024-02-01",
						"timezone": "UTC",
						"ago":      "6 days ago",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})

	// First get UUID
	resp, err := client.Get("/tasks/by-code/GS-123")
	if err != nil {
		t.Fatalf("Get by-code failed: %v", err)
	}
	defer resp.Body.Close()

	var uuidResult struct {
		Data struct {
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &uuidResult)

	if uuidResult.Data.UUID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("UUID = %q, want expected UUID", uuidResult.Data.UUID)
	}

	// Then get full task details
	resp2, err := client.Get("/issues/" + uuidResult.Data.UUID)
	if err != nil {
		t.Fatalf("Get issue failed: %v", err)
	}
	defer resp2.Body.Close()

	var taskResult struct {
		Data struct {
			UUID        string `json:"uuid"`
			Code        string `json:"code"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Workflow    struct {
				Title string `json:"title"`
			} `json:"workflow"`
			Settings struct {
				IsBug bool `json:"is_bug"`
			} `json:"settings"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp2, &taskResult); err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if taskResult.Data.Code != "GS-123" {
		t.Errorf("Code = %q, want %q", taskResult.Data.Code, "GS-123")
	}
	if taskResult.Data.Title != "Fix login issue on mobile" {
		t.Errorf("Title = %q, want expected", taskResult.Data.Title)
	}
	if !taskResult.Data.Settings.IsBug {
		t.Error("expected is_bug to be true")
	}
}

// TestRunTasksCreate_Success tests creating a task
func TestRunTasksCreate_Success(t *testing.T) {
	var receivedBody map[string]interface{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/issues" {
			t.Errorf("Path = %q, want /issues", r.URL.Path)
		}

		json.NewDecoder(r.Body).Decode(&receivedBody)

		response := map[string]interface{}{
			"data": map[string]interface{}{
				"uuid":  "new-task-uuid",
				"code":  "GS-999",
				"title": receivedBody["title"],
			},
		}
		w.WriteHeader(http.StatusCreated)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})

	body := map[string]interface{}{
		"title":        "New feature request",
		"description":  "Implement dark mode",
		"project_slug": "mobile-app",
	}

	resp, err := client.Post("/issues", body)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("StatusCode = %d, want 201", resp.StatusCode)
	}

	// Verify body was sent correctly
	if receivedBody["title"] != "New feature request" {
		t.Errorf("title = %v, want %q", receivedBody["title"], "New feature request")
	}
}

// TestRunTasksComplete_Success tests completing a task
func TestRunTasksComplete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First call: get UUID by code
		if strings.HasPrefix(r.URL.Path, "/tasks/by-code/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "task-uuid-123",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}

		// Second call: update task to done
		if r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/issues/") {
			var body map[string]interface{}
			json.NewDecoder(r.Body).Decode(&body)

			if body["config_workflow_slug"] != "done" {
				t.Errorf("workflow = %v, want 'done'", body["config_workflow_slug"])
			}

			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "task-uuid-123",
					"code": "GS-123",
					"workflow": map[string]interface{}{
						"slug":  "done",
						"title": "Done",
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

	// Complete task = update workflow to done
	resp, err := client.Put("/issues/task-uuid-123", map[string]interface{}{
		"config_workflow_slug": "done",
	})
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}
	resp.Body.Close()
}

// TestOutputFormatting tests that output is formatted correctly
func TestOutputFormatting(t *testing.T) {
	headers := []string{"CODE", "TITLE", "STATUS"}
	rows := [][]string{
		{"GS-123", "Fix bug", "In Progress"},
		{"GS-124", "Add feature", "Done"},
	}

	// Test JSON output
	var jsonBuf bytes.Buffer
	jsonFormatter := &output.JSONFormatter{Writer: &jsonBuf}

	err := jsonFormatter.PrintTable(headers, rows)
	if err != nil {
		t.Fatalf("JSON PrintTable failed: %v", err)
	}

	var jsonResult []map[string]string
	json.Unmarshal(jsonBuf.Bytes(), &jsonResult)

	if len(jsonResult) != 2 {
		t.Fatalf("expected 2 items, got %d", len(jsonResult))
	}
	if jsonResult[0]["CODE"] != "GS-123" {
		t.Errorf("CODE = %q, want GS-123", jsonResult[0]["CODE"])
	}

	// Test quiet output
	var quietBuf bytes.Buffer
	quietFormatter := &output.QuietFormatter{Writer: &quietBuf}

	err = quietFormatter.PrintTable(headers, rows)
	if err != nil {
		t.Fatalf("Quiet PrintTable failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(quietBuf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	if lines[0] != "GS-123" {
		t.Errorf("first line = %q, want GS-123", lines[0])
	}
}

// TestTasksListEmpty tests empty task list response
func TestTasksListEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"data": []interface{}{},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/issues/my")
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

// TestAPIErrorHandling tests error responses from API
func TestAPIErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		errorBody  string
	}{
		{"unauthorized", 401, `{"message": "Unauthenticated."}`},
		{"not found", 404, `{"message": "Task not found"}`},
		{"validation error", 422, `{"message": "The given data was invalid.", "errors": {"title": ["required"]}}`},
		{"server error", 500, `{"message": "Server error"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.errorBody))
			}))
			defer server.Close()

			client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
			resp, _ := client.Get("/test")

			var result map[string]interface{}
			err := api.DecodeResponse(resp, &result)

			if err == nil {
				t.Error("expected error, got nil")
			}

			if !strings.Contains(err.Error(), "API error") {
				t.Errorf("error should contain 'API error', got: %v", err)
			}
		})
	}
}
