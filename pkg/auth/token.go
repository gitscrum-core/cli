// Package auth handles authentication for GitScrum CLI
package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gitscrum-core/cli/pkg/config"
)

// Errors
var (
	ErrNotAuthenticated = errors.New("not authenticated. Run 'gitscrum login' to authenticate")
	ErrTokenExpired     = errors.New("token expired. Run 'gitscrum login' to re-authenticate")
)

// Token represents the OAuth token
type Token struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// IsExpired checks if the token has expired
func (t *Token) IsExpired() bool {
	if t.ExpiresIn <= 0 {
		return false // No expiry set
	}
	expiry := t.CreatedAt.Add(time.Duration(t.ExpiresIn) * time.Second)
	return time.Now().After(expiry)
}

// LoadToken reads the token from environment variable or file
// For CI/CD environments, set GITSCRUM_ACCESS_TOKEN environment variable
func LoadToken() (*Token, error) {
	// First, check for environment variable (for CI/CD)
	if accessToken := os.Getenv("GITSCRUM_ACCESS_TOKEN"); accessToken != "" {
		return &Token{
			AccessToken: accessToken,
			TokenType:   "Bearer",
			CreatedAt:   time.Now(),
			ExpiresIn:   0, // No expiry for env tokens
		}, nil
	}

	// Fall back to token file (for local development)
	tokenPath, err := config.TokenPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(tokenPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // No token file = not authenticated
		}
		return nil, err
	}

	var token Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("invalid token file: %w", err)
	}

	if token.IsExpired() {
		return nil, ErrTokenExpired
	}

	return &token, nil
}

// SaveToken writes the token to file
func SaveToken(token *Token) error {
	tokenPath, err := config.TokenPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(tokenPath[:len(tokenPath)-len("token.json")-1], 0700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(tokenPath, data, 0600)
}

// DeleteToken removes the token file
func DeleteToken() error {
	tokenPath, err := config.TokenPath()
	if err != nil {
		return err
	}

	err = os.Remove(tokenPath)
	if os.IsNotExist(err) {
		return nil // Already deleted
	}
	return err
}
