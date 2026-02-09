package services

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Sentinel errors for type checking
var (
	ErrUnauthorized    = errors.New("unauthorized")
	ErrPaymentRequired = errors.New("payment_required")
	ErrForbidden       = errors.New("forbidden")
	ErrNotFound        = errors.New("not_found")
	ErrConflict        = errors.New("conflict")
	ErrValidation      = errors.New("validation")
	ErrRateLimited     = errors.New("rate_limited")
	ErrServerError     = errors.New("server_error")
)

// Default user-friendly messages (fallbacks only)
var defaultMessages = map[error]string{
	ErrUnauthorized:    "Please run 'gitscrum auth login' to authenticate",
	ErrPaymentRequired: "This feature requires a PRO subscription",
	ErrForbidden:       "You don't have permission to perform this action",
	ErrNotFound:        "The requested resource was not found",
	ErrConflict:        "The resource already exists or was modified",
	ErrValidation:      "Invalid input provided",
	ErrRateLimited:     "Too many requests, please try again later",
	ErrServerError:     "Something went wrong, please try again later",
}

// APIError represents an error from the API with sanitized message
type APIError struct {
	StatusCode int                 `json:"-"`
	Type       error               `json:"-"`
	Message    string              `json:"message"`
	Errors     map[string][]string `json:"errors,omitempty"`
	Feature    string              `json:"feature,omitempty"`
	UpgradeURL string              `json:"upgrade_url,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.Type
}

// handleResponse processes the HTTP response and returns user-friendly errors
func handleResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	// Success responses
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if target == nil {
			return nil
		}
		return json.NewDecoder(resp.Body).Decode(target)
	}

	// Read error body
	body, _ := io.ReadAll(resp.Body)

	// Parse API response
	var apiResp struct {
		Message    string              `json:"message"`
		Error      string              `json:"error"`
		Errors     map[string][]string `json:"errors"`
		Feature    string              `json:"feature"`
		UpgradeURL string              `json:"upgrade_url"`
	}
	json.Unmarshal(body, &apiResp)

	// Create error based on status code
	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		Errors:     apiResp.Errors,
		Feature:    apiResp.Feature,
		UpgradeURL: apiResp.UpgradeURL,
	}

	switch resp.StatusCode {
	case http.StatusUnauthorized: // 401
		apiErr.Type = ErrUnauthorized
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrUnauthorized])

	case http.StatusPaymentRequired: // 402
		apiErr.Type = ErrPaymentRequired
		// Use API message directly - it contains the upgrade URL
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrPaymentRequired])

	case http.StatusForbidden: // 403
		apiErr.Type = ErrForbidden
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrForbidden])

	case http.StatusNotFound: // 404
		apiErr.Type = ErrNotFound
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrNotFound])

	case http.StatusConflict: // 409
		apiErr.Type = ErrConflict
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrConflict])

	case http.StatusUnprocessableEntity: // 422
		apiErr.Type = ErrValidation
		if len(apiResp.Errors) > 0 {
			apiErr.Message = formatValidationErrors(apiResp.Errors)
		} else {
			apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrValidation])
		}

	case http.StatusTooManyRequests: // 429
		apiErr.Type = ErrRateLimited
		apiErr.Message = sanitizeMessage(apiResp.Message, defaultMessages[ErrRateLimited])

	default:
		if resp.StatusCode >= 500 {
			// ALWAYS use default message for server errors - NEVER expose internal details
			apiErr.Type = ErrServerError
			apiErr.Message = defaultMessages[ErrServerError]
		} else {
			// For other unknown status codes, try to sanitize
			apiErr.Type = errors.New("unknown_error")
			apiErr.Message = sanitizeMessage(apiResp.Message, "An unexpected error occurred")
		}
	}

	return apiErr
}

// sanitizeMessage ensures the message is safe to display to users
// Removes stack traces, SQL errors, file paths, and other internal details
func sanitizeMessage(apiMessage, fallback string) string {
	if apiMessage == "" {
		return fallback
	}

	// Patterns that indicate raw/internal errors that should NOT be shown
	unsafePatterns := []string{
		"SQLSTATE",
		"firstOrFail",
		"Stack trace",
		"Exception",
		"at line",
		"vendor/",
		"app/",
		"/var/www",
		"\\Users\\",
		"PDOException",
		"QueryException",
		"ModelNotFoundException",
		"No query results for model",
		"Undefined variable",
		"Call to undefined",
		"syntax error",
		"Array to string conversion",
	}

	for _, pattern := range unsafePatterns {
		if strings.Contains(apiMessage, pattern) {
			return fallback
		}
	}

	// Remove any remaining file paths
	filePathRegex := regexp.MustCompile(`[A-Za-z]:\\[^\s]+|/[a-z]+/[^\s]+`)
	if filePathRegex.MatchString(apiMessage) {
		return fallback
	}

	// Remove any long error traces (multiple lines)
	if strings.Count(apiMessage, "\n") > 2 {
		// Take only the first line if it's clean
		firstLine := strings.Split(apiMessage, "\n")[0]
		if len(firstLine) > 0 && len(firstLine) < 200 {
			return firstLine
		}
		return fallback
	}

	// Limit message length
	if len(apiMessage) > 300 {
		return fallback
	}

	return apiMessage
}

// formatValidationErrors formats validation errors for display
func formatValidationErrors(errs map[string][]string) string {
	for field, msgs := range errs {
		if len(msgs) > 0 {
			// Return first error, sanitized
			return sanitizeMessage(msgs[0], field+": invalid value")
		}
	}
	return defaultMessages[ErrValidation]
}

// IsPaymentRequired checks if the error indicates PRO subscription is needed (402)
func IsPaymentRequired(err error) bool {
	return errors.Is(err, ErrPaymentRequired)
}

// IsUnauthorized checks if the error indicates authentication is needed (401)
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsNotFound checks if the error indicates resource not found (404)
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsRateLimited checks if the error indicates rate limiting (429)
func IsRateLimited(err error) bool {
	return errors.Is(err, ErrRateLimited)
}

// IsConflict checks if the error indicates a resource conflict (409)
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsServerError checks if the error is a server-side error (500+)
func IsServerError(err error) bool {
	return errors.Is(err, ErrServerError)
}
