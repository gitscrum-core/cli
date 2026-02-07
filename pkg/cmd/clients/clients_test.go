package clients

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

func TestNewCmdClients(t *testing.T) {
	f := factory.New()
	cmd := NewCmdClients(f)

	if !strings.HasPrefix(cmd.Use, "clients") {
		t.Errorf("Use should start with 'clients', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestClientsList_Success tests listing clients
func TestClientsList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/contact-companies/clients" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":          "client-1",
						"slug":          "techcorp",
						"name":          "TechCorp",
						"email":         "contact@techcorp.com",
						"company_name":  "TechCorp Inc",
						"projects":      3,
						"total_revenue": 150000.00,
					},
					{
						"uuid":          "client-2",
						"slug":          "startupxyz",
						"name":          "StartupXYZ",
						"email":         "hello@startupxyz.com",
						"company_name":  "StartupXYZ Ltd",
						"projects":      2,
						"total_revenue": 80000.00,
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
	resp, err := client.Get("/contact-companies/clients")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Client `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 clients, got %d", len(result.Data))
	}

	if result.Data[0].Name != "TechCorp" {
		t.Errorf("first client name = %q, want 'TechCorp'", result.Data[0].Name)
	}
}

// TestClientView_Success tests viewing a single client
func TestClientView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/contact-companies/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":          "client-1",
					"slug":          "techcorp",
					"name":          "TechCorp",
					"email":         "contact@techcorp.com",
					"phone":         "+1 555-123-4567",
					"company_name":  "TechCorp Inc",
					"projects":      3,
					"total_revenue": 150000.00,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/contact-companies/techcorp")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data Client `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Email != "contact@techcorp.com" {
		t.Errorf("email = %q, want 'contact@techcorp.com'", result.Data.Email)
	}
}

// TestClientCreate_Success tests creating a client
func TestClientCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/contact-companies" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "client-new",
					"slug": "new-client",
					"name": "New Client",
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
	resp, err := client.Post("/contact-companies", map[string]interface{}{
		"name":  "New Client",
		"email": "contact@newclient.com",
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

// TestClientStats_Success tests getting client statistics
func TestClientStats_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/contact-companies/stats" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"total_clients":  24,
					"active_clients": 18,
					"at_risk":        4,
					"churned":        2,
					"total_revenue":  450000.00,
					"mrr":            35000.00,
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/contact-companies/stats")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalClients  int     `json:"total_clients"`
			ActiveClients int     `json:"active_clients"`
			TotalRevenue  float64 `json:"total_revenue"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.TotalClients != 24 {
		t.Errorf("total clients = %d, want 24", result.Data.TotalClients)
	}

	if result.Data.TotalRevenue != 450000.00 {
		t.Errorf("total revenue = %.2f, want 450000.00", result.Data.TotalRevenue)
	}
}

// TestClientProjects_Success tests listing client projects
func TestClientProjects_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.RawQuery, "contact_company_uuid") {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":   "proj-1",
						"slug":   "website-redesign",
						"name":   "Website Redesign",
						"status": "active",
					},
					{
						"uuid":   "proj-2",
						"slug":   "mobile-app",
						"name":   "Mobile App",
						"status": "active",
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
	resp, err := client.Get("/projects?contact_company_uuid=client-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Slug string `json:"slug"`
			Name string `json:"name"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(result.Data))
	}
}
