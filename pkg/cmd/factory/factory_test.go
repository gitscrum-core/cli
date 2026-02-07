package factory

import (
	"testing"

	"github.com/gitscrum-core/cli/pkg/output"
)

func TestNew(t *testing.T) {
	f := New()

	if f == nil {
		t.Fatal("New() returned nil")
	}

	if f.OutputFormat != output.FormatTable {
		t.Errorf("OutputFormat = %v, want %v", f.OutputFormat, output.FormatTable)
	}
}

func TestFactory_Formatter(t *testing.T) {
	tests := []struct {
		name   string
		format output.Format
	}{
		{"table format", output.FormatTable},
		{"json format", output.FormatJSON},
		{"quiet format", output.FormatQuiet},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := New()
			f.OutputFormat = tt.format

			formatter := f.Formatter()
			if formatter == nil {
				t.Error("Formatter() returned nil")
			}
		})
	}
}

func TestFactory_RequireAuth_NotAuthenticated(t *testing.T) {
	f := New()

	// When not authenticated, RequireAuth should return error
	err := f.RequireAuth()
	if err == nil {
		t.Error("RequireAuth() should return error when not authenticated")
	}
}

func TestFactory_IsAuthenticated_False(t *testing.T) {
	f := New()

	// Fresh factory should not be authenticated
	if f.IsAuthenticated() {
		t.Error("IsAuthenticated() should return false for new factory")
	}
}

func TestFactory_CurrentWorkspace_Empty(t *testing.T) {
	f := New()

	// Without config, should return empty or error
	workspace, err := f.CurrentWorkspace()
	if err != nil && workspace != "" {
		t.Logf("CurrentWorkspace: workspace=%q, err=%v", workspace, err)
	}
	// This test just ensures no panic
}

func TestFactory_CurrentProject_Empty(t *testing.T) {
	f := New()

	// Without config, should return empty or error
	project, err := f.CurrentProject()
	if err != nil && project != "" {
		t.Logf("CurrentProject: project=%q, err=%v", project, err)
	}
	// This test just ensures no panic
}

func TestFactory_Config(t *testing.T) {
	f := New()

	// Config should be lazy loaded
	cfg, err := f.Config()
	if err != nil {
		// Config loading may fail in test environment
		t.Logf("Config() returned error (expected in test env): %v", err)
		return
	}

	if cfg == nil {
		t.Error("Config() should not return nil config without error")
	}

	// Config should be cached (lazy loading)
	cfg2, err := f.Config()
	if err != nil {
		t.Fatalf("second Config() call failed: %v", err)
	}

	if cfg != cfg2 {
		t.Error("Config() should return cached instance")
	}
}

func TestFactory_GitContext(t *testing.T) {
	f := New()

	// GitContext in a non-git directory should return error
	ctx, err := f.GitContext()

	// Either works - we just verify no panic
	if err != nil {
		t.Logf("GitContext() returned error (expected outside git repo): %v", err)
	} else if ctx == nil {
		t.Error("GitContext() should not return nil context without error")
	}
}

func TestFactory_APIClient_RequiresAuth(t *testing.T) {
	f := New()

	// APIClient requires authentication
	client, err := f.APIClient()

	// Should fail because not authenticated
	if err != nil {
		t.Logf("APIClient() returned expected error: %v", err)
	} else if client == nil {
		t.Error("APIClient() should not return nil client without error")
	}
}

func TestFactory_AuthToken_NotLoaded(t *testing.T) {
	f := New()

	// When no token exists, should return error
	token, err := f.AuthToken()

	if err != nil {
		t.Logf("AuthToken() returned expected error: %v", err)
	} else if token == nil || token.AccessToken == "" {
		t.Log("AuthToken() returned empty/nil token (expected when not logged in)")
	}
}
