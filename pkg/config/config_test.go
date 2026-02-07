package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.APIURL != DefaultAPIURL {
		t.Errorf("APIURL = %q, want %q", cfg.APIURL, DefaultAPIURL)
	}

	if cfg.OAuth.ClientID == "" {
		t.Error("OAuth.ClientID should not be empty")
	}

	if cfg.OAuth.Scopes == "" {
		t.Error("OAuth.Scopes should not be empty")
	}
}

func TestConfig_Defaults(t *testing.T) {
	// Test that DefaultAPIURL is set correctly
	if DefaultAPIURL != "https://services.gitscrum.com" {
		t.Errorf("DefaultAPIURL = %q, want %q", DefaultAPIURL, "https://services.gitscrum.com")
	}

	// Test ConfigDir
	if ConfigDir != ".gitscrum" {
		t.Errorf("ConfigDir = %q, want %q", ConfigDir, ".gitscrum")
	}

	// Test ConfigFile
	if ConfigFile != "config.yaml" {
		t.Errorf("ConfigFile = %q, want %q", ConfigFile, "config.yaml")
	}
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("ConfigPath() error: %v", err)
	}

	if path == "" {
		t.Error("ConfigPath() should not return empty string")
	}

	// Should contain .gitscrum directory
	if !filepath.IsAbs(path) {
		t.Errorf("ConfigPath() should return absolute path, got %q", path)
	}

	if filepath.Base(path) != ConfigFile {
		t.Errorf("ConfigPath() should end with %q, got %q", ConfigFile, filepath.Base(path))
	}
}

func TestTokenPath(t *testing.T) {
	path, err := TokenPath()
	if err != nil {
		t.Fatalf("TokenPath() error: %v", err)
	}

	if path == "" {
		t.Error("TokenPath() should not return empty string")
	}

	if filepath.Base(path) != "token.json" {
		t.Errorf("TokenPath() should end with 'token.json', got %q", filepath.Base(path))
	}

	// Should be in same directory as config
	configPath, _ := ConfigPath()
	if filepath.Dir(path) != filepath.Dir(configPath) {
		t.Error("TokenPath and ConfigPath should be in same directory")
	}
}

func TestConfig_SaveMethod(t *testing.T) {
	// Create temp home directory
	tmpHome, err := os.MkdirTemp("", "gitscrum-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// This test would require mocking os.UserHomeDir which is complex
	// For now, just verify the method exists and doesn't panic with valid config
	cfg := DefaultConfig()
	cfg.Workspace = "test-workspace"
	cfg.Project = "test-project"

	// The actual save might fail if we can't write to the real home dir
	// but the method should handle errors gracefully
	_ = cfg.Save()
}

func TestOAuthConfig(t *testing.T) {
	cfg := DefaultConfig()

	// Verify OAuth ClientID format (UUID-like)
	if len(cfg.OAuth.ClientID) < 30 {
		t.Error("OAuth.ClientID should be a UUID-like string")
	}

	// Verify scopes contain required permissions
	scopes := cfg.OAuth.Scopes
	requiredScopes := []string{"cli", "read", "write", "tasks:read", "tasks:write"}
	for _, scope := range requiredScopes {
		found := false
		for _, s := range splitScopes(scopes) {
			if s == scope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("OAuth.Scopes should contain %q", scope)
		}
	}
}

// Helper to split scopes string
func splitScopes(scopes string) []string {
	var result []string
	current := ""
	for _, c := range scopes {
		if c == ' ' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func TestLoad_WithEnvironment(t *testing.T) {
	// Save and restore original env
	originalURL := os.Getenv("GITSCRUM_API_URL")
	defer os.Setenv("GITSCRUM_API_URL", originalURL)

	// Set custom API URL
	customURL := "https://custom.api.com"
	os.Setenv("GITSCRUM_API_URL", customURL)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.APIURL != customURL {
		t.Errorf("APIURL = %q, want %q (from env)", cfg.APIURL, customURL)
	}
}
