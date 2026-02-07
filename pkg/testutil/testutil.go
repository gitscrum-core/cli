// Package testutil provides shared utilities for CLI testing
package testutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
)

// MockAPIServer creates a test HTTP server with configurable responses
type MockAPIServer struct {
	*httptest.Server
	Responses map[string]MockResponse
}

// MockResponse defines a mock API response
type MockResponse struct {
	StatusCode int
	Body       interface{}
	Headers    map[string]string
}

// NewMockAPIServer creates a new mock server
func NewMockAPIServer(responses map[string]MockResponse) *MockAPIServer {
	mock := &MockAPIServer{
		Responses: responses,
	}

	mock.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if resp, ok := mock.Responses[path]; ok {
			// Set headers
			for k, v := range resp.Headers {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "application/json")

			// Set status code
			if resp.StatusCode > 0 {
				w.WriteHeader(resp.StatusCode)
			}

			// Write body
			if resp.Body != nil {
				json.NewEncoder(w).Encode(resp.Body)
			}
			return
		}

		// Default 404
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))

	return mock
}

// Client returns an API client configured for this mock server
func (m *MockAPIServer) Client() *api.Client {
	return api.NewClient(m.Server.URL, &auth.Token{
		AccessToken: "test-token",
	})
}

// UnauthenticatedClient returns an API client without auth
func (m *MockAPIServer) UnauthenticatedClient() *api.Client {
	return api.NewClient(m.Server.URL, nil)
}

// Close shuts down the mock server
func (m *MockAPIServer) Close() {
	m.Server.Close()
}

// CaptureOutput captures stdout/stderr during test execution
type CaptureOutput struct {
	oldStdout *os.File
	oldStderr *os.File
	readPipe  *os.File
	writePipe *os.File
}

// NewCaptureOutput starts capturing stdout
func NewCaptureOutput() *CaptureOutput {
	r, w, _ := os.Pipe()
	return &CaptureOutput{
		oldStdout: os.Stdout,
		oldStderr: os.Stderr,
		readPipe:  r,
		writePipe: w,
	}
}

// Start begins capturing output
func (c *CaptureOutput) Start() {
	os.Stdout = c.writePipe
	os.Stderr = c.writePipe
}

// Stop stops capturing and returns the captured output
func (c *CaptureOutput) Stop() string {
	c.writePipe.Close()
	os.Stdout = c.oldStdout
	os.Stderr = c.oldStderr

	var buf bytes.Buffer
	io.Copy(&buf, c.readPipe)
	c.readPipe.Close()

	return buf.String()
}

// TempDir creates a temporary directory for testing
func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "gitscrum-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() {
		os.RemoveAll(dir)
	})
	return dir
}

// WriteFile writes content to a file in the given directory
func WriteFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return path
}

// AssertNoError fails the test if err is not nil
func AssertNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError fails the test if err is nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// AssertEqual fails if got != want
func AssertEqual(t *testing.T, got, want interface{}) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertContains fails if s does not contain substr
func AssertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !bytes.Contains([]byte(s), []byte(substr)) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}
