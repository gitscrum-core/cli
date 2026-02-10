package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// WikiPage represents a wiki page
type WikiPage struct {
	UUID      string    `json:"uuid"`
	Title     string    `json:"title"`
	Slug      string    `json:"slug"`
	Content   string    `json:"content"`
	Author    User      `json:"author"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// WikiService handles wiki operations
type WikiService interface {
	// List returns all wiki pages for a project
	List(project string) ([]WikiPage, error)

	// Get returns a wiki page by slug
	Get(project, slug string) (*WikiPage, error)

	// Create creates a new wiki page
	Create(project, title, content string) (*WikiPage, error)

	// Update updates a wiki page
	Update(project, slug, title, content string) (*WikiPage, error)
}

type wikiService struct {
	client *api.Client
}

func newWikiService(client *api.Client) WikiService {
	return &wikiService{client: client}
}

func (s *wikiService) List(project string) ([]WikiPage, error) {
	resp, err := s.client.Get("/projects/" + project + "/wiki")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []WikiPage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *wikiService) Get(project, slug string) (*WikiPage, error) {
	resp, err := s.client.Get("/projects/" + project + "/wiki/" + slug)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data WikiPage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *wikiService) Create(project, title, content string) (*WikiPage, error) {
	body := map[string]interface{}{
		"title":   title,
		"content": content,
	}

	resp, err := s.client.Post("/projects/"+project+"/wiki", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data WikiPage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *wikiService) Update(project, slug, title, content string) (*WikiPage, error) {
	body := map[string]interface{}{}
	if title != "" {
		body["title"] = title
	}
	if content != "" {
		body["content"] = content
	}

	resp, err := s.client.Put("/projects/"+project+"/wiki/"+slug, body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data WikiPage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
