package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// ProjectDetails represents detailed project info
type ProjectDetails struct {
	Project
	TasksCount      int     `json:"tasks_count"`
	CompletedCount  int     `json:"completed_count"`
	MembersCount    int     `json:"members_count"`
	Progress        float64 `json:"progress"`
	CurrentSprint   *Sprint `json:"current_sprint"`
}

// ProjectMember represents a project team member
type ProjectMember struct {
	User
	Role string `json:"role"`
}

// CreateProjectInput for creating a new project
type CreateProjectInput struct {
	Name        string
	Description string
	Workspace   string
}

// ProjectsService handles project operations
type ProjectsService interface {
	// List returns all projects
	List() ([]Project, error)

	// Get returns a project by slug
	Get(slug string) (*ProjectDetails, error)

	// Create creates a new project
	Create(input CreateProjectInput) (*Project, error)

	// ListMembers returns project team members
	ListMembers(slug string) ([]ProjectMember, error)

	// ListWorkflows returns project workflow stages
	ListWorkflows(slug string) ([]Status, error)
}

type projectsService struct {
	client *api.Client
}

func newProjectsService(client *api.Client) ProjectsService {
	return &projectsService{client: client}
}

func (s *projectsService) List() ([]Project, error) {
	resp, err := s.client.Get("/projects")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Project `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *projectsService) Get(slug string) (*ProjectDetails, error) {
	resp, err := s.client.Get("/projects/" + slug)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data ProjectDetails `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *projectsService) Create(input CreateProjectInput) (*Project, error) {
	body := map[string]interface{}{
		"name": input.Name,
	}
	if input.Description != "" {
		body["description"] = input.Description
	}
	if input.Workspace != "" {
		body["workspace"] = input.Workspace
	}

	resp, err := s.client.Post("/projects", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Project `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *projectsService) ListMembers(slug string) ([]ProjectMember, error) {
	resp, err := s.client.Get("/projects/" + slug + "/members")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []ProjectMember `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *projectsService) ListWorkflows(slug string) ([]Status, error) {
	resp, err := s.client.Get("/projects/" + slug + "/workflow-stages")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Status `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

