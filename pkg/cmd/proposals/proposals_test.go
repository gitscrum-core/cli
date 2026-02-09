package proposals

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

func TestNewCmdProposals(t *testing.T) {
	f := factory.New()
	cmd := NewCmdProposals(f)

	if !strings.HasPrefix(cmd.Use, "proposals") {
		t.Errorf("Use should start with 'proposals', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestProposalsList_Success tests listing proposals
func TestProposalsList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/proposals" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":                  "prop-1",
						"ref_code":              "PRP-2026-001",
						"code":                  "PRP-001",
						"title":                 "Website Redesign",
						"status":                "sent",
						"status_label":          "Sent",
						"total_value_formatted": "50,000.00",
						"currency":              "USD",
						"client_name":           "TechCorp",
					},
					{
						"uuid":                  "prop-2",
						"ref_code":              "PRP-2026-002",
						"code":                  "PRP-002",
						"title":                 "Mobile App",
						"status":                "approved",
						"status_label":          "Approved",
						"total_value_formatted": "80,000.00",
						"currency":              "USD",
						"client_name":           "StartupXYZ",
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
	resp, err := client.Get("/proposals")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Proposal `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 proposals, got %d", len(result.Data))
	}

	if result.Data[0].Title != "Website Redesign" {
		t.Errorf("first proposal title = %q, want 'Website Redesign'", result.Data[0].Title)
	}
}

// TestProposalView_Success tests viewing a single proposal
func TestProposalView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/proposals/") && !strings.Contains(r.URL.Path, "/send") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":                  "prop-1",
					"ref_code":              "PRP-2026-001",
					"code":                  "PRP-001",
					"title":                 "Website Redesign",
					"status":                "sent",
					"status_label":          "Sent",
					"total_value_formatted": "50,000.00",
					"currency":              "USD",
					"client_name":           "TechCorp",
					"expires_at":            map[string]interface{}{"date": "2026-02-28"},
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/proposals/prop-1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data Proposal `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.TotalValueFormatted != "50,000.00" {
		t.Errorf("total_value_formatted = %q, want '50,000.00'", result.Data.TotalValueFormatted)
	}
}

// TestProposalCreate_Success tests creating a proposal
func TestProposalCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/proposals" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":   "prop-new",
					"code":   "PRP-2026-003",
					"title":  "New Project",
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
	resp, err := client.Post("/proposals", map[string]interface{}{
		"title":       "New Project",
		"client_uuid": "client-1",
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

// TestProposalSend_Success tests sending a proposal
func TestProposalSend_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/send") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":   "prop-1",
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
	resp, err := client.Post("/proposals/prop-1/send", nil)
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

// TestProposalConvert_Success tests converting proposal to project
func TestProposalConvert_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/convert-to-project") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid": "proj-new",
					"slug": "website-redesign",
					"name": "Website Redesign",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Post("/proposals/prop-1/convert-to-project", nil)
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Slug string `json:"slug"`
		} `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Slug != "website-redesign" {
		t.Errorf("slug = %q, want 'website-redesign'", result.Data.Slug)
	}
}
