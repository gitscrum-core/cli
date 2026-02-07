// Package factory provides dependency injection for commands
package factory

import (
	"sync"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/git"
	"github.com/gitscrum-core/cli/pkg/output"
)

// Factory holds all dependencies for commands
type Factory struct {
	// Output format (table, json, quiet)
	OutputFormat output.Format

	// Lazy-loaded dependencies
	configOnce    sync.Once
	configValue   *config.Config
	configErr     error
	
	apiOnce       sync.Once
	apiClient     *api.Client
	apiErr        error
	
	authOnce      sync.Once
	authToken     *auth.Token
	authErr       error
	
	gitOnce       sync.Once
	gitContext    *git.Context
	gitErr        error
}

// New creates a new Factory with defaults
func New() *Factory {
	return &Factory{
		OutputFormat: output.FormatTable,
	}
}

// Config returns the configuration (lazy loaded)
func (f *Factory) Config() (*config.Config, error) {
	f.configOnce.Do(func() {
		f.configValue, f.configErr = config.Load()
	})
	return f.configValue, f.configErr
}

// APIClient returns the API client (lazy loaded)
func (f *Factory) APIClient() (*api.Client, error) {
	f.apiOnce.Do(func() {
		cfg, err := f.Config()
		if err != nil {
			f.apiErr = err
			return
		}
		
		token, err := f.AuthToken()
		if err != nil {
			f.apiErr = err
			return
		}
		
		f.apiClient = api.NewClient(cfg.APIURL, token)
	})
	return f.apiClient, f.apiErr
}

// AuthToken returns the auth token (lazy loaded)
func (f *Factory) AuthToken() (*auth.Token, error) {
	f.authOnce.Do(func() {
		f.authToken, f.authErr = auth.LoadToken()
	})
	return f.authToken, f.authErr
}

// GitContext returns the current git context (lazy loaded)
func (f *Factory) GitContext() (*git.Context, error) {
	f.gitOnce.Do(func() {
		f.gitContext, f.gitErr = git.ResolveContext(".")
	})
	return f.gitContext, f.gitErr
}

// Formatter returns the appropriate output formatter
func (f *Factory) Formatter() output.Formatter {
	return output.NewFormatter(f.OutputFormat)
}

// RequireAuth ensures user is authenticated
func (f *Factory) RequireAuth() error {
	token, err := f.AuthToken()
	if err != nil {
		return err
	}
	if token == nil || token.AccessToken == "" {
		return auth.ErrNotAuthenticated
	}
	return nil
}

// CurrentWorkspace returns the workspace from flag or config
func (f *Factory) CurrentWorkspace() (string, error) {
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	return cfg.Workspace, nil
}

// CurrentProject returns the project from flag or config
func (f *Factory) CurrentProject() (string, error) {
	cfg, err := f.Config()
	if err != nil {
		return "", err
	}
	return cfg.Project, nil
}

// IsAuthenticated checks if user is authenticated without returning error
func (f *Factory) IsAuthenticated() bool {
	token, err := f.AuthToken()
	if err != nil {
		return false
	}
	return token != nil && token.AccessToken != ""
}
