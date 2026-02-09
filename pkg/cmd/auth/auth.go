// Package auth provides auth commands for GitScrum CLI
package auth

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
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
	// Guard: prevent login if already authenticated
	if f.IsAuthenticated() {
		color.New(color.FgYellow).Println("  You are already logged in.")
		color.New(color.FgHiBlack).Println("  Run 'gitscrum logout' first, then try again.")
		return nil
	}

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

	fmt.Println()
	fmt.Println("Press Enter to open the browser and authorize access...")
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

	// Fetch user profile for personalized message
	// Note: We create a fresh API client because Factory caches the pre-login state (no token)
	userName := ""
	userEmail := ""
	freshClient := api.NewClient(config.DefaultAPIURL, token)
	resp, meErr := freshClient.Get("/me")
	if meErr == nil {
		var user struct {
			Data struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"data"`
		}
		if api.DecodeResponse(resp, &user) == nil {
			userName = user.Data.Name
			userEmail = user.Data.Email
		}
	}

	clearScreen()

	fmt.Println()
	color.New(color.FgHiCyan, color.Bold).Println(`       ██████╗ ██╗████████╗███████╗ ██████╗██████╗ ██╗   ██╗███╗   ███╗`)
	color.New(color.FgHiCyan, color.Bold).Println(`      ██╔════╝ ██║╚══██╔══╝██╔════╝██╔════╝██╔══██╗██║   ██║████╗ ████║`)
	color.New(color.FgCyan, color.Bold).Println(`     ██║  ███╗██║   ██║   ███████╗██║     ██████╔╝██║   ██║██╔████╔██║`)
	color.New(color.FgCyan, color.Bold).Println(`    ██║   ██║██║   ██║   ╚════██║██║     ██╔══██╗██║   ██║██║╚██╔╝██║`)
	color.New(color.FgCyan).Println(`   ╚██████╔╝██║   ██║   ███████║╚██████╗██║  ██║╚██████╔╝██║ ╚═╝ ██║`)
	color.New(color.FgCyan).Println(`    ╚═════╝ ╚═╝   ╚═╝   ╚══════╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝╚═╝     ╚═╝`)
	color.New(color.FgHiBlack).Println(`   ──────────────────────────── Command Line Interface ────────────`)
	fmt.Println()
	color.New(color.FgWhite).Println("  Your entire project workflow — tasks, time tracking, sprints,")
	color.New(color.FgWhite).Println("  invoices, proposals, chat, wiki, and analytics — all without")
	color.New(color.FgWhite).Println("  leaving your terminal. Ship faster, manage smarter.")
	fmt.Println()
	if userName != "" {
		color.New(color.FgGreen, color.Bold).Printf("  ✓ Logged in as: %s", userName)
		if userEmail != "" {
			color.New(color.FgHiBlack).Printf(" (%s)", userEmail)
		}
		fmt.Println()
	} else {
		color.New(color.FgGreen, color.Bold).Println("  ✓ Logged in successfully!")
	}
	fmt.Println()

	// Interactive onboarding
	if err := RunOnboarding(token); err != nil {
		// Non-fatal: show manual instructions as fallback
		color.New(color.FgWhite, color.Bold).Println("  Get started:")
		color.New(color.FgHiBlack).Print("    1. ")
		color.New(color.FgCyan).Println("gitscrum config set workspace <slug>")
		color.New(color.FgHiBlack).Print("    2. ")
		color.New(color.FgCyan).Println("gitscrum config set project <slug>")
		color.New(color.FgHiBlack).Print("    3. ")
		color.New(color.FgCyan).Println("gitscrum tasks")
	}
	fmt.Println()

	return nil
}

// onboardingWorkspace represents a workspace for the selector
type onboardingWorkspace struct {
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	Plan           string `json:"plan"`
	LoggedUserRole struct {
		Label string `json:"label"`
	} `json:"logged_user_role"`
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// onboardingProject represents a project for the selector
type onboardingProject struct {
	Slug     string `json:"slug"`
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
}

// RunOnboarding runs the interactive workspace and project selector
func RunOnboarding(token *auth.Token) error {
	// Step 1: Fetch all workspaces (paginated)
	color.New(color.FgWhite, color.Bold).Println("  Setting up your workspace...")
	fmt.Println()

	sp := spinner.New("Loading workspaces...")
	sp.Start()

	client := api.NewClient(config.DefaultAPIURL, token)

	var allWorkspaces []onboardingWorkspace
	page := 1
	for {
		path := fmt.Sprintf("/companies?page=%d", page)
		resp, err := client.Get(path)
		if err != nil {
			sp.Stop()
			return err
		}

		var wsResult struct {
			Data       []onboardingWorkspace `json:"data"`
			TotalPages int                   `json:"total_pages"`
		}
		if err := api.DecodeResponse(resp, &wsResult); err != nil {
			sp.Stop()
			return err
		}

		allWorkspaces = append(allWorkspaces, wsResult.Data...)

		if page >= wsResult.TotalPages {
			break
		}
		page++
	}
	sp.Stop()

	if len(allWorkspaces) == 0 {
		color.New(color.FgYellow).Println("  No workspaces found. Create one at https://gitscrum.com")
		return nil
	}

	// Build workspace labels
	wsItems := make([]string, len(allWorkspaces))
	for i, w := range allWorkspaces {
		plan := ""
		if w.Plan != "" {
			plan = fmt.Sprintf(" [%s]", strings.ToUpper(w.Plan))
		}
		role := ""
		if w.LoggedUserRole.Label != "" {
			role = fmt.Sprintf("  ·  %s", w.LoggedUserRole.Label)
		}
		wsItems[i] = fmt.Sprintf("%s%s%s", w.Name, plan, role)
	}

	// Scrollable selector with search — loop allows going back from project selection
selectWorkspace:
	wsPrompt := promptui.Select{
		Label: "Select workspace",
		Items: wsItems,
		Size:  7,
		Searcher: func(input string, index int) bool {
			return strings.Contains(
				strings.ToLower(wsItems[index]),
				strings.ToLower(input),
			)
		},
		StartInSearchMode: false,
	}

	wsIndex, _, err := wsPrompt.Run()
	if err != nil {
		return fmt.Errorf("workspace selection cancelled")
	}

	selectedWS := allWorkspaces[wsIndex]
	if err := config.SetWorkspace(selectedWS.Slug); err != nil {
		return err
	}
	color.New(color.FgGreen).Printf("  ✓ Workspace set to '%s'\n", selectedWS.Name)
	fmt.Println()

	// Step 2: Fetch all projects (paginated)
	sp = spinner.New("Loading projects...")
	sp.Start()

	allProjects := make([]onboardingProject, 0)
	page = 1
	for {
		path := fmt.Sprintf("/projects?company_slug=%s&page=%d", selectedWS.Slug, page)
		resp, err := client.Get(path)
		if err != nil {
			sp.Stop()
			return err
		}

		var pjResult struct {
			Data       []onboardingProject `json:"data"`
			TotalPages int                 `json:"total_pages"`
		}
		if err := api.DecodeResponse(resp, &pjResult); err != nil {
			sp.Stop()
			return err
		}

		allProjects = append(allProjects, pjResult.Data...)

		if page >= pjResult.TotalPages || pjResult.TotalPages == 0 {
			break
		}
		page++
	}
	sp.Stop()

	if len(allProjects) == 0 {
		color.New(color.FgYellow).Println("  No projects found in this workspace.")
		fmt.Println()

		emptyPrompt := promptui.Select{
			Label: "What would you like to do?",
			Items: []string{
				"Create a new project",
				"← Back to workspaces",
				"Skip for now",
			},
			Size: 3,
		}

		emptyIdx, _, err := emptyPrompt.Run()
		if err != nil {
			return nil
		}

		switch emptyIdx {
		case 0: // Create project inline
			namePrompt := promptui.Prompt{
				Label: "Project name",
			}
			projectName, err := namePrompt.Run()
			if err != nil || strings.TrimSpace(projectName) == "" {
				return nil
			}

			sp = spinner.New("Creating project...")
			sp.Start()
			body := map[string]interface{}{
				"name": strings.TrimSpace(projectName),
			}
			resp, err := client.Post("/projects", body)
			sp.Stop()
			if err != nil {
				return fmt.Errorf("failed to create project: %w", err)
			}

			var created struct {
				Data onboardingProject `json:"data"`
			}
			if err := api.DecodeResponse(resp, &created); err != nil {
				return err
			}

			if err := config.SetProject(created.Data.Slug); err != nil {
				return err
			}
			color.New(color.FgGreen).Printf("  ✓ Project '%s' created and selected\n", created.Data.Name)
			fmt.Println()

			color.New(color.FgWhite, color.Bold).Println("  You're all set! Try these commands:")
			color.New(color.FgHiBlack).Print("    ▸ ")
			color.New(color.FgCyan).Println("gitscrum tasks")
			color.New(color.FgHiBlack).Print("    ▸ ")
			color.New(color.FgCyan).Println("gitscrum timer start <task-id>")
			color.New(color.FgHiBlack).Print("    ▸ ")
			color.New(color.FgCyan).Println("gitscrum sprints current")
			return nil

		case 1: // Back
			clearScreen()
			goto selectWorkspace

		default: // Skip
			return nil
		}
	}

	// Build project labels with "Back" option
	pjItems := make([]string, len(allProjects)+1)
	pjItems[0] = "← Back to workspaces"
	for i, p := range allProjects {
		status := "●"
		if !p.IsActive {
			status = "○"
		}
		pjItems[i+1] = fmt.Sprintf("%s %s", status, p.Name)
	}

	pjPrompt := promptui.Select{
		Label: "Select project",
		Items: pjItems,
		Size:  7,
		Searcher: func(input string, index int) bool {
			return strings.Contains(
				strings.ToLower(pjItems[index]),
				strings.ToLower(input),
			)
		},
		StartInSearchMode: false,
	}

	pjIndex, _, err := pjPrompt.Run()
	if err != nil {
		return fmt.Errorf("project selection cancelled")
	}

	// If "Back" selected, go back to workspace selection
	if pjIndex == 0 {
		clearScreen()
		goto selectWorkspace
	}

	selectedPJ := allProjects[pjIndex-1]
	if err := config.SetProject(selectedPJ.Slug); err != nil {
		return err
	}
	color.New(color.FgGreen).Printf("  ✓ Project set to '%s'\n", selectedPJ.Name)
	fmt.Println()

	color.New(color.FgWhite, color.Bold).Println("  You're all set! Try these commands:")
	color.New(color.FgHiBlack).Print("    ▸ ")
	color.New(color.FgCyan).Println("gitscrum tasks")
	color.New(color.FgHiBlack).Print("    ▸ ")
	color.New(color.FgCyan).Println("gitscrum timer start <task-id>")
	color.New(color.FgHiBlack).Print("    ▸ ")
	color.New(color.FgCyan).Println("gitscrum sprints current")

	return nil
}

// NewCmdLogout creates the logout command
func NewCmdLogout(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Log out of GitScrum",
		Example: "  gitscrum auth logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, _ := f.AuthToken()
			if token == nil || token.AccessToken == "" {
				yellow := color.New(color.FgYellow)
				yellow.Println("You're not logged in")
				fmt.Println("  Run 'gitscrum auth login' to get started")
				return nil
			}
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
					Username string `json:"username"`
				} `json:"data"`
			}

			if err := api.DecodeResponse(resp, &user); err != nil {
				return err
			}

			if user.Data.Username != "" {
				output.Success(i18n.Tf("logged_in_as", map[string]interface{}{"Name": user.Data.Name, "Username": user.Data.Username}))
			} else {
				output.Success(i18n.Tf("logged_in_as_name", map[string]interface{}{"Name": user.Data.Name}))
			}

			workspace, _ := f.CurrentWorkspace()
			project, _ := f.CurrentProject()

			if workspace != "" {
				output.KeyValue(i18n.T("workspace"), workspace)
			}
			if project != "" {
				output.KeyValue(i18n.T("project"), project)
			}
			if workspace == "" && project == "" {
				output.Dim(i18n.T("no_workspace_selected"))
				output.Dim(i18n.T("run_switch"))
			}

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
			if err != nil || token == nil {
				yellow := color.New(color.FgYellow)
				yellow.Println("✗ Not authenticated")
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
