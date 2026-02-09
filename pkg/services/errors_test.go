package services

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestHandleResponse_Success(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{"200 OK", http.StatusOK, false},
		{"201 Created", http.StatusCreated, false},
		{"204 No Content", http.StatusNoContent, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.OnGet("/test", tt.statusCode, map[string]interface{}{"data": nil})

			resp, _ := mock.Get("/test")
			err := handleResponse(resp, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleResponse_UsesAPIMessage(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/pro-feature", http.StatusPaymentRequired, map[string]interface{}{
		"message":     "GitScrum PRO is required to access git branches. Upgrade at https://gitscrum.com/pricing",
		"error":       "pro_required",
		"feature":     "git branches",
		"upgrade_url": "https://gitscrum.com/pricing",
	})

	resp, _ := mock.Get("/pro-feature")
	err := handleResponse(resp, nil)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", err)
	}

	// Should use the API message, not hardcoded
	expectedMsg := "GitScrum PRO is required to access git branches. Upgrade at https://gitscrum.com/pricing"
	if apiErr.Message != expectedMsg {
		t.Errorf("Expected message from API: %q, got: %q", expectedMsg, apiErr.Message)
	}

	if apiErr.Feature != "git branches" {
		t.Errorf("Expected feature 'git branches', got %q", apiErr.Feature)
	}

	if apiErr.UpgradeURL != "https://gitscrum.com/pricing" {
		t.Errorf("Expected upgrade URL, got %q", apiErr.UpgradeURL)
	}
}

func TestHandleResponse_ServerError_AlwaysSanitizes(t *testing.T) {
	// For 500 errors, we ALWAYS use fallback - never expose any API message
	dangerousMessages := []string{
		"SQLSTATE[42000]: Syntax error or access violation",
		"No query results for model [App\\Models\\Task]",
		"Exception thrown at line 42\nStack trace:\n#0 /var/www/app/...",
		"Error in C:\\Users\\dev\\project\\file.php",
		"Error in /var/www/html/app/Models/Task.php",
		"PDOException: SQLSTATE connection refused",
		"Task not found", // Even clean messages should use fallback for 500
	}

	for _, msg := range dangerousMessages {
		t.Run("should_hide_"+msg[:min(len(msg), 15)], func(t *testing.T) {
			mock := NewMockClient()
			mock.OnGet("/test", http.StatusInternalServerError, map[string]interface{}{
				"message": msg,
			})

			resp, _ := mock.Get("/test")
			err := handleResponse(resp, nil)

			apiErr, ok := err.(*APIError)
			if !ok {
				t.Fatalf("Expected *APIError, got %T", err)
			}

			// 500 errors should ALWAYS use fallback, NEVER contain raw message
			if apiErr.Message == msg {
				t.Errorf("Server error should NOT expose API message: %q", msg)
			}

			// Should use user-friendly fallback
			expectedFallback := "Something went wrong, please try again later"
			if apiErr.Message != expectedFallback {
				t.Errorf("Expected fallback message, got: %q", apiErr.Message)
			}
		})
	}
}

func TestHandleResponse_ErrorTypes(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		checkFunc  func(error) bool
	}{
		{"401 Unauthorized", http.StatusUnauthorized, IsUnauthorized},
		{"402 Payment Required", http.StatusPaymentRequired, IsPaymentRequired},
		{"404 Not Found", http.StatusNotFound, IsNotFound},
		{"429 Rate Limited", http.StatusTooManyRequests, IsRateLimited},
		{"409 Conflict", http.StatusConflict, IsConflict},
		{"500 Server Error", http.StatusInternalServerError, IsServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockClient()
			mock.OnGet("/test", tt.statusCode, map[string]string{"message": "error"})

			resp, _ := mock.Get("/test")
			err := handleResponse(resp, nil)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !tt.checkFunc(err) {
				t.Errorf("Check function should return true for %s", tt.name)
			}
		})
	}
}

func TestHandleResponse_ValidationErrors(t *testing.T) {
	mock := NewMockClient()
	mock.OnGet("/tasks", http.StatusUnprocessableEntity, map[string]interface{}{
		"message": "Validation failed",
		"errors": map[string][]string{
			"title": {"Title is required", "Title must be at least 3 characters"},
		},
	})

	resp, _ := mock.Get("/tasks")
	err := handleResponse(resp, nil)

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("Expected *APIError, got %T", err)
	}

	// Should return first validation error
	if apiErr.Message != "Title is required" {
		t.Errorf("Expected 'Title is required', got %q", apiErr.Message)
	}

	// Should be wrapped as ErrValidation
	if !errors.Is(err, ErrValidation) {
		t.Error("Expected error to unwrap to ErrValidation")
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	apiErr := &APIError{
		StatusCode: 402,
		Type:       ErrPaymentRequired,
		Message:    "test",
	}

	if !errors.Is(apiErr, ErrPaymentRequired) {
		t.Error("APIError should unwrap to its Type")
	}
}

func TestSanitizeMessage(t *testing.T) {
	tests := []struct {
		input    string
		fallback string
		expected string
	}{
		{"", "fallback", "fallback"},
		{"Clean message", "fallback", "Clean message"},
		{"SQLSTATE error", "fallback", "fallback"},
		{"firstOrFail not found", "fallback", "fallback"},
		{"Error at vendor/laravel", "fallback", "fallback"},
		{strings.Repeat("x", 500), "fallback", "fallback"}, // Too long
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(len(tt.input), 20)], func(t *testing.T) {
			result := sanitizeMessage(tt.input, tt.fallback)
			if result != tt.expected {
				t.Errorf("sanitizeMessage(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
