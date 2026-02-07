package projectconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindConfigFile(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "sub", "deep")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Test when no config exists
	path, err := FindConfigFile(subDir)
	if err != nil {
		t.Errorf("FindConfigFile error: %v", err)
	}
	if path != "" {
		t.Errorf("expected empty path, got %s", path)
	}

	// Create config in parent directory
	configPath := filepath.Join(tmpDir, ConfigFileName)
	if err := os.WriteFile(configPath, []byte("version: \"1\"\nworkspace: test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Should find config from subdirectory
	foundPath, err := FindConfigFile(subDir)
	if err != nil {
		t.Errorf("FindConfigFile error: %v", err)
	}
	if foundPath != configPath {
		t.Errorf("expected %s, got %s", configPath, foundPath)
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFileName)

	content := `
version: "1"
workspace: my-workspace
project: my-project

branch:
  default_prefix: feature
  include_title: true
  max_length: 60

timer:
  auto_start: true
  round_to: 15

hooks:
  prepend_task_code: true
  commit_format: "%s: %s"

automation:
  on_pr_merge: done
  complete_on_merge: true
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFromFile(configPath)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}

	// Verify values
	if cfg.Workspace != "my-workspace" {
		t.Errorf("expected workspace 'my-workspace', got '%s'", cfg.Workspace)
	}
	if cfg.Project != "my-project" {
		t.Errorf("expected project 'my-project', got '%s'", cfg.Project)
	}
	if cfg.Branch.DefaultPrefix != "feature" {
		t.Errorf("expected branch prefix 'feature', got '%s'", cfg.Branch.DefaultPrefix)
	}
	if cfg.Branch.MaxLength != 60 {
		t.Errorf("expected max_length 60, got %d", cfg.Branch.MaxLength)
	}
	if !cfg.Timer.AutoStart {
		t.Error("expected auto_start true")
	}
	if cfg.Timer.RoundTo != 15 {
		t.Errorf("expected round_to 15, got %d", cfg.Timer.RoundTo)
	}
	if !cfg.Hooks.PrependTaskCode {
		t.Error("expected prepend_task_code true")
	}
	if cfg.Automation.OnPRMerge != "done" {
		t.Errorf("expected on_pr_merge 'done', got '%s'", cfg.Automation.OnPRMerge)
	}
	if !cfg.Automation.CompleteOnMerge {
		t.Error("expected complete_on_merge true")
	}
}

func TestSetDefaults(t *testing.T) {
	cfg := &ProjectConfig{}
	cfg.setDefaults()

	if cfg.Version != "1" {
		t.Errorf("expected version '1', got '%s'", cfg.Version)
	}
	if cfg.Branch.DefaultPrefix != "feature" {
		t.Errorf("expected default prefix 'feature', got '%s'", cfg.Branch.DefaultPrefix)
	}
	if cfg.Branch.MaxLength != 50 {
		t.Errorf("expected max_length 50, got %d", cfg.Branch.MaxLength)
	}
	if cfg.Timer.MinDuration != 1 {
		t.Errorf("expected min_duration 1, got %d", cfg.Timer.MinDuration)
	}
	if cfg.Hooks.CommitFormat != "[%s] %s" {
		t.Errorf("expected commit format '[%%s] %%s', got '%s'", cfg.Hooks.CommitFormat)
	}
}

func TestInit(t *testing.T) {
	// Change to temp directory
	tmpDir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	cfg, err := Init("test-workspace", "test-project")
	if err != nil {
		t.Fatalf("Init error: %v", err)
	}

	if cfg.Workspace != "test-workspace" {
		t.Errorf("expected workspace 'test-workspace', got '%s'", cfg.Workspace)
	}
	if cfg.Project != "test-project" {
		t.Errorf("expected project 'test-project', got '%s'", cfg.Project)
	}

	// Verify file was created
	if _, err := os.Stat(ConfigFileName); os.IsNotExist(err) {
		t.Error("expected .gitscrum.yml to be created")
	}

	// Verify it can be loaded
	loaded, err := LoadFromFile(ConfigFileName)
	if err != nil {
		t.Fatalf("LoadFromFile error: %v", err)
	}
	if loaded.Workspace != "test-workspace" {
		t.Errorf("loaded workspace mismatch")
	}
}

func TestMerge(t *testing.T) {
	cfg := &ProjectConfig{
		Workspace: "project-workspace",
		Project:   "", // Empty - should use global
	}

	workspace, project := cfg.Merge("global-workspace", "global-project")

	if workspace != "project-workspace" {
		t.Errorf("expected project workspace to take precedence, got '%s'", workspace)
	}
	if project != "global-project" {
		t.Errorf("expected global project when empty, got '%s'", project)
	}
}

func TestAltConfigFileName(t *testing.T) {
	tmpDir := t.TempDir()
	altConfigPath := filepath.Join(tmpDir, AltConfigFileName)

	// Create .gitscrum.yaml (alternative name)
	if err := os.WriteFile(altConfigPath, []byte("version: \"1\"\nworkspace: alt"), 0644); err != nil {
		t.Fatal(err)
	}

	foundPath, err := FindConfigFile(tmpDir)
	if err != nil {
		t.Errorf("FindConfigFile error: %v", err)
	}
	if foundPath != altConfigPath {
		t.Errorf("expected to find .gitscrum.yaml, got %s", foundPath)
	}
}
