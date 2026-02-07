package crm

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

func TestNewCmdCRM(t *testing.T) {
	f := factory.New()
	cmd := NewCmdCRM(f)

	if !strings.HasPrefix(cmd.Use, "crm") {
		t.Errorf("Use should start with 'crm', got %q", cmd.Use)
	}

	// Verify command has subcommands
	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}

	// Check for expected subcommands
	subcommands := make(map[string]bool)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Use] = true
	}

	expected := []string{"revenue", "at-risk", "pipeline", "projects", "leaderboard"}
	for _, name := range expected {
		if !subcommands[name] {
			t.Errorf("expected subcommand %q not found", name)
		}
	}
}

// TestCRMDashboard_Success tests the main CRM dashboard
func TestCRMDashboard_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/overview" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"summary": map[string]interface{}{
						"total_clients":    15,
						"active_projects":  8,
						"total_projects":   12,
						"projects_at_risk": 2,
					},
					"invoices": map[string]interface{}{
						"total":          50,
						"paid":           30,
						"pending":        15,
						"overdue":        5,
						"paid_amount":    150000.00,
						"pending_amount": 75000.00,
						"overdue_amount": 25000.00,
					},
					"proposals": map[string]interface{}{
						"total":            20,
						"approved":         10,
						"pending_approval": 8,
						"expiring_soon":    2,
						"approved_value":   300000.00,
						"pending_value":    150000.00,
					},
					"alerts": map[string]interface{}{
						"overdue_invoices":   5,
						"expiring_proposals": 2,
						"projects_at_risk":   2,
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
	resp, err := client.Get("/client-flow/dashboard/overview")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Summary struct {
				TotalClients   int `json:"total_clients"`
				ActiveProjects int `json:"active_projects"`
			} `json:"summary"`
			Invoices struct {
				Paid       int     `json:"paid"`
				PaidAmount float64 `json:"paid_amount"`
			} `json:"invoices"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Summary.TotalClients != 15 {
		t.Errorf("total clients = %d, want 15", result.Data.Summary.TotalClients)
	}

	if result.Data.Invoices.PaidAmount != 150000.00 {
		t.Errorf("paid amount = %.2f, want 150000.00", result.Data.Invoices.PaidAmount)
	}
}

// TestCRMRevenue_Success tests the revenue pipeline
func TestCRMRevenue_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/revenue-pipeline" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"invoices_summary": map[string]interface{}{
						"paid":    map[string]interface{}{"count": 30, "total": 150000.00},
						"pending": map[string]interface{}{"count": 15, "total": 75000.00},
						"overdue": map[string]interface{}{"count": 5, "total": 25000.00},
					},
					"proposals_summary": map[string]interface{}{
						"approved": map[string]interface{}{"count": 10, "total": 300000.00},
						"pending":  map[string]interface{}{"count": 8, "total": 150000.00},
					},
					"overdue_invoices": []map[string]interface{}{
						{
							"uuid":            "inv-1",
							"series":          "INV-2026-001",
							"client":          map[string]interface{}{"name": "TechCorp"},
							"amount":          5000.00,
							"currency_symbol": "$",
							"days_overdue":    15,
						},
					},
					"monthly_revenue": []map[string]interface{}{
						{"month": "2026-01", "total": 75000.00, "count": 15},
						{"month": "2026-02", "total": 50000.00, "count": 10},
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
	resp, err := client.Get("/client-flow/dashboard/revenue-pipeline")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			OverdueInvoices []struct {
				Series      string  `json:"series"`
				DaysOverdue int     `json:"days_overdue"`
				Amount      float64 `json:"amount"`
			} `json:"overdue_invoices"`
			MonthlyRevenue []struct {
				Month string  `json:"month"`
				Total float64 `json:"total"`
			} `json:"monthly_revenue"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data.OverdueInvoices) != 1 {
		t.Fatalf("expected 1 overdue invoice, got %d", len(result.Data.OverdueInvoices))
	}

	if result.Data.OverdueInvoices[0].DaysOverdue != 15 {
		t.Errorf("days overdue = %d, want 15", result.Data.OverdueInvoices[0].DaysOverdue)
	}

	if len(result.Data.MonthlyRevenue) != 2 {
		t.Errorf("monthly revenue entries = %d, want 2", len(result.Data.MonthlyRevenue))
	}
}

// TestCRMAtRisk_Success tests clients at risk
func TestCRMAtRisk_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/clients-at-risk" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"clients_at_risk": []map[string]interface{}{
						{
							"uuid":  "client-1",
							"name":  "TechCorp",
							"email": "contact@techcorp.com",
							"risks": []map[string]interface{}{
								{"type": "overdue_invoice", "label": "3 overdue invoices"},
								{"type": "stalled_project", "label": "Project stalled for 2 weeks"},
							},
						},
						{
							"uuid":  "client-2",
							"name":  "StartupXYZ",
							"email": "hello@startupxyz.com",
							"risks": []map[string]interface{}{
								{"type": "expiring_proposal", "label": "Proposal expires in 3 days"},
							},
						},
					},
					"summary": map[string]interface{}{
						"with_overdue_invoices":   1,
						"with_stalled_projects":   1,
						"with_expiring_proposals": 1,
						"total_at_risk":           2,
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
	resp, err := client.Get("/client-flow/dashboard/clients-at-risk")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			ClientsAtRisk []struct {
				Name  string `json:"name"`
				Risks []struct {
					Type  string `json:"type"`
					Label string `json:"label"`
				} `json:"risks"`
			} `json:"clients_at_risk"`
			Summary struct {
				TotalAtRisk int `json:"total_at_risk"`
			} `json:"summary"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Summary.TotalAtRisk != 2 {
		t.Errorf("total at risk = %d, want 2", result.Data.Summary.TotalAtRisk)
	}

	if len(result.Data.ClientsAtRisk[0].Risks) != 2 {
		t.Errorf("TechCorp risks = %d, want 2", len(result.Data.ClientsAtRisk[0].Risks))
	}
}

// TestCRMPipeline_Success tests pending approvals
func TestCRMPipeline_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/pending-approvals" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"all": []map[string]interface{}{
						{
							"type":            "proposal",
							"uuid":            "prop-1",
							"code":            "PRP-2026-001",
							"title":           "Website Redesign",
							"client":          map[string]interface{}{"name": "TechCorp"},
							"amount":          50000.00,
							"currency_symbol": "$",
							"days_waiting":    5,
						},
						{
							"type":            "invoice",
							"uuid":            "inv-1",
							"code":            "INV-2026-050",
							"title":           "",
							"client":          map[string]interface{}{"name": "StartupXYZ"},
							"amount":          12000.00,
							"currency_symbol": "$",
							"days_waiting":    3,
						},
					},
					"summary": map[string]interface{}{
						"total":           5,
						"proposals":       3,
						"change_requests": 1,
						"invoices":        1,
						"proposals_value": 120000.00,
						"invoices_value":  12000.00,
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
	resp, err := client.Get("/client-flow/dashboard/pending-approvals")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			All []struct {
				Type        string  `json:"type"`
				Code        string  `json:"code"`
				Amount      float64 `json:"amount"`
				DaysWaiting int     `json:"days_waiting"`
			} `json:"all"`
			Summary struct {
				Total          int     `json:"total"`
				ProposalsValue float64 `json:"proposals_value"`
			} `json:"summary"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Summary.Total != 5 {
		t.Errorf("total pending = %d, want 5", result.Data.Summary.Total)
	}

	if len(result.Data.All) != 2 {
		t.Fatalf("pending items = %d, want 2", len(result.Data.All))
	}

	if result.Data.All[0].Type != "proposal" {
		t.Errorf("first item type = %q, want 'proposal'", result.Data.All[0].Type)
	}
}

// TestCRMProjects_Success tests projects health
func TestCRMProjects_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/projects-health" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"projects": []map[string]interface{}{
						{
							"uuid":                    "proj-1",
							"slug":                    "website-redesign",
							"name":                    "Website Redesign",
							"client_name":             "TechCorp",
							"budget_hours":            100.0,
							"hours_used":              85.0,
							"budget_usage_percentage": 85.0,
							"is_over_budget":          false,
							"progress_percentage":     70.0,
							"health_status":           "warning",
						},
						{
							"uuid":                    "proj-2",
							"slug":                    "mobile-app",
							"name":                    "Mobile App",
							"client_name":             "StartupXYZ",
							"budget_hours":            200.0,
							"hours_used":              250.0,
							"budget_usage_percentage": 125.0,
							"is_over_budget":          true,
							"progress_percentage":     90.0,
							"health_status":           "critical",
						},
					},
					"summary": map[string]interface{}{
						"total":       10,
						"healthy":     7,
						"warning":     2,
						"critical":    1,
						"over_budget": 1,
						"at_risk":     2,
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
	resp, err := client.Get("/client-flow/dashboard/projects-health")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Projects []struct {
				Name         string  `json:"name"`
				HealthStatus string  `json:"health_status"`
				IsOverBudget bool    `json:"is_over_budget"`
				HoursUsed    float64 `json:"hours_used"`
			} `json:"projects"`
			Summary struct {
				Total      int `json:"total"`
				Healthy    int `json:"healthy"`
				Critical   int `json:"critical"`
				OverBudget int `json:"over_budget"`
			} `json:"summary"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Summary.Total != 10 {
		t.Errorf("total projects = %d, want 10", result.Data.Summary.Total)
	}

	if result.Data.Summary.OverBudget != 1 {
		t.Errorf("over budget = %d, want 1", result.Data.Summary.OverBudget)
	}

	// Check critical project
	criticalFound := false
	for _, p := range result.Data.Projects {
		if p.HealthStatus == "critical" {
			criticalFound = true
			if !p.IsOverBudget {
				t.Error("critical project should be over budget")
			}
		}
	}
	if !criticalFound {
		t.Error("expected to find a critical project")
	}
}

// TestCRMLeaderboard_Success tests client leaderboard
func TestCRMLeaderboard_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/client-flow/dashboard/leaderboard" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"clients": []map[string]interface{}{
						{
							"uuid":              "client-1",
							"name":              "TechCorp",
							"total_revenue":     250000.00,
							"pending_revenue":   50000.00,
							"total_projects":    5,
							"active_projects":   3,
							"reliability_score": 0.95,
							"avg_payment_days":  12.5,
						},
						{
							"uuid":              "client-2",
							"name":              "StartupXYZ",
							"total_revenue":     120000.00,
							"pending_revenue":   30000.00,
							"total_projects":    3,
							"active_projects":   2,
							"reliability_score": 0.88,
							"avg_payment_days":  18.0,
						},
					},
					"sort_by": "total_revenue",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/client-flow/dashboard/leaderboard")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Clients []struct {
				Name             string  `json:"name"`
				TotalRevenue     float64 `json:"total_revenue"`
				ReliabilityScore float64 `json:"reliability_score"`
				AvgPaymentDays   float64 `json:"avg_payment_days"`
			} `json:"clients"`
			SortBy string `json:"sort_by"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data.Clients) != 2 {
		t.Fatalf("clients = %d, want 2", len(result.Data.Clients))
	}

	// First client should have highest revenue
	if result.Data.Clients[0].TotalRevenue != 250000.00 {
		t.Errorf("top client revenue = %.2f, want 250000.00", result.Data.Clients[0].TotalRevenue)
	}

	if result.Data.SortBy != "total_revenue" {
		t.Errorf("sort_by = %q, want 'total_revenue'", result.Data.SortBy)
	}

	// Check reliability score
	if result.Data.Clients[0].ReliabilityScore != 0.95 {
		t.Errorf("reliability score = %.2f, want 0.95", result.Data.Clients[0].ReliabilityScore)
	}
}
