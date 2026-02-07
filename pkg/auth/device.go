// Package auth - OAuth 2.0 Device Flow implementation
package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/browser"
)

const (
	// OAuth endpoints (matching Laravel API routes)
	DeviceCodeEndpoint = "/oauth/device/code"
	TokenEndpoint      = "/oauth/device/token"
	
	// Grant types
	GrantTypeDeviceCode = "urn:ietf:params:oauth:grant-type:device_code"
)

// DeviceCodeResponse from device authorization endpoint
type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete,omitempty"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// DeviceAuthenticator handles OAuth Device Flow
type DeviceAuthenticator struct {
	BaseURL    string
	ClientID   string
	Scopes     string
	HTTPClient *http.Client
}

// NewDeviceAuthenticator creates a new authenticator
func NewDeviceAuthenticator(baseURL, clientID, scopes string) *DeviceAuthenticator {
	return &DeviceAuthenticator{
		BaseURL:  baseURL,
		ClientID: clientID,
		Scopes:   scopes,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// StartDeviceFlow initiates the device authorization flow
func (a *DeviceAuthenticator) StartDeviceFlow() (*DeviceCodeResponse, error) {
	data := url.Values{}
	data.Set("client_id", a.ClientID)
	data.Set("scope", a.Scopes)

	resp, err := a.HTTPClient.Post(
		a.BaseURL+DeviceCodeEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device code request failed: %s", string(body))
	}

	var dcr DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&dcr); err != nil {
		return nil, fmt.Errorf("failed to parse device code response: %w", err)
	}

	return &dcr, nil
}

// OpenBrowser opens the verification URL in the default browser
func (a *DeviceAuthenticator) OpenBrowser(verificationURI string) error {
	return browser.OpenURL(verificationURI)
}

// PollForToken polls the token endpoint until authorization is complete
func (a *DeviceAuthenticator) PollForToken(deviceCode string, interval, expiresIn int) (*Token, error) {
	deadline := time.Now().Add(time.Duration(expiresIn) * time.Second)
	pollInterval := time.Duration(interval) * time.Second

	for time.Now().Before(deadline) {
		token, err := a.requestToken(deviceCode)
		if err == nil {
			return token, nil
		}

		// Check for specific errors
		if strings.Contains(err.Error(), "authorization_pending") {
			time.Sleep(pollInterval)
			continue
		}
		if strings.Contains(err.Error(), "slow_down") {
			pollInterval += 5 * time.Second
			time.Sleep(pollInterval)
			continue
		}
		if strings.Contains(err.Error(), "access_denied") {
			return nil, fmt.Errorf("authorization denied by user")
		}
		if strings.Contains(err.Error(), "expired_token") {
			return nil, fmt.Errorf("device code expired. Please try again")
		}

		return nil, err
	}

	return nil, fmt.Errorf("authorization timed out")
}

// requestToken attempts to exchange device code for access token
func (a *DeviceAuthenticator) requestToken(deviceCode string) (*Token, error) {
	data := url.Values{}
	data.Set("grant_type", GrantTypeDeviceCode)
	data.Set("device_code", deviceCode)
	data.Set("client_id", a.ClientID)

	resp, err := a.HTTPClient.Post(
		a.BaseURL+TokenEndpoint,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed: %s", string(body))
	}

	var token Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, err
	}

	token.CreatedAt = time.Now()
	return &token, nil
}
