// Package config handles configuration management
package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// DefaultAPIURL is the default GitScrum API URL
	DefaultAPIURL = "https://services.gitscrum.com"
	
	// ConfigDir is the config directory name
	ConfigDir = ".gitscrum"
	
	// ConfigFile is the config file name
	ConfigFile = "config.yaml"
)

// Config holds the CLI configuration
type Config struct {
	// API URL
	APIURL string `mapstructure:"api_url"`
	
	// Default workspace slug
	Workspace string `mapstructure:"workspace"`
	
	// Default project slug
	Project string `mapstructure:"project"`
	
	// OAuth settings
	OAuth OAuthConfig `mapstructure:"oauth"`
}

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	ClientID string `mapstructure:"client_id"`
	Scopes   string `mapstructure:"scopes"`
}

// Load reads the configuration from file and environment
func Load() (*Config, error) {
	cfg := &Config{
		APIURL: DefaultAPIURL,
		OAuth: OAuthConfig{
			// CLI client ID from OAuthClientSeeder
			ClientID: "3e2d1c0b-a9b8-c7d6-e5f4-a3b2c1d0e9f8",
			// CLI scope + standard read/write permissions
			Scopes: "cli read write tasks:read tasks:write time-tracking:read time-tracking:write projects:read sprints:read standup:read standup:write",
		},
	}

	// Set defaults
	viper.SetDefault("api_url", DefaultAPIURL)
	viper.SetDefault("oauth.client_id", "3e2d1c0b-a9b8-c7d6-e5f4-a3b2c1d0e9f8")
	viper.SetDefault("oauth.scopes", "cli read write tasks:read tasks:write time-tracking:read time-tracking:write projects:read sprints:read standup:read standup:write")

	// Unmarshal config
	if err := viper.Unmarshal(cfg); err != nil {
		return nil, err
	}

	// Override from environment
	if apiURL := os.Getenv("GITSCRUM_API_URL"); apiURL != "" {
		cfg.APIURL = apiURL
	}
	if workspace := viper.GetString("workspace"); workspace != "" {
		cfg.Workspace = workspace
	}
	if project := viper.GetString("project"); project != "" {
		cfg.Project = project
	}

	return cfg, nil
}

// Save writes the configuration to file
func Save(cfg *Config) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ConfigDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	viper.Set("api_url", cfg.APIURL)
	viper.Set("workspace", cfg.Workspace)
	viper.Set("project", cfg.Project)
	viper.Set("oauth.client_id", cfg.OAuth.ClientID)
	viper.Set("oauth.scopes", cfg.OAuth.Scopes)

	configPath := filepath.Join(configDir, ConfigFile)
	return viper.WriteConfigAs(configPath)
}

// Save method saves config (receiver method for compatibility)
func (c *Config) Save() error {
	return Save(c)
}

// DefaultConfig returns a new Config with default values
func DefaultConfig() *Config {
	return &Config{
		APIURL: DefaultAPIURL,
		OAuth: OAuthConfig{
			ClientID: "3e2d1c0b-a9b8-c7d6-e5f4-a3b2c1d0e9f8",
			Scopes:   "cli read write tasks:read tasks:write time-tracking:read time-tracking:write projects:read sprints:read standup:read standup:write",
		},
	}
}

// SetWorkspace sets the default workspace
func SetWorkspace(slug string) error {
	viper.Set("workspace", slug)
	return saveConfig()
}

// SetProject sets the default project
func SetProject(slug string) error {
	viper.Set("project", slug)
	return saveConfig()
}

// saveConfig persists current viper config to file
func saveConfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(home, ConfigDir)
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, ConfigFile)
	return viper.WriteConfigAs(configPath)
}

// ConfigPath returns the path to config file
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDir, ConfigFile), nil
}

// TokenPath returns the path to token file
func TokenPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ConfigDir, "token.json"), nil
}
