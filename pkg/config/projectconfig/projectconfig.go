// Package projectconfig handles project-level configuration (.gitscrum.yml)
package projectconfig

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// ConfigFileName is the project config file name
	ConfigFileName = ".gitscrum.yml"

	// AltConfigFileName is the alternative config file name
	AltConfigFileName = ".gitscrum.yaml"
)

// ProjectConfig holds project-level configuration
// This is read from .gitscrum.yml in the project root
type ProjectConfig struct {
	// Version of the config schema
	Version string `yaml:"version,omitempty"`

	// Workspace slug for this project
	Workspace string `yaml:"workspace,omitempty"`

	// Project slug for this project
	Project string `yaml:"project,omitempty"`

	// Branch patterns for automatic task detection
	Branch BranchConfig `yaml:"branch,omitempty"`

	// Timer settings
	Timer TimerConfig `yaml:"timer,omitempty"`

	// Hooks configuration
	Hooks HooksConfig `yaml:"hooks,omitempty"`

	// Labels to apply automatically
	Labels []string `yaml:"labels,omitempty"`

	// Custom task code pattern (regex)
	TaskCodePattern string `yaml:"task_code_pattern,omitempty"`

	// Team configuration
	Team TeamConfig `yaml:"team,omitempty"`

	// Automation settings
	Automation AutomationConfig `yaml:"automation,omitempty"`
}

// BranchConfig configures branch naming conventions
type BranchConfig struct {
	// Prefix patterns for different task types
	// e.g., feature/ → feature, bugfix/ → bug
	Prefixes map[string]string `yaml:"prefixes,omitempty"`

	// Default prefix when creating branches
	DefaultPrefix string `yaml:"default_prefix,omitempty"`

	// Include task title in branch name
	IncludeTitle bool `yaml:"include_title,omitempty"`

	// Max branch name length
	MaxLength int `yaml:"max_length,omitempty"`
}

// TimerConfig configures time tracking behavior
type TimerConfig struct {
	// Auto-start timer when switching to task branch
	AutoStart bool `yaml:"auto_start,omitempty"`

	// Auto-stop timer when switching away from task branch
	AutoStop bool `yaml:"auto_stop,omitempty"`

	// Minimum duration to log (in minutes)
	MinDuration int `yaml:"min_duration,omitempty"`

	// Round to nearest (in minutes): 5, 15, 30
	RoundTo int `yaml:"round_to,omitempty"`

	// Remind to stop timer after X hours of inactivity
	RemindAfter int `yaml:"remind_after,omitempty"`
}

// HooksConfig configures git hooks behavior
type HooksConfig struct {
	// Prepend task code to commit messages
	PrependTaskCode bool `yaml:"prepend_task_code,omitempty"`

	// Commit message format: "[CODE] message" or "CODE: message"
	CommitFormat string `yaml:"commit_format,omitempty"`

	// Validate task exists before push
	ValidateOnPush bool `yaml:"validate_on_push,omitempty"`

	// Show task info on checkout
	ShowTaskOnCheckout bool `yaml:"show_task_on_checkout,omitempty"`
}

// TeamConfig configures team-specific settings
type TeamConfig struct {
	// Default assignee for new tasks
	DefaultAssignee string `yaml:"default_assignee,omitempty"`

	// Reviewers to request on PRs
	Reviewers []string `yaml:"reviewers,omitempty"`
}

// AutomationConfig configures automation behavior
type AutomationConfig struct {
	// Move task to column on PR open
	OnPROpen string `yaml:"on_pr_open,omitempty"`

	// Move task to column on PR merge
	OnPRMerge string `yaml:"on_pr_merge,omitempty"`

	// Move task to column on PR close (without merge)
	OnPRClose string `yaml:"on_pr_close,omitempty"`

	// Move task to column on deploy to staging
	OnDeployStaging string `yaml:"on_deploy_staging,omitempty"`

	// Move task to column on deploy to production
	OnDeployProd string `yaml:"on_deploy_prod,omitempty"`

	// Complete task on PR merge
	CompleteOnMerge bool `yaml:"complete_on_merge,omitempty"`
}

// Load reads the project configuration from the current directory or its parents
func Load() (*ProjectConfig, error) {
	return LoadFromPath(".")
}

// LoadFromPath reads the project configuration starting from a given path
func LoadFromPath(startPath string) (*ProjectConfig, error) {
	configPath, err := FindConfigFile(startPath)
	if err != nil {
		return nil, err
	}

	if configPath == "" {
		// No config file found, return empty config
		return &ProjectConfig{}, nil
	}

	return LoadFromFile(configPath)
}

// LoadFromFile reads the project configuration from a specific file
func LoadFromFile(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Set defaults
	cfg.setDefaults()

	return &cfg, nil
}

// FindConfigFile searches for .gitscrum.yml starting from startPath up to root
func FindConfigFile(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	for {
		// Check for .gitscrum.yml
		configPath := filepath.Join(absPath, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		// Check for .gitscrum.yaml
		altConfigPath := filepath.Join(absPath, AltConfigFileName)
		if _, err := os.Stat(altConfigPath); err == nil {
			return altConfigPath, nil
		}

		// Move to parent directory
		parent := filepath.Dir(absPath)
		if parent == absPath {
			// Reached root
			break
		}
		absPath = parent
	}

	return "", nil
}

// Exists checks if a project config file exists in the current directory tree
func Exists() bool {
	path, _ := FindConfigFile(".")
	return path != ""
}

// GetConfigPath returns the path where config was found, or empty string
func GetConfigPath() string {
	path, _ := FindConfigFile(".")
	return path
}

// setDefaults sets default values for the config
func (c *ProjectConfig) setDefaults() {
	if c.Branch.DefaultPrefix == "" {
		c.Branch.DefaultPrefix = "feature"
	}
	if c.Branch.MaxLength == 0 {
		c.Branch.MaxLength = 50
	}
	if c.Timer.MinDuration == 0 {
		c.Timer.MinDuration = 1
	}
	if c.Hooks.CommitFormat == "" {
		c.Hooks.CommitFormat = "[%s] %s" // [CODE] message
	}
	if c.Version == "" {
		c.Version = "1"
	}
}

// Save writes the configuration to a file
func (c *ProjectConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Init creates a new .gitscrum.yml file with sensible defaults
func Init(workspace, project string) (*ProjectConfig, error) {
	cfg := &ProjectConfig{
		Version:   "1",
		Workspace: workspace,
		Project:   project,
		Branch: BranchConfig{
			DefaultPrefix: "feature",
			IncludeTitle:  true,
			MaxLength:     50,
			Prefixes: map[string]string{
				"feature": "feature",
				"bug":     "bugfix",
				"hotfix":  "hotfix",
				"chore":   "chore",
			},
		},
		Timer: TimerConfig{
			AutoStart:   false,
			AutoStop:    false,
			MinDuration: 1,
			RoundTo:     5,
		},
		Hooks: HooksConfig{
			PrependTaskCode:    true,
			CommitFormat:       "[%s] %s",
			ValidateOnPush:     false,
			ShowTaskOnCheckout: true,
		},
		Automation: AutomationConfig{
			OnPROpen:        "in-progress",
			OnPRMerge:       "done",
			CompleteOnMerge: true,
		},
	}

	if err := cfg.Save(ConfigFileName); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Merge combines project config with global config, project takes precedence
func (c *ProjectConfig) Merge(globalWorkspace, globalProject string) (workspace, project string) {
	workspace = globalWorkspace
	project = globalProject

	if c.Workspace != "" {
		workspace = c.Workspace
	}
	if c.Project != "" {
		project = c.Project
	}

	return
}
