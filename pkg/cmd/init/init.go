// Package init provides the init command for GitScrum CLI
package init

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/config"
	"github.com/gitscrum-core/cli/pkg/config/projectconfig"
	"github.com/gitscrum-core/cli/pkg/git"
	"github.com/gitscrum-core/cli/pkg/output"
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
	output.Header("Initializing GitScrum CLI")

	// Step 1: Detect git repository
	fmt.Print("  Checking git repository... ")
	gitCtx, err := git.NewContext(".")
	if err != nil {
		output.Warning("Not found")
		output.Dim("  This directory is not a git repository.")
		output.Dim("  GitScrum can still work, but git-aware features will be disabled.")
	} else {
		output.Success("ok")
		output.KeyValue("Repository", gitCtx.RepoFullName)
		output.KeyValue("Provider", gitCtx.Provider)
		output.KeyValue("Branch", gitCtx.Branch)
	}
	fmt.Println()

	// Step 2: Check authentication
	if !skipAuth {
		fmt.Print("  Checking authentication... ")
		if f.IsAuthenticated() {
			output.Success("ok")
		} else {
			output.Warning("Not authenticated")
			output.Infof("Run 'gitscrum auth login' to authenticate")
		}
		fmt.Println()
	}

	// Step 3: Set up global configuration
	output.SubHeader("Global Configuration")

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

	output.KeyValue("Workspace", valueOrNotSet(cfg.Workspace))
	output.KeyValue("Project", valueOrNotSet(cfg.Project))

	// Save global config if we have updates
	if workspace != "" || project != "" {
		if err := cfg.Save(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}
	}
	fmt.Println()

	// Step 4: Create .gitscrum.yml project config
	fmt.Print("  Creating project config... ")

	existingPath := projectconfig.GetConfigPath()
	if existingPath != "" {
		output.Warning("exists")
		output.Dimf("Found: %s", existingPath)
	} else if workspace != "" {
		// Create project config
		projCfg, err := projectconfig.Init(workspace, project)
		if err != nil {
			output.Warningf("Failed: %v", err)
		} else {
			output.Success("ok")
			output.Dimf("Created: %s", projectconfig.ConfigFileName)
			fmt.Println()
			output.SubHeader("Project Config Settings")
			output.KeyValuef("Branch prefix", "%s/", projCfg.Branch.DefaultPrefix)
			output.KeyValuef("Prepend task code", "%v", projCfg.Hooks.PrependTaskCode)
			output.KeyValuef("Auto-timer", "%v", projCfg.Timer.AutoStart)
			output.KeyValuef("Complete on merge", "%v", projCfg.Automation.CompleteOnMerge)
		}
	} else {
		output.Info("skip")
		output.Infof("Set workspace first: gitscrum config set workspace <slug>")
	}
	fmt.Println()

	// Step 5: Create .gitscrum directory for local cache if in git repo
	if gitCtx != nil {
		gitscrumDir := ".gitscrum"
		if _, err := os.Stat(gitscrumDir); os.IsNotExist(err) {
			if err := os.MkdirAll(gitscrumDir, 0755); err == nil {
				output.Infof("Created %s/ directory (local cache)", gitscrumDir)

				// Add to .gitignore if not present
				gitignore := ".gitignore"
				content, _ := os.ReadFile(gitignore)
				if !strings.Contains(string(content), ".gitscrum/") {
					file, err := os.OpenFile(gitignore, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					if err == nil {
						file.WriteString("\n# GitScrum CLI cache\n.gitscrum/\n")
						file.Close()
						output.Dim("  Added .gitscrum/ to .gitignore")
					}
				}
				fmt.Println()
			}
		}
	}

	output.Success("GitScrum CLI initialized!")
	fmt.Println()
	output.SubHeader("Next Steps")

	step := 1
	if !f.IsAuthenticated() {
		output.Bulletf("%d. gitscrum auth login     # Authenticate", step)
		step++
	}
	if cfg.Workspace == "" {
		output.Bulletf("%d. gitscrum config set workspace <slug>", step)
		step++
	}
	if existingPath == "" && workspace != "" {
		output.Bulletf("%d. Review and commit .gitscrum.yml", step)
		step++
	}
	output.Bulletf("%d. gitscrum tasks          # View your tasks", step)
	step++
	output.Bulletf("%d. gitscrum hooks install  # Install git hooks (optional)", step)

	fmt.Println()
	return nil
}

func valueOrNotSet(s string) string {
	if s == "" {
		return "(not set)"
	}
	return s
}
