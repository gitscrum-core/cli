package services

import (
	"fmt"
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// Task represents a task/issue in the system
type Task struct {
	UUID        string    `json:"uuid"`
	Code        string    `json:"code"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Priority    int       `json:"priority"`
	Effort      int       `json:"effort"`
	DueDate     string    `json:"due_date"`
	StartDate   string    `json:"start_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relationships
	Status   Status  `json:"status"`
	Assignee User    `json:"assignee"`
	Author   User    `json:"author"`
	Project  Project `json:"project"`
	Labels   []Label `json:"labels"`

	// Computed
	TimeSpent float64 `json:"time_spent"`
	Progress  int     `json:"progress"`
}

// TasksListOptions for filtering task lists
type TasksListOptions struct {
	Project  string
	Assignee string
	Status   string
	Filter   string // my, today, overdue
	Limit    int
	Page     int
}

// CreateTaskInput for creating a new task
type CreateTaskInput struct {
	Title       string
	Description string
	Project     string
	Assignee    string
	Type        string
	Priority    string
	Parent      string
}

// UpdateTaskInput for updating a task
type UpdateTaskInput struct {
	Title       string
	Description string
	Status      string
	Assignee    string
	Priority    string
}

// TasksService handles task operations
type TasksService interface {
	// List returns tasks matching the given options
	List(opts TasksListOptions) ([]Task, error)

	// ListToday returns tasks due today for the current user
	ListToday() ([]Task, error)

	// Get returns a task by code (e.g., "GS-123")
	Get(code string) (*Task, error)

	// Create creates a new task
	Create(input CreateTaskInput) (*Task, error)

	// Update updates an existing task
	Update(code string, input UpdateTaskInput) (*Task, error)

	// Complete marks a task as complete
	Complete(code string) error

	// Assign assigns a user to a task
	Assign(code, username string) error

	// Move moves a task to another project
	Move(code, toProject string) (*Task, error)

	// Duplicate creates a copy of a task
	Duplicate(code, toProject string, withSubtasks bool) (*Task, error)

	// ListSubtasks returns subtasks for a task
	ListSubtasks(code string) ([]Task, error)
}

type tasksService struct {
	client HTTPClient
}

func newTasksService(client *api.Client) TasksService {
	return &tasksService{client: client}
}

func (s *tasksService) List(opts TasksListOptions) ([]Task, error) {
	path := "/tasks?"

	if opts.Project != "" {
		path += "project=" + opts.Project + "&"
	}
	if opts.Assignee != "" {
		path += "assignee=" + opts.Assignee + "&"
	}
	if opts.Status != "" {
		path += "status=" + opts.Status + "&"
	}
	if opts.Filter != "" {
		path += "filter=" + opts.Filter + "&"
	}
	if opts.Limit > 0 {
		path += fmt.Sprintf("limit=%d&", opts.Limit)
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *tasksService) ListToday() ([]Task, error) {
	resp, err := s.client.Get("/tasks?filter=today")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *tasksService) Get(code string) (*Task, error) {
	resp, err := s.client.Get("/tasks/by-code/" + code)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *tasksService) Create(input CreateTaskInput) (*Task, error) {
	body := map[string]interface{}{
		"title":   input.Title,
		"project": input.Project,
	}

	if input.Description != "" {
		body["description"] = input.Description
	}
	if input.Assignee != "" {
		body["assignee"] = input.Assignee
	}
	if input.Type != "" {
		body["type"] = input.Type
	}
	if input.Priority != "" {
		body["priority"] = input.Priority
	}
	if input.Parent != "" {
		body["parent"] = input.Parent
	}

	resp, err := s.client.Post("/tasks", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *tasksService) Update(code string, input UpdateTaskInput) (*Task, error) {
	body := map[string]interface{}{}

	if input.Title != "" {
		body["title"] = input.Title
	}
	if input.Description != "" {
		body["description"] = input.Description
	}
	if input.Status != "" {
		body["workflow_stage"] = input.Status
	}
	if input.Assignee != "" {
		body["assignee"] = input.Assignee
	}

	resp, err := s.client.Put("/tasks/by-code/"+code, body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *tasksService) Complete(code string) error {
	body := map[string]interface{}{
		"workflow_stage": "done",
	}

	resp, err := s.client.Put("/tasks/by-code/"+code, body)
	if err != nil {
		return err
	}
	var result struct {
		Data Task `json:"data"`
	}
	return handleResponse(resp, &result)
}

func (s *tasksService) Assign(code, username string) error {
	body := map[string]interface{}{
		"assignee": username,
	}

	resp, err := s.client.Put("/tasks/by-code/"+code, body)
	if err != nil {
		return err
	}
	var result struct {
		Data Task `json:"data"`
	}
	return handleResponse(resp, &result)
}

func (s *tasksService) Move(code, toProject string) (*Task, error) {
	// First get task by code to get UUID
	task, err := s.Get(code)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"project": toProject,
	}

	// Use UUID endpoint
	resp, err := s.client.Post("/tasks/"+task.UUID+"/move", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *tasksService) Duplicate(code, toProject string, withSubtasks bool) (*Task, error) {
	// First get task by code to get UUID
	task, err := s.Get(code)
	if err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"with_subtasks": withSubtasks,
	}
	if toProject != "" {
		body["project"] = toProject
	}

	// Use UUID endpoint
	resp, err := s.client.Post("/tasks/"+task.UUID+"/duplicate", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *tasksService) ListSubtasks(code string) ([]Task, error) {
	resp, err := s.client.Get("/tasks/by-code/" + code + "/children")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Task `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

