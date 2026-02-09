package services

import "time"

// Common types used across services

// User represents a user in the system
type User struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

// Status represents a workflow status
type Status struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Color string `json:"color"`
	Slug  string `json:"slug"`
}

// Project represents a project
type Project struct {
	UUID        string `json:"uuid"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// Workspace represents a workspace
type Workspace struct {
	UUID string `json:"uuid"`
	Slug string `json:"slug"`
	Name string `json:"name"`
}

// Label represents a task label
type Label struct {
	UUID  string `json:"uuid"`
	Title string `json:"title"`
	Color string `json:"color"`
}

// Pagination for list responses
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
}

// ListResponse wraps paginated list responses
type ListResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"meta"`
}

// SingleResponse wraps single item responses
type SingleResponse[T any] struct {
	Data T `json:"data"`
}

// TimeEntry represents a time tracking entry
type TimeEntry struct {
	UUID      string    `json:"uuid"`
	Start     time.Time `json:"start"`
	End       time.Time `json:"end"`
	Duration  float64   `json:"duration"`
	Comment   string    `json:"comment"`
	Billable  bool      `json:"billable"`
	CreatedAt time.Time `json:"created_at"`
}

// Branch represents a linked git branch
type Branch struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	CreatedAt string `json:"created_at"`
}

// PullRequest represents a linked pull request
type PullRequest struct {
	UUID   string `json:"uuid"`
	Title  string `json:"title"`
	Number int    `json:"number"`
	State  string `json:"state"`
	URL    string `json:"url"`
}
