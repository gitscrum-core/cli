package invoices

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

func TestNewCmdInvoices(t *testing.T) {
	f := factory.New()
	cmd := NewCmdInvoices(f)

	if !strings.HasPrefix(cmd.Use, "invoices") {
		t.Errorf("Use should start with 'invoices', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestInvoicesList_Success tests listing invoices
func TestInvoicesList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/company-invoices" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":     "inv-1",
						"code":     "INV-2026-001",
						"amount":   12500.00,
						"currency": "USD",
						"status":   "paid",
						"due_date": "2026-01-15",
						"client":   map[string]interface{}{"name": "TechCorp"},
					},
					{
						"uuid":     "inv-2",
						"code":     "INV-2026-002",
						"amount":   8500.00,
						"currency": "USD",
						"status":   "overdue",
						"due_date": "2026-02-01",
						"client":   map[string]interface{}{"name": "StartupXYZ"},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/company-invoices")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Invoice `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 invoices, got %d", len(result.Data))
	}

	if result.Data[0].Code != "INV-2026-001" {
		t.Errorf("first invoice code = %q, want 'INV-2026-001'", result.Data[0].Code)
	}

	if result.Data[1].Status != "overdue" {
		t.Errorf("second invoice status = %q, want 'overdue'", result.Data[1].Status)
	}
}

// TestInvoiceView_Success tests viewing a single invoice
func TestInvoiceView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/company-invoices/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":     "inv-1",
					"code":     "INV-2026-001",
					"amount":   12500.00,
					"currency": "USD",
					"status":   "paid",
					"due_date": "2026-01-15",
					"client":   map[string]interface{}{"name": "TechCorp"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/company-invoices/inv-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data Invoice `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Amount != 12500.00 {
		t.Errorf("amount = %.2f, want 12500.00", result.Data.Amount)
	}
}

// TestInvoiceCreate_Success tests creating an invoice
func TestInvoiceCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/company-invoices" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":   "inv-new",
					"code":   "INV-2026-003",
					"amount": 15000.00,
					"status": "draft",
				},
			}
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/company-invoices", map[string]interface{}{
		"client_uuid": "client-1",
		"amount":      15000.00,
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

// TestInvoiceSend_Success tests sending an invoice
func TestInvoiceSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/send") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":   "inv-1",
					"status": "sent",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/company-invoices/inv-1/send", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Status != "sent" {
		t.Errorf("status = %q, want 'sent'", result.Data.Status)
	}
}

// TestInvoiceMarkPaid_Success tests marking invoice as paid
func TestInvoiceMarkPaid_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/paid") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":   "inv-1",
					"status": "paid",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/company-invoices/inv-1/paid", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Status string `json:"status"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Status != "paid" {
		t.Errorf("status = %q, want 'paid'", result.Data.Status)
	}
}
