// Package init provides the init command for GitScrum CLI
package init

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/config/projectconfig"
	"github.com/gitscrum-core/cli/pkg/git"
)

// NewCmdInit creates the init command
func NewCmdInit(f *factory.Factory) *cobra.Command {
	var workspace string
	var project string
	var skipAuth bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize GitScrum in current directory",
		Long: `Set up GitScrum CLI in the current directory.

This command will:
1. Detect if you're in a git repository
2. Check authentication status (or prompt to login)
3. Set up default workspace and project
4. Optionally install git hooks`,
		Example: `  gitscrum init
  gitscrum init --workspace my-workspace --project my-project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(f, workspace, project, skipAuth)
		},
	}

	cmd.Flags().StringVarP(&workspace, "workspace", "w", "", "Default workspace slug")
	cmd.Flags().StringVarP(&project, "project", "p", "", "Default project slug")
	cmd.Flags().BoolVar(&skipAuth, "skip-auth", false, "Skip authentication check")

	return cmd
}

func runInit(f *factory.Factory, workspace, project string, skipAuth bool) error {
	green := color.New(color.FgGreen, color.Bold)
	cyan := color.New(color.FgCyan)
	yellow := color.New(color.FgYellow)
	
	cyan.Println("Initializing GitScrum CLI...")
	fmt.Println()

	// Step 1: Detect git repository
	fmt.Print("Checking git repository... ")
	gitCtx, err := git.NewContext(".")
	if err != nil {
		yellow.Println("[!] Not found")
		fmt.Println("   This directory is not a git repository.")
		fmt.Println("   GitScrum can still work, but git-aware features will be disabled.")
	} else {
		green.Println("ok")
		fmt.Printf("   Repository: %s\n", gitCtx.RepoFullName)
		fmt.Printf("   Provider: %s\n", gitCtx.Provider)
		fmt.Printf("   Branch: %s\n", gitCtx.Branch)
	}
	fmt.Println()

	// Step 2: Check authentication
	if !skipAuth {
		fmt.Print("Checking authentication... ")
		if f.IsAuthenticated() {
			green.Println("ok")
		} else {
			yellow.Println("[!] Not authenticated")
			fmt.Println("   Run 'gitscrum auth login' to authenticate")
		}
		fmt.Println()
	}

	// Step 3: Set up global configuration
	fmt.Println("Global Configuration:")

	cfg, err := f.Config()
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if workspace != "" {
		cfg.Workspace = workspace
	}
	if project != "" {
		cfg.Project = project
	}
	
	// Use global config values if flags not provided
	if workspace == "" {
		workspace = cfg.Workspace
	}
	if project == "" {
		project = cfg.Project
	}

	fmt.Printf("   Workspace: %s\n", valueOrNotSet(cfg.Workspace))
	fmt.Printf("   Project:   %s\n", valueOrNotSet(cfg.Project))

	// Save global config if we have updates
	if workspace != "" || project != "" {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}
	fmt.Println()

	// Step 4: Create .gitscrum.yml project config
	fmt.Print("Creating project config... ")
	
	existingPath := projectconfig.GetConfigPath()
	if existingPath != "" {
		yellow.Println("[exists]")
		fmt.Printf("   Found: %s\n", existingPath)
	} else if workspace != "" {
		// Create project config
		projCfg, err := projectconfig.Init(workspace, project)
		if err != nil {
			yellow.Printf("[!] Failed: %v\n", err)
		} else {
			green.Println("ok")
			fmt.Printf("   Created: %s\n", projectconfig.ConfigFileName)
			fmt.Println()
			fmt.Println("   Project Config Settings:")
			fmt.Printf("     Branch prefix:     %s/\n", projCfg.Branch.DefaultPrefix)
			fmt.Printf("     Prepend task code: %v\n", projCfg.Hooks.PrependTaskCode)
			fmt.Printf("     Auto-timer:        %v\n", projCfg.Timer.AutoStart)
			fmt.Printf("     Complete on merge: %v\n", projCfg.Automation.CompleteOnMerge)
		}
	} else {
		yellow.Println("[skip]")
		fmt.Println("   Set workspace first: gitscrum config set workspace <slug>")
	}
	fmt.Println()

	// Step 5: Create .gitscrum directory for local cache if in git repo
	if gitCtx != nil {
		gitscrumDir := ".gitscrum"
		if _, err := os.Stat(gitscrumDir); os.IsNotExist(err) {
			if err := os.MkdirAll(gitscrumDir, 0755); err == nil {
				fmt.Printf("Created %s/ directory (local cache)\n", gitscrumDir)

				// Add to .gitignore if not present
				gitignore := ".gitignore"
				content, _ := os.ReadFile(gitignore)
				if !contains(string(content), ".gitscrum/") {
					file, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						file.WriteString("\n# GitScrum CLI cache\n.gitscrum/\n")
						file.Close()
						fmt.Println("   Added .gitscrum/ to .gitignore")
					}
				}
				fmt.Println()
			}
		}
	}

	green.Println("✅ GitScrum CLI initialized!")
	fmt.Println()
	fmt.Println("Next steps:")
	
	step := 1
	if !f.IsAuthenticated() {
		fmt.Printf("  %d. gitscrum auth login     # Authenticate\n", step)
		step++
	}
	if cfg.Workspace == "" {
		fmt.Printf("  %d. gitscrum config set workspace <slug>\n", step)
		step++
	}
	if existingPath == "" && workspace != "" {
		fmt.Printf("  %d. Review and commit .gitscrum.yml\n", step)
		step++
	}
	fmt.Printf("  %d. gitscrum tasks          # View your tasks\n", step)
	step++
	fmt.Printf("  %d. gitscrum hooks install  # Install git hooks (optional)\n", step)

	return nil
}

func valueOrNotSet(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
