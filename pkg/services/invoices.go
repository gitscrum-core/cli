package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// Invoice represents an invoice
type Invoice struct {
	UUID      string    `json:"uuid"`
	Number    string    `json:"number"`
	Client    Client    `json:"client"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"` // draft, sent, paid, overdue
	DueDate   time.Time `json:"due_date"`
	PaidAt    time.Time `json:"paid_at"`
	CreatedAt time.Time `json:"created_at"`
}

// InvoicesService handles invoice operations
type InvoicesService interface {
	// List returns all invoices
	List(status string) ([]Invoice, error)

	// Get returns an invoice by ID
	Get(uuid string) (*Invoice, error)

	// MarkAsPaid marks invoice as paid
	MarkAsPaid(uuid string) (*Invoice, error)
}

type invoicesService struct {
	client *api.Client
}

func newInvoicesService(client *api.Client) InvoicesService {
	return &invoicesService{client: client}
}

func (s *invoicesService) List(status string) ([]Invoice, error) {
	path := "/invoices"
	if status != "" {
		path += "?status=" + status
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Invoice `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *invoicesService) Get(uuid string) (*Invoice, error) {
	resp, err := s.client.Get("/invoices/" + uuid)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Invoice `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *invoicesService) MarkAsPaid(uuid string) (*Invoice, error) {
	resp, err := s.client.Post("/invoices/"+uuid+"/mark-paid", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Invoice `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
