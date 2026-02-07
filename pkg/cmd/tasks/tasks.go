// Package tasks provides task commands for GitScrum CLI
package tasks

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdTasks creates the tasks command group
func NewCmdTasks(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks [command]",
		Short: "Manage tasks",
		Long: `View and manage tasks in GitScrum.

Create, view, update, and organize your tasks. Integrates with Git to
automatically detect tasks from branch names and link branches/PRs.

Without a subcommand, lists your assigned tasks.`,
		Example: `  # List your assigned tasks
  gitscrum tasks

  # List tasks from a specific project
  gitscrum tasks list -p my-project

  # View task details
  gitscrum tasks view GS-123

  # Create a new task
  gitscrum tasks create -t "Fix login bug" -p my-project

  # Mark task as complete
  gitscrum tasks complete GS-123

  # Create branch for task (Git-aware)
  gitscrum tasks branch GS-123

  # View task from current branch
  gitscrum tasks current`,
		Aliases: []string{"task", "t"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksList(f, &ListOptions{})
		},
	}

	// Core commands
	cmd.AddCommand(NewCmdTasksList(f))
	cmd.AddCommand(NewCmdTasksToday(f))
	cmd.AddCommand(NewCmdTasksView(f))
	cmd.AddCommand(NewCmdTasksCreate(f))
	cmd.AddCommand(NewCmdTasksUpdate(f))
	cmd.AddCommand(NewCmdTasksComplete(f))
	cmd.AddCommand(NewCmdTasksAssign(f))

	// Git-aware commands
	cmd.AddCommand(NewCmdTasksCurrent(f))
	cmd.AddCommand(NewCmdTasksBranch(f))
	cmd.AddCommand(NewCmdTasksBranches(f))
	cmd.AddCommand(NewCmdTasksPR(f))
	cmd.AddCommand(NewCmdTasksPRs(f))
	cmd.AddCommand(NewCmdTasksUnlinkBranch(f))

	// Advanced commands
	cmd.AddCommand(NewCmdTasksMove(f))
	cmd.AddCommand(NewCmdTasksDuplicate(f))
	cmd.AddCommand(NewCmdTasksSubtasks(f))

	return cmd
}

// ListOptions for tasks list command
type ListOptions struct {
	Project  string
	Assignee string
	Filter   string
	Status   string
	Limit    int
}

// NewCmdTasksList creates the tasks list command
func NewCmdTasksList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		Example: `  gitscrum tasks list
  gitscrum tasks list -p my-project
  gitscrum tasks list --filter blocker
  gitscrum tasks list --assignee @user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksList(f, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.Project, "project", "p", "", "Filter by project")
	cmd.Flags().StringVar(&opts.Assignee, "assignee", "", "Filter by assignee (@username)")
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter (blocker, bug, feature)")
	cmd.Flags().StringVar(&opts.Status, "status", "", "Filter by status column")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n", 20, "Maximum tasks to show")

	return cmd
}

func runTasksList(f *factory.Factory, opts *ListOptions) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading tasks...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Build query path
	path := "/issues/my"
	if opts.Project != "" {
		path = fmt.Sprintf("/projects/%s/issues", opts.Project)
	}

	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse response
	var result struct {
		Data []struct {
			UUID      string `json:"uuid"`
			Number    int    `json:"number"`
			Title     string `json:"title"`
			Code      string `json:"code"`
			Workflow  string `json:"config_workflow_title"`
			Assignees []struct {
				Username string `json:"username"`
			} `json:"users"`
			Project struct {
				Code string `json:"code"`
			} `json:"project"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	formatter := f.Formatter()

	if len(result.Data) == 0 {
		fmt.Println("No tasks found")
		return nil
	}

	headers := []string{"CODE", "TITLE", "STATUS", "ASSIGNEE"}
	rows := make([][]string, 0, len(result.Data))

	for _, task := range result.Data {
		code := task.Code
		if code == "" {
			code = fmt.Sprintf("%s-%d", task.Project.Code, task.Number)
		}

		assignee := ""
		if len(task.Assignees) > 0 {
			assignee = "@" + task.Assignees[0].Username
		}

		// Apply filters
		if opts.Assignee != "" && !strings.Contains(assignee, opts.Assignee) {
			continue
		}

		title := task.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}

		rows = append(rows, []string{code, title, task.Workflow, assignee})

		if len(rows) >= opts.Limit {
			break
		}
	}

	return formatter.PrintTable(headers, rows)
}

// NewCmdTasksToday creates the tasks today command
func NewCmdTasksToday(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "today",
		Short:   "List tasks due today",
		Example: "  gitscrum tasks today",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksToday(f)
		},
	}
}

func runTasksToday(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading today's tasks...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	resp, err := client.Get("/issues/my?due=today")
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data []struct {
			Code     string `json:"code"`
			Title    string `json:"title"`
			Workflow string `json:"config_workflow_title"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	formatter := f.Formatter()

	if len(result.Data) == 0 {
		fmt.Println("No tasks due today")
		return nil
	}

	fmt.Printf("TASKS DUE TODAY (%d)\n\n", len(result.Data))

	headers := []string{"CODE", "TITLE", "STATUS"}
	rows := make([][]string, 0, len(result.Data))

	for _, task := range result.Data {
		title := task.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		rows = append(rows, []string{task.Code, title, task.Workflow})
	}

	return formatter.PrintTable(headers, rows)
}

// NewCmdTasksView creates the tasks view command
func NewCmdTasksView(f *factory.Factory) *cobra.Command {
	var web bool

	cmd := &cobra.Command{
		Use:     "view <code>",
		Short:   "View task details",
		Example: "  gitscrum tasks view GS-123\n  gitscrum tasks view GS-123 --web",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if web {
				return runTasksViewWeb(f, args[0])
			}
			return runTasksView(f, args[0])
		},
	}

	cmd.Flags().BoolVarP(&web, "web", "w", false, "Open in browser")

	return cmd
}

func runTasksView(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Parse code to get project and number
	parts := strings.SplitN(code, "-", 2)
	if len(parts) != 2 {
		sp.Stop()
		return fmt.Errorf("invalid task code format: %s (expected: XX-123)", code)
	}

	resp, err := client.Get(fmt.Sprintf("/issues/by-code/%s", code))
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UUID        string `json:"uuid"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Number      int    `json:"number"`
			Workflow    string `json:"config_workflow_title"`
			Type        string `json:"config_issue_type_title"`
			Priority    string `json:"config_priority_title"`
			DueDate     string `json:"due_date"`
			Effort      int    `json:"effort"`
			Assignees   []struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"users"`
			Labels []struct {
				Title string `json:"title"`
				Color string `json:"color"`
			} `json:"labels"`
			Project struct {
				Title string `json:"title"`
				Code  string `json:"code"`
			} `json:"project"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	task := result.Data

	// Display task details
	fmt.Printf("\n%s-%d: %s\n", task.Project.Code, task.Number, task.Title)
	fmt.Println(strings.Repeat("─", 60))

	fmt.Printf("Status:   %s\n", task.Workflow)
	if task.Type != "" {
		fmt.Printf("Type:     %s\n", task.Type)
	}
	if task.Priority != "" {
		fmt.Printf("Priority: %s\n", task.Priority)
	}
	if task.DueDate != "" {
		fmt.Printf("Due:      %s\n", task.DueDate)
	}
	if task.Effort > 0 {
		fmt.Printf("Effort:   %d points\n", task.Effort)
	}

	if len(task.Assignees) > 0 {
		assignees := make([]string, len(task.Assignees))
		for i, a := range task.Assignees {
			assignees[i] = "@" + a.Username
		}
		fmt.Printf("Assigned: %s\n", strings.Join(assignees, ", "))
	}

	if len(task.Labels) > 0 {
		labels := make([]string, len(task.Labels))
		for i, l := range task.Labels {
			labels[i] = l.Title
		}
		fmt.Printf("Labels:   %s\n", strings.Join(labels, ", "))
	}

	if task.Description != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println(task.Description)
	}

	fmt.Println()

	return nil
}

func runTasksViewWeb(f *factory.Factory, code string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	// Build web URL
	url := fmt.Sprintf("%s/tasks/%s", strings.Replace(cfg.APIURL, "api.", "app.", 1), code)

	fmt.Printf("Opening %s in browser...\n", url)
	return browser.OpenURL(url)
}

// NewCmdTasksCreate creates the tasks create command
func NewCmdTasksCreate(f *factory.Factory) *cobra.Command {
	var project string
	var description string
	var assignee string
	var taskType string
	var priority string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Example: `  gitscrum tasks create "Fix authentication bug" -p my-project
  gitscrum tasks create "New feature" -p my-project -a @john --type feature`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksCreate(f, args[0], project, description, assignee, taskType, priority)
		},
	}

	cmd.Flags().StringVarP(&project, "project", "p", "", "Project slug (required)")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Task description")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assign to user (@username)")
	cmd.Flags().StringVarP(&taskType, "type", "t", "", "Task type (bug, feature, task)")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority (low, medium, high, critical)")

	cmd.MarkFlagRequired("project")

	return cmd
}

func runTasksCreate(f *factory.Factory, title, project, description, assignee, taskType, priority string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	if project == "" {
		// Try to get default project
		cfg, err := f.Config()
		if err != nil || cfg.Project == "" {
			return fmt.Errorf("project is required. Use -p flag or set default: gitscrum config set project <slug>")
		}
		project = cfg.Project
	}

	sp := spinner.New("Creating task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"title": title,
	}
	if description != "" {
		body["description"] = description
	}

	resp, err := client.Post(fmt.Sprintf("/projects/%s/issues", project), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UUID   string `json:"uuid"`
			Number int    `json:"number"`
			Title  string `json:"title"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Created task: %s (ID: %d)\n", result.Data.Title, result.Data.Number)

	return nil
}

// NewCmdTasksUpdate creates the tasks update command
func NewCmdTasksUpdate(f *factory.Factory) *cobra.Command {
	var column string
	var title string
	var description string

	cmd := &cobra.Command{
		Use:   "update <code>",
		Short: "Update a task",
		Example: `  gitscrum tasks update GS-123 --column="In Progress"
  gitscrum tasks update GS-123 --title="New title"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksUpdate(f, args[0], column, title, description)
		},
	}

	cmd.Flags().StringVar(&column, "column", "", "Move to status column")
	cmd.Flags().StringVar(&title, "title", "", "Update title")
	cmd.Flags().StringVarP(&description, "description", "d", "", "Update description")

	return cmd
}

func runTasksUpdate(f *factory.Factory, code, column, title, description string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Updating task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := make(map[string]interface{})
	if column != "" {
		body["config_workflow_title"] = column
	}
	if title != "" {
		body["title"] = title
	}
	if description != "" {
		body["description"] = description
	}

	if len(body) == 0 {
		sp.Stop()
		return fmt.Errorf("no updates specified. Use --column, --title, or --description")
	}

	resp, err := client.Patch(fmt.Sprintf("/issues/by-code/%s", code), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Updated task %s\n", code)

	return nil
}

// NewCmdTasksComplete creates the tasks complete command
func NewCmdTasksComplete(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "complete <code>",
		Short:   "Mark task as done",
		Aliases: []string{"done"},
		Example: "  gitscrum tasks complete GS-123",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksComplete(f, args[0])
		},
	}
}

func runTasksComplete(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Completing task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"is_closed": true,
	}

	resp, err := client.Patch(fmt.Sprintf("/issues/by-code/%s", code), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Marked %s as done\n", code)

	return nil
}

// NewCmdTasksAssign creates the tasks assign command
func NewCmdTasksAssign(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "assign <code> <@username>",
		Short:   "Assign task to user",
		Example: "  gitscrum tasks assign GS-123 @john",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksAssign(f, args[0], args[1])
		},
	}
}

func runTasksAssign(f *factory.Factory, code, username string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	username = strings.TrimPrefix(username, "@")

	sp := spinner.New("Assigning task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"username": username,
	}

	resp, err := client.Post(fmt.Sprintf("/issues/by-code/%s/assign", code), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Assigned %s to @%s\n", code, username)

	return nil
}

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
		return fmt.Errorf("not in a git repository: %w", err)
	}

	if gitCtx.TaskCode == "" {
		fmt.Printf("Branch '%s' does not contain a task code.\n", gitCtx.Branch)
		fmt.Println("Expected format: feature/GS-123-description")
		return nil
	}

	fmt.Printf("Current branch: %s\n", gitCtx.Branch)
	fmt.Printf("Task code: %s\n\n", gitCtx.TaskCode)

	// Fetch and display task details
	return runTasksView(f, gitCtx.TaskCode)
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
		return fmt.Errorf("not in a git repository: %w", err)
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
			CreatedAt string `json:"created_at"`
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
		return fmt.Errorf("not in a git repository: %w", err)
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

// NewCmdTasksMove creates the tasks move command
func NewCmdTasksMove(f *factory.Factory) *cobra.Command {
	var toProject string

	cmd := &cobra.Command{
		Use:   "move <code>",
		Short: "Move task to another project",
		Example: `  gitscrum tasks move GS-123 --to=other-project
  gitscrum tasks move GS-123 --to="My New Project"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksMove(f, args[0], toProject)
		},
	}

	cmd.Flags().StringVar(&toProject, "to", "", "Destination project slug (required)")
	cmd.MarkFlagRequired("to")

	return cmd
}

func runTasksMove(f *factory.Factory, code, toProject string) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	sp := spinner.New("Moving task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// First, fetch the task by code to get its UUID
	lookupPath := fmt.Sprintf("/tasks/by-code/%s", code)
	lookupResp, err := client.Get(lookupPath)
	if err != nil {
		sp.Stop()
		return err
	}

	var taskData struct {
		Data struct {
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(lookupResp, &taskData); err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"new_project_slug": toProject,
	}

	path := fmt.Sprintf("/tasks/%s/move", taskData.Data.UUID)
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Message string `json:"message"`
		Status  string `json:"status"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Task %s move initiated to %s\n", code, toProject)

	return nil
}

// NewCmdTasksDuplicate creates the tasks duplicate command
func NewCmdTasksDuplicate(f *factory.Factory) *cobra.Command {
	var toProject string
	var withSubtasks bool

	cmd := &cobra.Command{
		Use:   "duplicate <code>",
		Short: "Duplicate a task",
		Example: `  gitscrum tasks duplicate GS-123
  gitscrum tasks duplicate GS-123 --to=other-project
  gitscrum tasks duplicate GS-123 --with-subtasks`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksDuplicate(f, args[0], toProject, withSubtasks)
		},
	}

	cmd.Flags().StringVar(&toProject, "to", "", "Destination project slug (optional)")
	cmd.Flags().BoolVar(&withSubtasks, "with-subtasks", false, "Include subtasks")

	return cmd
}

func runTasksDuplicate(f *factory.Factory, code, toProject string, withSubtasks bool) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	sp := spinner.New("Duplicating task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// First, fetch the task by code to get its UUID
	lookupPath := fmt.Sprintf("/tasks/by-code/%s", code)
	lookupResp, err := client.Get(lookupPath)
	if err != nil {
		sp.Stop()
		return err
	}

	var taskData struct {
		Data struct {
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(lookupResp, &taskData); err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"include_subtasks": withSubtasks,
	}
	if toProject != "" {
		body["project_slug"] = toProject
	}

	path := fmt.Sprintf("/tasks/%s/duplicate", taskData.Data.UUID)
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Code  string `json:"code"`
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Task duplicated: %s -> %s\n", code, result.Data.Code)
	fmt.Printf("  %s\n", result.Data.Title)

	return nil
}

// NewCmdTasksSubtasks creates the tasks subtasks command
func NewCmdTasksSubtasks(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "subtasks <code>",
		Short: "View subtasks of a task",
		Example: `  gitscrum tasks subtasks GS-123`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksSubtasks(f, args[0])
		},
	}
}

func runTasksSubtasks(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// First, fetch the task by code to get its UUID
	lookupPath := fmt.Sprintf("/tasks/by-code/%s", code)
	lookupResp, err := client.Get(lookupPath)
	if err != nil {
		return err
	}

	var taskData struct {
		Data struct {
			UUID string `json:"uuid"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(lookupResp, &taskData); err != nil {
		return err
	}

	path := fmt.Sprintf("/tasks/%s/sub-tasks", taskData.Data.UUID)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Code   string `json:"code"`
			Title  string `json:"title"`
			Status struct {
				Title string `json:"title"`
				Color string `json:"color"`
			} `json:"status"`
			Assignee struct {
				Name string `json:"name"`
			} `json:"assignee"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("SUBTASKS OF %s:\n\n", code)

	if len(result.Data) == 0 {
		fmt.Println("  No subtasks found")
		fmt.Println()
		fmt.Printf("  Create one with: gitscrum tasks create \"Subtask title\" --parent=%s\n", code)
		return nil
	}

	for _, t := range result.Data {
		status := getStatusIcon(t.Status.Title)
		assignee := ""
		if t.Assignee.Name != "" {
			assignee = fmt.Sprintf(" @%s", t.Assignee.Name)
		}
		fmt.Printf("  %s [%s] %s%s\n", status, t.Code, t.Title, assignee)
	}

	return nil
}

func getStatusIcon(status string) string {
	switch strings.ToLower(status) {
	case "done", "completed", "closed":
		return "[x]"
	case "in progress", "doing":
		return "[>]"
	case "todo", "to do", "open":
		return "[ ]"
	case "blocked":
		return "[!]"
	case "review", "in review":
		return "[~]"
	default:
		return "[-]"
	}
}

