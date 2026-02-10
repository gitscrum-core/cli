package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// Proposal represents a proposal/quote
type Proposal struct {
	UUID      string    `json:"uuid"`
	Title     string    `json:"title"`
	Client    Client    `json:"client"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // draft, sent, accepted, rejected
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// ProposalsService handles proposal operations
type ProposalsService interface {
	// List returns all proposals
	List(status string) ([]Proposal, error)

	// Get returns a proposal by ID
	Get(uuid string) (*Proposal, error)

	// Send sends a proposal to the client
	Send(uuid string) (*Proposal, error)
}

type proposalsService struct {
	client *api.Client
}

func newProposalsService(client *api.Client) ProposalsService {
	return &proposalsService{client: client}
}

func (s *proposalsService) List(status string) ([]Proposal, error) {
	path := "/proposals"
	if status != "" {
		path += "?status=" + status
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Proposal `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *proposalsService) Get(uuid string) (*Proposal, error) {
	resp, err := s.client.Get("/proposals/" + uuid)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Proposal `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *proposalsService) Send(uuid string) (*Proposal, error) {
	resp, err := s.client.Post("/proposals/"+uuid+"/send", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Proposal `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
