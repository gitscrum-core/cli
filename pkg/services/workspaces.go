package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// WorkspaceDetails represents detailed workspace info
type WorkspaceDetails struct {
	Workspace
	ProjectsCount int    `json:"projects_count"`
	MembersCount  int    `json:"members_count"`
	Plan          string `json:"plan"`
}

// WorkspaceMember represents a workspace member
type WorkspaceMember struct {
	User
	Role     string `json:"role"` // owner, admin, member
	JoinedAt string `json:"joined_at"`
}

// WorkspacesService handles workspace operations
type WorkspacesService interface {
	// List returns all workspaces
	List() ([]Workspace, error)

	// Get returns a workspace by slug
	Get(slug string) (*WorkspaceDetails, error)

	// ListMembers returns workspace members
	ListMembers(slug string) ([]WorkspaceMember, error)

	// ListProjects returns projects in a workspace
	ListProjects(slug string) ([]Project, error)
}

type workspacesService struct {
	client *api.Client
}

func newWorkspacesService(client *api.Client) WorkspacesService {
	return &workspacesService{client: client}
}

func (s *workspacesService) List() ([]Workspace, error) {
	resp, err := s.client.Get("/workspaces")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Workspace `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *workspacesService) Get(slug string) (*WorkspaceDetails, error) {
	resp, err := s.client.Get("/workspaces/" + slug)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data WorkspaceDetails `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *workspacesService) ListMembers(slug string) ([]WorkspaceMember, error) {
	resp, err := s.client.Get("/workspaces/" + slug + "/members")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []WorkspaceMember `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *workspacesService) ListProjects(slug string) ([]Project, error) {
	resp, err := s.client.Get("/workspaces/" + slug + "/projects")
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
