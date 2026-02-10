package services

import (
	"net/http"
	"testing"
)

func TestTasksService_Get_Success(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks/by-code/GS-123", http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"uuid":  "abc-123",
			"code":  "GS-123",
			"title": "Test Task",
		},
	})

	svc := &tasksService{client: mock}
	task, err := svc.Get("GS-123")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if task.Code != "GS-123" {
		t.Errorf("Expected code GS-123, got %s", task.Code)
	}

	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got %s", task.Title)
	}
}

func TestTasksService_Get_NotFound(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks/by-code/INVALID", http.StatusNotFound,
		map[string]string{"message": "Task not found"})

	svc := &tasksService{client: mock}
	_, err := svc.Get("INVALID")

	if !IsNotFound(err) {
		t.Errorf("Expected ErrNotFound, got %v", err)
	}
}

func TestTasksService_Get_ProRequired(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks/by-code/GS-123", http.StatusPaymentRequired,
		map[string]string{"message": "PRO subscription required"})

	svc := &tasksService{client: mock}
	_, err := svc.Get("GS-123")

	if !IsPaymentRequired(err) {
		t.Errorf("Expected ErrPaymentRequired, got %v", err)
	}
}

func TestTasksService_Get_Unauthorized(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks/by-code/GS-123", http.StatusUnauthorized,
		map[string]string{"message": "Unauthorized"})

	svc := &tasksService{client: mock}
	_, err := svc.Get("GS-123")

	if !IsUnauthorized(err) {
		t.Errorf("Expected ErrUnauthorized, got %v", err)
	}
}

func TestTasksService_Complete_Success(t *testing.T) {
	mock := NewMockClient()
	mock.OnPut("/tasks/by-code/GS-123", http.StatusOK, map[string]interface{}{
		"data": map[string]interface{}{
			"uuid":   "abc-123",
			"code":   "GS-123",
			"status": map[string]string{"slug": "done"},
		},
	})

	svc := &tasksService{client: mock}
	err := svc.Complete("GS-123")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !mock.AssertCalled("PUT", "/tasks/by-code/GS-123") {
		t.Error("Expected PUT to be called")
	}
}

func TestTasksService_List_Success(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks?project=my-project&", http.StatusOK, map[string]interface{}{
		"data": []map[string]interface{}{
			{"code": "GS-1", "title": "Task 1"},
			{"code": "GS-2", "title": "Task 2"},
		},
	})

	svc := &tasksService{client: mock}
	tasks, err := svc.List(TasksListOptions{Project: "my-project"})

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(tasks))
	}
}

func TestTasksService_List_RateLimited(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks?", http.StatusTooManyRequests,
		map[string]string{"message": "Rate limit exceeded"})

	svc := &tasksService{client: mock}
	_, err := svc.List(TasksListOptions{})

	if !IsRateLimited(err) {
		t.Errorf("Expected ErrRateLimited, got %v", err)
	}
}

func TestTasksService_Create_ValidationError(t *testing.T) {
	mock := NewMockClient()
	mock.OnPost("/tasks", http.StatusUnprocessableEntity, map[string]interface{}{
		"message": "Validation failed",
		"errors": map[string][]string{
			"title": {"Title is required"},
		},
	})

	svc := &tasksService{client: mock}
	_, err := svc.Create(CreateTaskInput{Project: "test"})

	if err == nil {
		t.Error("Expected validation error")
		return
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Errorf("Expected *APIError, got %T", err)
		return
	}

	if len(apiErr.Errors["title"]) == 0 {
		t.Error("Expected title validation error")
	}
}

func TestTasksService_Create_ServerError(t *testing.T) {
	mock := NewMockClient()
	mock.OnPost("/tasks", http.StatusInternalServerError,
		map[string]string{"message": "Internal server error"})

	svc := &tasksService{client: mock}
	_, err := svc.Create(CreateTaskInput{Title: "Test", Project: "test"})

	if !IsServerError(err) {
		t.Errorf("Expected ErrServerError, got %v", err)
	}
}
