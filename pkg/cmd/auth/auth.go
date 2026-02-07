// Package auth provides auth commands for GitScrum CLI
package auth

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdAuth creates the auth command group
func NewCmdAuth(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth <command>",
		Short: "Authenticate with GitScrum",
		Long: `Manage authentication state for GitScrum CLI.

Authenticate using OAuth Device Flow to connect your CLI to GitScrum.
Your credentials are securely stored locally.`,
		Example: `  # Login to GitScrum
  gitscrum auth login

  # Check who you're logged in as
  gitscrum auth whoami

  # View authentication status
  gitscrum auth status

  # Logout and clear credentials
  gitscrum auth logout`,
	}

	cmd.AddCommand(NewCmdLogin(f))
	cmd.AddCommand(NewCmdLogout(f))
	cmd.AddCommand(NewCmdWhoami(f))
	cmd.AddCommand(NewCmdStatus(f))

	return cmd
}

// NewCmdLogin creates the login command
func NewCmdLogin(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate with GitScrum",
		Long: `Authenticate with GitScrum using OAuth Device Flow.
		
This will open your browser to complete authentication.`,
		Example: "  gitscrum auth login",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLogin(f)
		},
	}
}

func runLogin(f *factory.Factory) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	// Create authenticator
	authenticator := auth.NewDeviceAuthenticator(
		cfg.APIURL,
		cfg.OAuth.ClientID,
		cfg.OAuth.Scopes,
	)

	// Start device flow
	sp := spinner.New("Starting authentication...")
	sp.Start()

	deviceCode, err := authenticator.StartDeviceFlow()
	sp.Stop()
	if err != nil {
		return fmt.Errorf("failed to start authentication: %w", err)
	}

	// Show user code
	fmt.Println()
	fmt.Println("! First, copy your one-time code:", deviceCode.UserCode)
	fmt.Println()
	fmt.Println("Press Enter to open browser and authenticate...")
	fmt.Scanln()

	// Open browser
	verificationURL := deviceCode.VerificationURIComplete
	if verificationURL == "" {
		verificationURL = deviceCode.VerificationURI
	}

	if err := authenticator.OpenBrowser(verificationURL); err != nil {
		fmt.Printf("Could not open browser. Please visit: %s\n", verificationURL)
	}

	// Poll for token
	sp = spinner.New("Waiting for authentication...")
	sp.Start()

	token, err := authenticator.PollForToken(
		deviceCode.DeviceCode,
		deviceCode.Interval,
		deviceCode.ExpiresIn,
	)
	sp.Stop()

	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Save token
	if err := auth.SaveToken(token); err != nil {
		return fmt.Errorf("failed to save token: %w", err)
	}

	fmt.Println()
	fmt.Println("✓ Logged in successfully!")
	fmt.Println()
	fmt.Println("Tip: Run 'gitscrum config set workspace <slug>' to set your default workspace")

	return nil
}

// NewCmdLogout creates the logout command
func NewCmdLogout(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Log out of GitScrum",
		Example: "  gitscrum auth logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := auth.DeleteToken(); err != nil {
				return err
			}
			fmt.Println("✓ Logged out successfully")
			return nil
		},
	}
}

// NewCmdWhoami creates the whoami command
func NewCmdWhoami(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "whoami",
		Short:   "Show authenticated user",
		Example: "  gitscrum auth whoami",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireAuth(); err != nil {
				return err
			}

			client, err := f.APIClient()
			if err != nil {
				return err
			}

			resp, err := client.Get("/me")
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			// Parse user response
			var user struct {
				Data struct {
					Name     string `json:"name"`
					Email    string `json:"email"`
					Username string `json:"username"`
				} `json:"data"`
			}

			if err := api.DecodeResponse(resp, &user); err != nil {
				return err
			}

			fmt.Printf("Logged in as %s (%s)\n", user.Data.Name, user.Data.Email)
			return nil
		},
	}
}

// NewCmdStatus creates the status command
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "status",
		Short:   "Show authentication status",
		Example: "  gitscrum auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := auth.LoadToken()
			if err != nil {
				fmt.Println("✗ Not authenticated")
				fmt.Println("  Run 'gitscrum auth login' to authenticate")
				return nil
			}

			if token == nil {
				fmt.Println("✗ Not authenticated")
				fmt.Println("  Run 'gitscrum auth login' to authenticate")
				return nil
			}

			fmt.Println("✓ Authenticated")
			
			cfg, _ := f.Config()
			if cfg != nil {
				if cfg.Workspace != "" {
					fmt.Printf("  Workspace: %s\n", cfg.Workspace)
				}
				if cfg.Project != "" {
					fmt.Printf("  Project: %s\n", cfg.Project)
				}
			}

			return nil
		},
	}
}
