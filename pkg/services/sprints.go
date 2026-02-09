package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// Sprint represents a sprint
type Sprint struct {
	UUID      string    `json:"uuid"`
	Title     string    `json:"title"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Status    string    `json:"status"` // planned, active, completed
	Goal      string    `json:"goal"`
	Progress  float64   `json:"progress"`
}

// SprintStats represents sprint statistics
type SprintStats struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	TotalPoints    int     `json:"total_points"`
	BurndownData   []Point `json:"burndown"`
}

// Point for burndown charts
type Point struct {
	Date   string  `json:"date"`
	Ideal  float64 `json:"ideal"`
	Actual float64 `json:"actual"`
}

// CreateSprintInput for creating a new sprint
type CreateSprintInput struct {
	Title     string
	StartDate time.Time
	EndDate   time.Time
	Goal      string
	Project   string
}

// SprintsService handles sprint operations
type SprintsService interface {
	// List returns all sprints for a project
	List(project string) ([]Sprint, error)

	// GetActive returns the active sprint for a project
	GetActive(project string) (*Sprint, error)

	// Get returns a sprint by ID
	Get(uuid string) (*Sprint, error)

	// Create creates a new sprint
	Create(input CreateSprintInput) (*Sprint, error)

	// Start activates a sprint
	Start(uuid string) (*Sprint, error)

	// Complete ends a sprint
	Complete(uuid string) (*Sprint, error)

	// Stats returns sprint statistics
	Stats(uuid string) (*SprintStats, error)

	// ListTasks returns tasks in a sprint
	ListTasks(uuid string) ([]Task, error)
}

type sprintsService struct {
	client *api.Client
}

func newSprintsService(client *api.Client) SprintsService {
	return &sprintsService{client: client}
}

func (s *sprintsService) List(project string) ([]Sprint, error) {
	path := "/sprints"
	if project != "" {
		path += "?project=" + project
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *sprintsService) GetActive(project string) (*Sprint, error) {
	path := "/sprints?state=active"
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	// Return first active sprint, or nil if none
	if len(result.Data) == 0 {
		return nil, nil
	}

	return &result.Data[0], nil
}

func (s *sprintsService) Get(uuid string) (*Sprint, error) {
	resp, err := s.client.Get("/sprints/" + uuid)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *sprintsService) Create(input CreateSprintInput) (*Sprint, error) {
	body := map[string]interface{}{
		"title":      input.Title,
		"start_date": input.StartDate.Format("2006-01-02"),
		"end_date":   input.EndDate.Format("2006-01-02"),
	}
	if input.Goal != "" {
		body["goal"] = input.Goal
	}
	if input.Project != "" {
		body["project"] = input.Project
	}

	resp, err := s.client.Post("/sprints", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *sprintsService) Start(slug string) (*Sprint, error) {
	// Use PUT /sprints/{slug} with status active
	body := map[string]interface{}{
		"status": "active",
	}
	resp, err := s.client.Put("/sprints/"+slug, body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *sprintsService) Complete(slug string) (*Sprint, error) {
	// Use PUT /sprints/{slug} with status completed
	body := map[string]interface{}{
		"status": "completed",
	}
	resp, err := s.client.Put("/sprints/"+slug, body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Sprint `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *sprintsService) Stats(uuid string) (*SprintStats, error) {
	resp, err := s.client.Get("/sprints/" + uuid + "/stats")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data SprintStats `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *sprintsService) ListTasks(uuid string) ([]Task, error) {
	// Use tasks endpoint with sprint filter and pagination
	path := "/tasks?sprint_slug=" + uuid + "&per_page=100"
	
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

