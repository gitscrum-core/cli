package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// AuthUser represents the authenticated user
type AuthUser struct {
	User
	Workspaces []Workspace `json:"workspaces"`
}

// AuthService handles authentication operations
type AuthService interface {
	// WhoAmI returns the current authenticated user
	WhoAmI() (*AuthUser, error)

	// ValidateToken checks if current token is valid
	ValidateToken() (bool, error)
}

type authService struct {
	client *api.Client
}

func newAuthService(client *api.Client) AuthService {
	return &authService{client: client}
}

func (s *authService) WhoAmI() (*AuthUser, error) {
	resp, err := s.client.Get("/me")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data AuthUser `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *authService) ValidateToken() (bool, error) {
	resp, err := s.client.Get("/me")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200, nil
}
