package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// Client represents a client/customer
type Client struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
	Status  string `json:"status"` // active, inactive
}

// ClientsService handles client operations
type ClientsService interface {
	// List returns all clients
	List() ([]Client, error)

	// Get returns a client by ID
	Get(uuid string) (*Client, error)

	// Create creates a new client
	Create(name, email, company string) (*Client, error)
}

type clientsService struct {
	client *api.Client
}

func newClientsService(client *api.Client) ClientsService {
	return &clientsService{client: client}
}

func (s *clientsService) List() ([]Client, error) {
	resp, err := s.client.Get("/clients")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Client `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *clientsService) Get(uuid string) (*Client, error) {
	resp, err := s.client.Get("/clients/" + uuid)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Client `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *clientsService) Create(name, email, company string) (*Client, error) {
	body := map[string]interface{}{
		"name": name,
	}
	if email != "" {
		body["email"] = email
	}
	if company != "" {
		body["company"] = company
	}

	resp, err := s.client.Post("/clients", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Client `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

