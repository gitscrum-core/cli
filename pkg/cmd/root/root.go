// Package root contains the root command for GitScrum CLI
package root

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	clierrors "github.com/gitscrum-core/cli/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/analytics"
	"github.com/gitscrum-core/cli/pkg/cmd/auth"
	"github.com/gitscrum-core/cli/pkg/cmd/chat"
	"github.com/gitscrum-core/cli/pkg/cmd/clients"
	"github.com/gitscrum-core/cli/pkg/cmd/config"
	"github.com/gitscrum-core/cli/pkg/cmd/crm"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/cmd/hooks"
	initcmd "github.com/gitscrum-core/cli/pkg/cmd/init"
	"github.com/gitscrum-core/cli/pkg/cmd/invoices"
	"github.com/gitscrum-core/cli/pkg/cmd/notifications"
	"github.com/gitscrum-core/cli/pkg/cmd/projects"
	"github.com/gitscrum-core/cli/pkg/cmd/proposals"
	"github.com/gitscrum-core/cli/pkg/cmd/sprints"
	"github.com/gitscrum-core/cli/pkg/cmd/standup"
	"github.com/gitscrum-core/cli/pkg/cmd/tasks"
	"github.com/gitscrum-core/cli/pkg/cmd/timer"
	"github.com/gitscrum-core/cli/pkg/cmd/wiki"
	"github.com/gitscrum-core/cli/pkg/cmd/workspaces"
	cfgpkg "github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
)

var (
	// Global flags
	cfgFile    string
	jsonOutput bool
	quietMode  bool
)

// Execute runs the root command and returns exit code
func Execute(version, commit, date string) int {
	rootCmd := NewCmdRoot(version, commit, date)
	
	if err := rootCmd.Execute(); err != nil {
		if cliErr, ok := err.(*clierrors.CLIError); ok {
			cliErr.Print()
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		}
		return 1
	}
	return 0
}

// NewCmdRoot creates the root command
func NewCmdRoot(version, commit, date string) *cobra.Command {
	f := factory.New()
	
	cmd := &cobra.Command{
		Use:   "gitscrum <command> [flags]",
		Short: "GitScrum CLI - Project Management from your terminal",
		Long: `GitScrum CLI brings project management to your terminal.
		
Manage tasks, track time, run sprints, and collaborate with your team
without leaving your development environment.

USAGE EXAMPLES:
  $ gitscrum login                    # Authenticate
  $ gitscrum tasks                    # List my tasks
  $ gitscrum tasks current            # Show task for current git branch
  $ gitscrum timer start GS-123       # Start time tracking
  $ gitscrum sprints current          # View current sprint

GETTING STARTED:
  1. Run 'gitscrum login' to authenticate
  2. Run 'gitscrum config set workspace <slug>' to set default workspace
  3. Run 'gitscrum tasks' to see your tasks`,
		SilenceErrors: true,
		SilenceUsage:  true,
		Version:       fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize config
			if err := initConfig(); err != nil {
				return err
			}
			
			// Set output format in factory
			if jsonOutput {
				f.OutputFormat = output.FormatJSON
			} else if quietMode {
				f.OutputFormat = output.FormatQuiet
			}
			
			return nil
		},
	}

	// Persistent flags (global)
	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default ~/.gitscrum/config.yaml)")
	cmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.PersistentFlags().BoolVarP(&quietMode, "quiet", "q", false, "quiet mode (only IDs)")

	// Add subcommands
	cmd.AddCommand(initcmd.NewCmdInit(f))
	cmd.AddCommand(auth.NewCmdAuth(f))
	cmd.AddCommand(config.NewCmdConfig(f))
	cmd.AddCommand(tasks.NewCmdTasks(f))
	cmd.AddCommand(timer.NewCmdTimer(f))
	cmd.AddCommand(sprints.NewCmdSprints(f))
	cmd.AddCommand(standup.NewCmdStandup(f))
	cmd.AddCommand(projects.NewCmdProjects(f))
	cmd.AddCommand(workspaces.NewCmdWorkspaces(f))
	cmd.AddCommand(chat.NewCmdChat(f))
	cmd.AddCommand(wiki.NewCmdWiki(f))
	cmd.AddCommand(notifications.NewCmdNotifications(f))
	cmd.AddCommand(notifications.NewCmdSearch(f))
	cmd.AddCommand(analytics.NewCmdAnalytics(f))
	// CRM commands
	cmd.AddCommand(clients.NewCmdClients(f))
	cmd.AddCommand(invoices.NewCmdInvoices(f))
	cmd.AddCommand(proposals.NewCmdProposals(f))
	cmd.AddCommand(crm.NewCmdCRM(f))
	cmd.AddCommand(hooks.NewCmdHooks(f))
	cmd.AddCommand(NewCmdVersion(version, commit, date))
	cmd.AddCommand(NewCmdCompletion())
	cmd.AddCommand(NewCmdStatus(f))

	// Top-level aliases for better DX
	loginCmd := auth.NewCmdLogin(f)
	loginCmd.Example = "  gitscrum login"
	cmd.AddCommand(loginCmd)

	logoutCmd := auth.NewCmdLogout(f)
	logoutCmd.Example = "  gitscrum logout"
	cmd.AddCommand(logoutCmd)

	cmd.AddCommand(NewCmdSwitch(f))

	return cmd
}

// initConfig reads in config file and ENV variables
func initConfig() error {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		// Search config in ~/.gitscrum directory
		viper.AddConfigPath(home + "/.gitscrum")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	// Read environment variables with GITSCRUM_ prefix
	viper.SetEnvPrefix("GITSCRUM")
	viper.AutomaticEnv()

	// Read config file (ignore if not found)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	// Initialize i18n with configured language
	lang := cfgpkg.GetLanguage()
	i18n.Init(lang)

	return nil
}

// NewCmdVersion creates version command
func NewCmdVersion(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GitScrum CLI %s\n", version)
			fmt.Printf("  Commit: %s\n", commit)
			fmt.Printf("  Built:  %s\n", date)
		},
	}
}

// NewCmdCompletion creates shell completion command
func NewCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for GitScrum CLI.

EXAMPLES:
  # Bash (add to ~/.bashrc)
  $ gitscrum completion bash > /etc/bash_completion.d/gitscrum

  # Zsh (add to ~/.zshrc)
  $ gitscrum completion zsh > "${fpath[1]}/_gitscrum"

  # Fish
  $ gitscrum completion fish > ~/.config/fish/completions/gitscrum.fish

  # PowerShell (add to $PROFILE)
  $ gitscrum completion powershell >> $PROFILE`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}

// NewCmdStatus creates the top-level status command
func NewCmdStatus(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show current authentication and configuration status",
		Long:  "Show the authenticated user, configured workspace, and project.",
		Example: `  gitscrum status`,
		RunE: func(cmd *cobra.Command, args []string) error {
			green := color.New(color.FgGreen)
			yellow := color.New(color.FgYellow)
			gray := color.New(color.FgHiBlack)
			white := color.New(color.FgWhite, color.Bold)

			fmt.Println()
			white.Println("  GitScrum Status")
			gray.Println("  ─────────────────────")

			// Auth status
			if !f.IsAuthenticated() {
				yellow.Println("  ✗ Not authenticated")
				gray.Println("    Run 'gitscrum auth login' to authenticate")
				fmt.Println()
				return nil
			}

			// Fetch user profile
			client, err := f.APIClient()
			if err == nil {
				resp, meErr := client.Get("/me")
				if meErr == nil {
					var user struct {
						Data struct {
							Name  string `json:"name"`
							Email string `json:"email"`
						} `json:"data"`
					}
					if api.DecodeResponse(resp, &user) == nil && user.Data.Name != "" {
						green.Printf("  ✓ Logged in as: ")
						fmt.Printf("%s ", user.Data.Name)
						gray.Printf("(%s)\n", user.Data.Email)
					} else {
						green.Println("  ✓ Authenticated")
					}
				} else {
					green.Println("  ✓ Authenticated")
				}
			}

			// Workspace
			ws, wsErr := f.CurrentWorkspace()
			if wsErr == nil && ws != "" {
				green.Printf("  ✓ Workspace: ")
				fmt.Println(ws)
			} else {
				yellow.Println("  ✗ No workspace configured")
				gray.Println("    Run 'gitscrum config set workspace <slug>'")
			}

			// Project
			pj, pjErr := f.CurrentProject()
			if pjErr == nil && pj != "" {
				green.Printf("  ✓ Project: ")
				fmt.Println(pj)
			} else {
				yellow.Println("  ✗ No project configured")
				gray.Println("    Run 'gitscrum config set project <slug>'")
			}

			fmt.Println()
			return nil
		},
	}
}

// NewCmdSwitch creates the switch command for changing workspace/project
func NewCmdSwitch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "switch",
		Short:   "Switch workspace and project interactively",
		Example: "  gitscrum switch",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := f.RequireAuth(); err != nil {
				return err
			}

			token, err := f.AuthToken()
			if err != nil {
				return err
			}

			return auth.RunOnboarding(token)
		},
	}
}
