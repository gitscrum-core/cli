package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// HTTPClient interface for mocking
type HTTPClient interface {
	Get(path string) (*http.Response, error)
	Post(path string, body interface{}) (*http.Response, error)
	Put(path string, body interface{}) (*http.Response, error)
	Delete(path string) (*http.Response, error)
}

// MockClient for testing
type MockClient struct {
	// Responses to return for each method
	GetResponses    map[string]*MockResponse
	PostResponses   map[string]*MockResponse
	PutResponses    map[string]*MockResponse
	DeleteResponses map[string]*MockResponse

	// Default response if path not found
	DefaultResponse *MockResponse

	// Track calls for assertions
	Calls []MockCall
}

// MockResponse represents a mocked HTTP response
type MockResponse struct {
	StatusCode int
	Body       interface{}
	Error      error
}

// MockCall tracks a method call
type MockCall struct {
	Method string
	Path   string
	Body   interface{}
}

// NewMockClient creates a new mock client
func NewMockClient() *MockClient {
	return &MockClient{
		GetResponses:    make(map[string]*MockResponse),
		PostResponses:   make(map[string]*MockResponse),
		PutResponses:    make(map[string]*MockResponse),
		DeleteResponses: make(map[string]*MockResponse),
		DefaultResponse: &MockResponse{StatusCode: http.StatusOK, Body: map[string]interface{}{"data": nil}},
	}
}

// OnGet sets the response for a GET request
func (m *MockClient) OnGet(path string, statusCode int, body interface{}) *MockClient {
	m.GetResponses[path] = &MockResponse{StatusCode: statusCode, Body: body}
	return m
}

// OnGetError sets an error for a GET request
func (m *MockClient) OnGetError(path string, err error) *MockClient {
	m.GetResponses[path] = &MockResponse{Error: err}
	return m
}

// OnPost sets the response for a POST request
func (m *MockClient) OnPost(path string, statusCode int, body interface{}) *MockClient {
	m.PostResponses[path] = &MockResponse{StatusCode: statusCode, Body: body}
	return m
}

// OnPut sets the response for a PUT request
func (m *MockClient) OnPut(path string, statusCode int, body interface{}) *MockClient {
	m.PutResponses[path] = &MockResponse{StatusCode: statusCode, Body: body}
	return m
}

// OnDelete sets the response for a DELETE request
func (m *MockClient) OnDelete(path string, statusCode int, body interface{}) *MockClient {
	m.DeleteResponses[path] = &MockResponse{StatusCode: statusCode, Body: body}
	return m
}

// OnProRequired sets a 409 PRO required response
func (m *MockClient) OnProRequired(path string) *MockClient {
	m.GetResponses[path] = &MockResponse{
		StatusCode: http.StatusConflict,
		Body:       map[string]string{"message": "PRO subscription required"},
	}
	return m
}

// OnUnauthorized sets a 401 response
func (m *MockClient) OnUnauthorized(path string) *MockClient {
	m.GetResponses[path] = &MockResponse{
		StatusCode: http.StatusUnauthorized,
		Body:       map[string]string{"message": "Unauthorized"},
	}
	return m
}

func (m *MockClient) Get(path string) (*http.Response, error) {
	m.Calls = append(m.Calls, MockCall{Method: "GET", Path: path})
	return m.respond(m.GetResponses, path)
}

func (m *MockClient) Post(path string, body interface{}) (*http.Response, error) {
	m.Calls = append(m.Calls, MockCall{Method: "POST", Path: path, Body: body})
	return m.respond(m.PostResponses, path)
}

func (m *MockClient) Put(path string, body interface{}) (*http.Response, error) {
	m.Calls = append(m.Calls, MockCall{Method: "PUT", Path: path, Body: body})
	return m.respond(m.PutResponses, path)
}

func (m *MockClient) Delete(path string) (*http.Response, error) {
	m.Calls = append(m.Calls, MockCall{Method: "DELETE", Path: path})
	return m.respond(m.DeleteResponses, path)
}

func (m *MockClient) respond(responses map[string]*MockResponse, path string) (*http.Response, error) {
	resp, ok := responses[path]
	if !ok {
		resp = m.DefaultResponse
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	body, _ := json.Marshal(resp.Body)
	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(bytes.NewReader(body)),
	}, nil
}

// AssertCalled checks if a path was called with the given method
func (m *MockClient) AssertCalled(method, path string) bool {
	for _, call := range m.Calls {
		if call.Method == method && call.Path == path {
			return true
		}
	}
	return false
}

// CallCount returns the number of calls to a path
func (m *MockClient) CallCount(method, path string) int {
	count := 0
	for _, call := range m.Calls {
		if call.Method == method && call.Path == path {
			count++
		}
	}
	return count
}
