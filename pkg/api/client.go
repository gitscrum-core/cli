// Package api provides HTTP client for GitScrum API
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gitscrum-core/cli/pkg/auth"
)

// Client handles HTTP requests to GitScrum API
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Token      *auth.Token
}

// NewClient creates a new API client
func NewClient(baseURL string, token *auth.Token) *Client {
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Request makes an HTTP request to the API
func (c *Client) Request(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "GitScrum-CLI/1.0")

	// Add auth token if available
	if c.Token != nil && c.Token.AccessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token.AccessToken)
	}

	return c.HTTPClient.Do(req)
}

// Get makes a GET request
func (c *Client) Get(path string) (*http.Response, error) {
	return c.Request(http.MethodGet, path, nil)
}

// Post makes a POST request
func (c *Client) Post(path string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPost, path, body)
}

// Put makes a PUT request
func (c *Client) Put(path string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPut, path, body)
}

// Patch makes a PATCH request
func (c *Client) Patch(path string, body interface{}) (*http.Response, error) {
	return c.Request(http.MethodPatch, path, body)
}

// Delete makes a DELETE request
func (c *Client) Delete(path string) (*http.Response, error) {
	return c.Request(http.MethodDelete, path, nil)
}

// DecodeResponse decodes JSON response body into target struct
func DecodeResponse(resp *http.Response, target interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

// APIError represents an error from the API
type APIError struct {
	StatusCode int
	Message    string
	Errors     map[string][]string `json:"errors,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Message)
}
