// Package errors provides user-friendly error handling for GitScrum CLI
package errors

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
)

// Common error types
var (
	ErrNotAuthenticated = &CLIError{
		Message:    "You are not logged in",
		Suggestion: "Run 'gitscrum auth login' to authenticate",
		Code:       "AUTH_REQUIRED",
	}

	ErrNoWorkspace = &CLIError{
		Message:    "No workspace configured",
		Suggestion: "Run 'gitscrum config set workspace <slug>' to set a default workspace",
		Code:       "NO_WORKSPACE",
	}

	ErrNoProject = &CLIError{
		Message:    "No project specified",
		Suggestion: "Use --project flag or run 'gitscrum config set project <slug>'",
		Code:       "NO_PROJECT",
	}

	ErrNotInGitRepo = &CLIError{
		Message:    "Not in a git repository",
		Suggestion: "Navigate to a git repository or use explicit flags",
		Code:       "NOT_GIT_REPO",
	}

	ErrNoTaskCode = &CLIError{
		Message:    "Could not detect task code from branch",
		Suggestion: "Specify the task code explicitly, e.g. 'gitscrum tasks view GS-123'",
		Code:       "NO_TASK_CODE",
	}
)

// CLIError represents a user-friendly CLI error
type CLIError struct {
	Message    string
	Suggestion string
	Code       string
	Cause      error
}

func (e *CLIError) Error() string {
	return e.Message
}

func (e *CLIError) Unwrap() error {
	return e.Cause
}

// WithCause adds the underlying error
func (e *CLIError) WithCause(err error) *CLIError {
	return &CLIError{
		Message:    e.Message,
		Suggestion: e.Suggestion,
		Code:       e.Code,
		Cause:      err,
	}
}

// Print displays the error in a user-friendly format
func (e *CLIError) Print() {
	red := color.New(color.FgRed, color.Bold)
	yellow := color.New(color.FgYellow)

	red.Printf("Error: %s\n", e.Message)
	if e.Suggestion != "" {
		yellow.Printf("Hint: %s\n", e.Suggestion)
	}
}

// HandleAPIError converts HTTP errors to friendly messages
func HandleAPIError(resp *http.Response, body string) *CLIError {
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return &CLIError{
			Message:    "Authentication failed or session expired",
			Suggestion: "Run 'gitscrum auth login' to re-authenticate",
			Code:       "UNAUTHORIZED",
		}
	case http.StatusForbidden:
		return &CLIError{
			Message:    "You don't have permission to perform this action",
			Suggestion: "Check your role in the workspace/project settings",
			Code:       "FORBIDDEN",
		}
	case http.StatusNotFound:
		return &CLIError{
			Message:    "Resource not found",
			Suggestion: "Check that the task/project/sprint exists and you have access",
			Code:       "NOT_FOUND",
		}
	case http.StatusUnprocessableEntity:
		return parseValidationError(body)
	case http.StatusTooManyRequests:
		return &CLIError{
			Message:    "Rate limit exceeded",
			Suggestion: "Wait a moment and try again",
			Code:       "RATE_LIMITED",
		}
	case http.StatusInternalServerError, http.StatusBadGateway, http.StatusServiceUnavailable:
		return &CLIError{
			Message:    "GitScrum service is temporarily unavailable",
			Suggestion: "Try again in a few moments. If the problem persists, check status.gitscrum.com",
			Code:       "SERVER_ERROR",
		}
	default:
		return &CLIError{
			Message:    fmt.Sprintf("Request failed (status %d)", resp.StatusCode),
			Suggestion: "Try again or run with --debug for more details",
			Code:       "REQUEST_FAILED",
		}
	}
}

// parseValidationError extracts validation errors from API response
func parseValidationError(body string) *CLIError {
	// Try to extract meaningful message from Laravel validation errors
	// Example: {"message":"The given data was invalid.","errors":{"title":["required"]}}
	
	if strings.Contains(body, "title") && strings.Contains(body, "required") {
		return &CLIError{
			Message:    "Title is required",
			Suggestion: "Provide a title using --title or -t flag",
			Code:       "VALIDATION_ERROR",
		}
	}
	
	if strings.Contains(body, "project") && strings.Contains(body, "required") {
		return ErrNoProject
	}

	return &CLIError{
		Message:    "Validation failed",
		Suggestion: "Check your input and try again",
		Code:       "VALIDATION_ERROR",
	}
}

// Wrap creates a CLIError from a generic error
func Wrap(err error, message string) *CLIError {
	if err == nil {
		return nil
	}

	// Check if already a CLIError
	if cliErr, ok := err.(*CLIError); ok {
		return cliErr
	}

	return &CLIError{
		Message: message,
		Cause:   err,
	}
}

// WrapWithSuggestion creates a CLIError with a suggestion
func WrapWithSuggestion(err error, message, suggestion string) *CLIError {
	if err == nil {
		return nil
	}

	return &CLIError{
		Message:    message,
		Suggestion: suggestion,
		Cause:      err,
	}
}

// New creates a new CLIError
func New(message string) *CLIError {
	return &CLIError{Message: message}
}

// NewWithSuggestion creates a new CLIError with suggestion
func NewWithSuggestion(message, suggestion string) *CLIError {
	return &CLIError{
		Message:    message,
		Suggestion: suggestion,
	}
}

// FormatError returns a formatted error string for printing
func FormatError(err error) string {
	if cliErr, ok := err.(*CLIError); ok {
		if cliErr.Suggestion != "" {
			return fmt.Sprintf("Error: %s\nHint: %s", cliErr.Message, cliErr.Suggestion)
		}
		return fmt.Sprintf("Error: %s", cliErr.Message)
	}
	return fmt.Sprintf("Error: %s", err.Error())
}
