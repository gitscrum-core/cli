// Package tasks provides git-aware task commands for GitScrum CLI
package tasks

import (
	"fmt"
	"os/exec"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	clierrors "github.com/gitscrum-core/cli/pkg/errors"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// ============================================
// GIT-AWARE COMMANDS
// ============================================

// NewCmdTasksCurrent creates the tasks current command (git-aware)
func NewCmdTasksCurrent(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show task for current git branch",
		Long: `Detect the current git branch and show the linked task.

Branch names like 'feature/GS-123-fix-bug' will be parsed
to find the task code (GS-123).`,
		Example: "  gitscrum tasks current",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksCurrent(f)
		},
	}
}

func runTasksCurrent(f *factory.Factory) error {
	gitCtx, err := f.GitContext()
	if err != nil {
		return clierrors.ErrNotInGitRepo
	}

	if gitCtx.TaskCode == "" {
		fmt.Printf("Branch '%s' does not contain a task code.\n", gitCtx.Branch)
		fmt.Println("Expected format: feature/GS-123-description")
		return nil
	}

	fmt.Printf("Current branch: %s\n", gitCtx.Branch)
	fmt.Printf("Task code: %s\n\n", gitCtx.TaskCode)

	// Fetch and display task details
	return runTasksView(f, gitCtx.TaskCode, false, false)
}

// NewCmdTasksBranch creates the tasks branch command (git-aware)
func NewCmdTasksBranch(f *factory.Factory) *cobra.Command {
	var checkout bool
	var branchName string

	cmd := &cobra.Command{
		Use:   "branch <code>",
		Short: "Create a git branch for a task",
		Long: `Create a git branch on the remote repository and optionally checkout locally.

The branch name will be formatted as: feature/{CODE}-{number}-{slug}
Example: feature/GS-123-fix-authentication-bug`,
		Example: `  gitscrum tasks branch GS-123
  gitscrum tasks branch GS-123 --name my-custom-branch
  gitscrum tasks branch GS-123 --no-checkout`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksBranch(f, args[0], branchName, checkout)
		},
	}

	cmd.Flags().BoolVar(&checkout, "checkout", true, "Checkout branch after creation")
	cmd.Flags().StringVar(&branchName, "name", "", "Custom branch name")

	return cmd
}

func runTasksBranch(f *factory.Factory, code, branchName string, checkout bool) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	gitCtx, err := f.GitContext()
	if err != nil {
		return clierrors.ErrNotInGitRepo
	}

	fmt.Printf("Repository: %s (%s)\n", gitCtx.RepoFullName, gitCtx.Provider)

	sp := spinner.New("Creating branch...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{}
	if branchName != "" {
		body["branch_name"] = branchName
	}

	// Call API to create branch
	resp, err := client.Post(fmt.Sprintf("/integrations/%s/branches/by-code/%s", gitCtx.Provider, code), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Created branch: %s\n", result.Data.Name)

	if checkout {
		fmt.Println("  Fetching and checking out...")

		// Fetch from origin
		if err := exec.Command("git", "fetch", "origin").Run(); err != nil {
			return fmt.Errorf("failed to fetch: %w", err)
		}

		// Checkout the branch
		if err := exec.Command("git", "checkout", result.Data.Name).Run(); err != nil {
			// Try creating tracking branch
			if err := exec.Command("git", "checkout", "-b", result.Data.Name, "origin/"+result.Data.Name).Run(); err != nil {
				return fmt.Errorf("failed to checkout: %w", err)
			}
		}

		fmt.Printf("Switched to branch '%s'\n", result.Data.Name)
	}

	return nil
}

// NewCmdTasksBranches creates the tasks branches command
func NewCmdTasksBranches(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "branches <code>",
		Short:   "List branches linked to a task",
		Example: "  gitscrum tasks branches GS-123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksBranches(f, args[0])
		},
	}
}

func runTasksBranches(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	gitCtx, _ := f.GitContext()
	provider := "github"
	if gitCtx != nil && gitCtx.Provider != "" {
		provider = gitCtx.Provider
	}

	sp := spinner.New("Loading branches...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/integrations/%s/branches/by-code/%s", provider, code))
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			UUID      string `json:"uuid"`
			Name      string `json:"name"`
			URL       string `json:"url"`
			CreatedAt *api.DateResource `json:"created_at"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Printf("No branches linked to %s\n", code)
		return nil
	}

	fmt.Printf("Branches for %s (%d)\n\n", code, len(result.Data))

	formatter := f.Formatter()
	headers := []string{"NAME", "URL"}
	rows := make([][]string, 0, len(result.Data))

	for _, branch := range result.Data {
		rows = append(rows, []string{branch.Name, branch.URL})
	}

	return formatter.PrintTable(headers, rows)
}

// NewCmdTasksPR creates the tasks pr command
func NewCmdTasksPR(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "pr <code>",
		Short: "Open browser to create a pull request",
		Long: `Opens the browser with a pre-filled pull request for the task.

Uses the current branch as the source and the default branch as target.`,
		Example: "  gitscrum tasks pr GS-123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksPR(f, args[0])
		},
	}
}

func runTasksPR(f *factory.Factory, code string) error {
	gitCtx, err := f.GitContext()
	if err != nil {
		return clierrors.ErrNotInGitRepo
	}

	if gitCtx.Branch == "" || gitCtx.Branch == "main" || gitCtx.Branch == "master" {
		return fmt.Errorf("cannot create PR from %s branch", gitCtx.Branch)
	}

	// Build PR URL based on provider
	var prURL string
	switch gitCtx.Provider {
	case "github":
		prURL = fmt.Sprintf("https://github.com/%s/compare/%s?expand=1&title=%s",
			gitCtx.RepoFullName, gitCtx.Branch, code)
	case "gitlab":
		prURL = fmt.Sprintf("https://gitlab.com/%s/-/merge_requests/new?merge_request[source_branch]=%s&merge_request[title]=%s",
			gitCtx.RepoFullName, gitCtx.Branch, code)
	case "bitbucket":
		prURL = fmt.Sprintf("https://bitbucket.org/%s/pull-requests/new?source=%s&t=1",
			gitCtx.RepoFullName, gitCtx.Branch)
	default:
		return fmt.Errorf("unsupported git provider: %s", gitCtx.Provider)
	}

	fmt.Printf("Opening pull request creation for %s...\n", code)
	fmt.Printf("  Branch: %s\n", gitCtx.Branch)

	return browser.OpenURL(prURL)
}

// NewCmdTasksPRs creates the tasks prs command
func NewCmdTasksPRs(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "prs <code>",
		Short:   "List pull requests linked to a task",
		Example: "  gitscrum tasks prs GS-123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksPRs(f, args[0])
		},
	}
}

func runTasksPRs(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	gitCtx, _ := f.GitContext()
	provider := "github"
	if gitCtx != nil && gitCtx.Provider != "" {
		provider = gitCtx.Provider
	}

	sp := spinner.New("Loading pull requests...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	resp, err := client.Get(fmt.Sprintf("/integrations/%s/pull-requests/by-code/%s", provider, code))
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			UUID   string `json:"uuid"`
			Title  string `json:"title"`
			Number int    `json:"number"`
			State  string `json:"state"`
			URL    string `json:"url"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Printf("No pull requests linked to %s\n", code)
		return nil
	}

	fmt.Printf("Pull Requests for %s (%d)\n\n", code, len(result.Data))

	formatter := f.Formatter()
	headers := []string{"#", "TITLE", "STATE", "URL"}
	rows := make([][]string, 0, len(result.Data))

	for _, pr := range result.Data {
		title := pr.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}
		rows = append(rows, []string{
			fmt.Sprintf("#%d", pr.Number),
			title,
			pr.State,
			pr.URL,
		})
	}

	return formatter.PrintTable(headers, rows)
}

// NewCmdTasksUnlinkBranch creates the unlink-branch command
func NewCmdTasksUnlinkBranch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "unlink-branch <branch-uuid>",
		Short:   "Remove link between task and branch",
		Example: "  gitscrum tasks unlink-branch abc123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksUnlinkBranch(f, args[0])
		},
	}
}

func runTasksUnlinkBranch(f *factory.Factory, branchUUID string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	gitCtx, _ := f.GitContext()
	provider := "github"
	if gitCtx != nil && gitCtx.Provider != "" {
		provider = gitCtx.Provider
	}

	sp := spinner.New("Unlinking branch...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	resp, err := client.Delete(fmt.Sprintf("/integrations/%s/branches/%s", provider, branchUUID))
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Branch unlinked\n")

	return nil
}
