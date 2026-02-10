package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// CRMContact represents a CRM contact
type CRMContact struct {
	UUID    string `json:"uuid"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	Phone   string `json:"phone"`
	Status  string `json:"status"`
	Source  string `json:"source"`
	Notes   string `json:"notes"`
}

// CRMDeal represents a CRM deal
type CRMDeal struct {
	UUID        string     `json:"uuid"`
	Title       string     `json:"title"`
	Value       float64    `json:"value"`
	Stage       string     `json:"stage"`
	Contact     CRMContact `json:"contact"`
	Probability int        `json:"probability"`
}

// CRMService handles CRM operations
type CRMService interface {
	// ListContacts returns all contacts
	ListContacts() ([]CRMContact, error)

	// ListDeals returns all deals
	ListDeals() ([]CRMDeal, error)
}

type crmService struct {
	client *api.Client
}

func newCRMService(client *api.Client) CRMService {
	return &crmService{client: client}
}

func (s *crmService) ListContacts() ([]CRMContact, error) {
	resp, err := s.client.Get("/crm/contacts")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []CRMContact `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *crmService) ListDeals() ([]CRMDeal, error) {
	resp, err := s.client.Get("/crm/deals")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []CRMDeal `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
