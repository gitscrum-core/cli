package wiki

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

func TestNewCmdWiki(t *testing.T) {
	f := factory.New()
	cmd := NewCmdWiki(f)

	if !strings.HasPrefix(cmd.Use, "wiki") {
		t.Errorf("Use should start with 'wiki', got %q", cmd.Use)
	}

	if len(cmd.Commands()) == 0 {
		t.Error("expected subcommands, got none")
	}
}

// TestWikiList_Success tests listing wiki pages
func TestWikiList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/wiki/pages" && r.Method == http.MethodGet {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":       "page-1",
						"slug":       "getting-started",
						"title":      "Getting Started",
						"updated_at": "2026-02-07T10:00:00Z",
						"author":     map[string]interface{}{"name": "Alice"},
						"children":   []map[string]interface{}{},
					},
					{
						"uuid":       "page-2",
						"slug":       "api-reference",
						"title":      "API Reference",
						"updated_at": "2026-02-06T15:00:00Z",
						"author":     map[string]interface{}{"name": "Bob"},
						"children": []map[string]interface{}{
							{"uuid": "page-3", "slug": "authentication", "title": "Authentication"},
						},
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
	resp, err := client.Get("/wiki/pages")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Page `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(result.Data))
	}

	if result.Data[0].Title != "Getting Started" {
		t.Errorf("first page title = %q, want 'Getting Started'", result.Data[0].Title)
	}

	// Check nested children
	if len(result.Data[1].Children) != 1 {
		t.Errorf("API Reference children = %d, want 1", len(result.Data[1].Children))
	}
}

// TestWikiView_Success tests viewing a wiki page
func TestWikiView_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/wiki/pages/") && r.Method == http.MethodGet {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":       "page-1",
					"slug":       "getting-started",
					"title":      "Getting Started",
					"page":       "# Getting Started\n\nWelcome to the documentation.",
					"updated_at": "2026-02-07T10:00:00Z",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Get("/wiki/pages/getting-started")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data Page `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if result.Data.Page == "" {
		t.Error("expected page content, got empty")
	}

	if result.Data.Title != "Getting Started" {
		t.Errorf("title = %q, want 'Getting Started'", result.Data.Title)
	}
}

// TestWikiCreate_Success tests creating a wiki page
func TestWikiCreate_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/wiki/pages" {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":  "page-new",
					"slug":  "new-page",
					"title": "New Page",
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
	resp, err := client.Post("/wiki/pages", map[string]interface{}{
		"title":   "New Page",
		"content": "# New Page\n\nContent here.",
	})
	if err != nil {
		t.Fatalf("Post failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("status = %d, want 201", resp.StatusCode)
	}
}

// TestWikiSearch_Success tests searching wiki pages
func TestWikiSearch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/search") && r.URL.Query().Get("q") != "" {
			response := map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"uuid":  "page-1",
						"slug":  "getting-started",
						"title": "Getting Started",
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
	resp, err := client.Get("/wiki/pages/search?q=started")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data []Page `json:"data"`
	}
	api.DecodeResponse(resp, &result)

	if len(result.Data) == 0 {
		t.Fatal("expected search results")
	}
}

// TestWikiEdit_Success tests editing a wiki page
func TestWikiEdit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch && strings.Contains(r.URL.Path, "/wiki/pages/") {
			response := map[string]interface{}{
				"data": map[string]interface{}{
					"uuid":  "page-1",
					"slug":  "getting-started",
					"title": "Getting Started",
				},
			}
			json.NewEncoder(w).Encode(response)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := api.NewClient(server.URL, &auth.Token{AccessToken: "test-token"})
	resp, err := client.Patch("/wiki/pages/getting-started", map[string]interface{}{
		"content": "# Updated Content",
	})
	if err != nil {
		t.Fatalf("Patch failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}
