// Package tasks provides task commands for GitScrum CLI
package tasks

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pkg/browser"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdTasks creates the tasks command group

func NewCmdTasks(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{Page: 1, Limit: 20}

	cmd := &cobra.Command{
		Use:   "tasks [command]",
		Short: "Manage tasks",
		Long: `View and manage tasks in GitScrum.

Create, view, update, and organize your tasks. Integrates with Git to
automatically detect tasks from branch names and link branches/PRs.

Without a subcommand, lists your assigned tasks.`,
		Example: `  # List your assigned tasks
  gitscrum tasks
  gitscrum tasks --page 2

  # View task details
  gitscrum tasks view a1b2c3d4

  # View task comments
  gitscrum tasks view a1b2c3d4 --comments

  # View time entries
  gitscrum tasks view a1b2c3d4 --timers

  # Add a comment to a task
  gitscrum tasks comment a1b2c3d4 -m "Your message"

  # Create a new task
  gitscrum tasks create -t "Fix login bug"

  # Mark task as complete
  gitscrum tasks complete a1b2c3d4

  # Create branch for task (Git-aware)
  gitscrum tasks branch a1b2c3d4

  # View task from current branch
  gitscrum tasks current`,
		Aliases: []string{"task", "t"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksList(f, opts)
		},
	}

	// Pagination and filter flags for main command
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number for pagination")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n", 20, "Maximum tasks to show")
	cmd.Flags().StringVar(&opts.Assignee, "assignee", "", "Filter by assignee (@username)")
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter (blocker, bug, draft, unassigned)")
	cmd.Flags().StringVar(&opts.Workflow, "workflow", "", "Filter by workflow (todo, in-progress, done)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by type (feature, bug, improvement)")
	cmd.Flags().StringVar(&opts.Board, "board", "", "Filter by board (uuid or name)")

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
	cmd.AddCommand(NewCmdTasksComment(f))

	return cmd
}

// ListOptions for tasks list command
type ListOptions struct {
	Project  string
	Assignee string
	Filter   string
	Workflow string
	Type     string
	Board    string
	Limit    int
	Page     int
}

// NewCmdTasksList creates the tasks list command
func NewCmdTasksList(f *factory.Factory) *cobra.Command {
	opts := &ListOptions{}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List tasks",
		Aliases: []string{"ls"},
		Example: `  gitscrum tasks list
  gitscrum tasks list --page 2
  gitscrum tasks list --filter blocker
  gitscrum tasks list --assignee @user`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksList(f, opts)
		},
	}

	cmd.Flags().StringVar(&opts.Assignee, "assignee", "", "Filter by assignee (@username)")
	cmd.Flags().StringVar(&opts.Filter, "filter", "", "Filter (blocker, bug, draft, unassigned)")
	cmd.Flags().StringVar(&opts.Workflow, "workflow", "", "Filter by workflow (todo, in-progress, done)")
	cmd.Flags().StringVar(&opts.Type, "type", "", "Filter by type (feature, bug, improvement)")
	cmd.Flags().StringVar(&opts.Board, "board", "", "Filter by board (uuid or name)")
	cmd.Flags().IntVarP(&opts.Limit, "limit", "n", 20, "Maximum tasks to show")
	cmd.Flags().IntVar(&opts.Page, "page", 1, "Page number for pagination")

	return cmd
}

func runTasksList(f *factory.Factory, opts *ListOptions) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	// Cap limit at 50
	if opts.Limit > 50 {
		opts.Limit = 50
	}

	sp := spinner.New("Loading tasks...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Build query path with pagination and filters
	path := fmt.Sprintf("/tasks?company_slug=%s&project_slug=%s&page=%d&per_page=%d", workspace, project, opts.Page, opts.Limit)

	// Resolve workflow filter: fetch workflows from API, find matching ID by title
	if opts.Workflow != "" {
		workflowsResp, err := client.Get(fmt.Sprintf("/projects-workflows/?company_slug=%s&project_slug=%s", workspace, project))
		if err == nil {
			defer workflowsResp.Body.Close()
			var wfResult struct {
				Data []struct {
					ID    int    `json:"id"`
					Title string `json:"title"`
				} `json:"data"`
			}
			if err := api.DecodeResponse(workflowsResp, &wfResult); err == nil {
				workflowFilter := strings.ToLower(opts.Workflow)
				var matchedIDs []string
				for _, wf := range wfResult.Data {
					if strings.ToLower(wf.Title) == workflowFilter || strings.Contains(strings.ToLower(wf.Title), workflowFilter) {
						matchedIDs = append(matchedIDs, fmt.Sprintf("%d", wf.ID))
					}
				}
				if len(matchedIDs) > 0 {
					path += "&workflows=" + strings.Join(matchedIDs, ",")
				}
			}
		}
	}
	// Resolve type filter: fetch types from API, find matching ID by title
	if opts.Type != "" {
		typesResp, err := client.Get(fmt.Sprintf("/project-templates/type?company_slug=%s&project_slug=%s", workspace, project))
		if err == nil {
			defer typesResp.Body.Close()
			var typeResult struct {
				Data []struct {
					ID    int    `json:"id"`
					Title string `json:"title"`
				} `json:"data"`
			}
			if err := api.DecodeResponse(typesResp, &typeResult); err == nil {
				typeFilter := strings.ToLower(opts.Type)
				var matchedIDs []string
				for _, t := range typeResult.Data {
					if strings.ToLower(t.Title) == typeFilter || strings.Contains(strings.ToLower(t.Title), typeFilter) {
						matchedIDs = append(matchedIDs, fmt.Sprintf("%d", t.ID))
					}
				}
				if len(matchedIDs) > 0 {
					path += "&types=" + strings.Join(matchedIDs, ",")
				}
			}
		}
	}

	// Resolve --filter flag: map known keywords to API parameters
	if opts.Filter != "" {
		filterLower := strings.ToLower(opts.Filter)
		switch filterLower {
		case "blocker":
			path += "&is_blocker=1"
		case "bug":
			path += "&is_bug=1"
		case "draft":
			path += "&is_draft=1"
		case "unassigned":
			path += "&unassigned=1"
		default:
			// Treat as a label filter
			path += "&labels=" + opts.Filter
		}
	}

	// Resolve --board flag: fetch boards from API, find matching UUID by label
	if opts.Board != "" {
		boardsResp, err := client.Get(fmt.Sprintf("/project-boards?company_slug=%s&project_slug=%s", workspace, project))
		if err == nil {
			defer boardsResp.Body.Close()
			var boardResult struct {
				Data []struct {
					UUID  string `json:"uuid"`
					Label string `json:"label"`
				} `json:"data"`
			}
			if err := api.DecodeResponse(boardsResp, &boardResult); err == nil {
				boardFilter := strings.ToLower(opts.Board)
				matchedUUID := ""
				for _, b := range boardResult.Data {
					if strings.ToLower(b.Label) == boardFilter || strings.Contains(strings.ToLower(b.Label), boardFilter) {
						matchedUUID = b.UUID
						break
					}
				}
				if matchedUUID != "" {
					path += "&board_uuid=" + matchedUUID
				}
			}
		}
	}

	if opts.Assignee != "" {
		path += "&assignee=" + strings.TrimPrefix(opts.Assignee, "@")
	}

	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Parse response with pagination
	var result struct {
		Data []struct {
			UUID    string `json:"uuid"`
			Number  int    `json:"number"`
			Title   string `json:"title"`
			Code    string `json:"code"`
			RefCode string `json:"ref_code"`
			Workflow *struct {
				Title string `json:"title"`
			} `json:"workflow"`
			Assignees []struct {
				Username string `json:"username"`
			} `json:"users"`
			Project *struct {
				Code string `json:"code"`
			} `json:"project"`
		} `json:"data"`
		Total       int `json:"total"`
		CurrentPage int `json:"current_page"`
		TotalPages  int `json:"total_pages"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	formatter := f.Formatter()

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		project, _ := f.CurrentProject()
		output.EmptyContext(i18n.T("no_tasks"), workspace, project, i18n.T("create_task_hint"))
		return nil
	}

	headers := []string{i18n.T("col_code"), i18n.T("col_title"), "Workflow", "Assignee"}
	rows := make([][]string, 0, len(result.Data))

	for _, task := range result.Data {
		// Prefer ref_code, fallback to code
		code := task.RefCode
		if code == "" {
			code = task.Code
		}
		if code == "" && task.Project != nil {
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

		workflow := ""
		if task.Workflow != nil {
			workflow = task.Workflow.Title
		}

		rows = append(rows, []string{code, title, workflow, assignee})

		if opts.Limit > 0 && len(rows) >= opts.Limit {
			break
		}
	}

	formatter.PrintTable(headers, rows)

	// Show pagination info and navigation hint
	if result.TotalPages > 1 {
		fmt.Println()
		fmt.Println("─────────────────────────────────────")
		fmt.Printf("Page %d of %d (Total: %d tasks)\n", result.CurrentPage, result.TotalPages, result.Total)
		if result.CurrentPage < result.TotalPages {
			fmt.Printf("Next: gitscrum tasks list --page %d\n", result.CurrentPage+1)
		}
		if result.CurrentPage > 1 {
			fmt.Printf("Prev: gitscrum tasks list --page %d\n", result.CurrentPage-1)
		}
	}

	return nil
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

	resp, err := client.Get("/tasks/my-today")
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
		workspace, _ := f.CurrentWorkspace()
		project, _ := f.CurrentProject()
		output.EmptyContext(i18n.T("no_tasks_today"), workspace, project, i18n.T("great_job_no_tasks_today"))
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
		Example: "  gitscrum tasks view a1b2c3d4\n  gitscrum tasks view a1b2c3d4 --web\n  gitscrum tasks view a1b2c3d4 --comments\n  gitscrum tasks view a1b2c3d4 --timers",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if web {
				return runTasksViewWeb(f, args[0])
			}
			showComments, _ := cmd.Flags().GetBool("comments")
			showTimers, _ := cmd.Flags().GetBool("timers")
			return runTasksView(f, args[0], showComments, showTimers)
		},
	}

	cmd.Flags().BoolVarP(&web, "web", "w", false, "Open in browser")
	cmd.Flags().Bool("comments", false, "Show task comments")
	cmd.Flags().Bool("timers", false, "Show time entries")

	return cmd
}

func runTasksView(f *factory.Factory, code string, showComments bool, showTimeTracking bool) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading task...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Parse code to validate format
	// ref_code is an 8-char hex identifier (e.g., a1b2c3d4)
	isRefCode := len(code) == 8 && regexp.MustCompile(`^[a-fA-F0-9]{8}$`).MatchString(code)
	isTaskCode := strings.Contains(code, "-")
	
	if !isRefCode && !isTaskCode {
		sp.Stop()
		return fmt.Errorf("invalid task code format: %s (expected: XX-123 or 8-char ref_code like a1b2c3d4)", code)
	}

	// Use different endpoint based on format
	var endpoint string
	if isRefCode {
		endpoint = fmt.Sprintf("/tasks/ref/%s", code)
	} else {
		endpoint = fmt.Sprintf("/tasks/by-code/%s", code)
	}

	resp, err := client.Get(endpoint)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			UUID        string `json:"uuid"`
			Code        string `json:"code"`
			RefCode     string `json:"ref_code"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Slug        string `json:"slug"`
			Estimative  int    `json:"estimative"`
			EstimatedMinutes     int `json:"estimated_minutes"`
			TotalTrackedMinutes  int `json:"total_tracked_minutes"`
			Workflow    *struct {
				Title string `json:"title"`
				Color string `json:"color"`
			} `json:"workflow"`
			Type *struct {
				Title string `json:"title"`
			} `json:"type"`
			Effort *struct {
				Title string `json:"title"`
				Value int    `json:"value"`
			} `json:"effort"`
			StartDate *api.DateResource `json:"start_date"`
			DueDate   *api.DateResource `json:"due_date"`
			CompletedDate *api.DateResource `json:"completed_date"`
			CreatedAt *api.DateResource `json:"created_at"`
			User *struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"user"`
			Assignees []struct {
				Name     string `json:"name"`
				Username string `json:"username"`
			} `json:"users"`
			Labels []struct {
				Title string `json:"title"`
				Color string `json:"color"`
			} `json:"labels"`
			Company *struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"company"`
			Project *struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"project"`
			Sprint *struct {
				Title   string `json:"title"`
				Code    string `json:"code"`
				RefCode string `json:"ref_code"`
			} `json:"sprint"`
			Board *struct {
				Label string `json:"label"`
				UUID  string `json:"uuid"`
			} `json:"board"`
			UserStory *struct {
				Title string `json:"title"`
				Code  string `json:"code"`
			} `json:"user_story"`
			Settings *struct {
				IsBlocker  bool `json:"is_blocker"`
				IsBug      bool `json:"is_bug"`
				IsDraft    bool `json:"is_draft"`
				IsArchived bool `json:"is_archived"`
			} `json:"settings"`
			Stats *struct {
				Comments            int `json:"comments"`
				Checklists          int `json:"checklists"`
				Attachments         int `json:"attachments"`
				Subtasks            int `json:"subtasks"`
				ChecklistPercentage int `json:"checklist_percentage"`
			} `json:"stats"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	task := result.Data

	// Header
	taskCode := task.Code
	if taskCode == "" {
		taskCode = task.RefCode
	}
	fmt.Printf("\n%s: %s\n", taskCode, task.Title)
	fmt.Println(strings.Repeat("─", 60))

	// If --comments flag is passed, only show comments and return
	if showComments {
		fmt.Println()
		
		// Fetch comments using ref_code or uuid
		commentableId := task.RefCode
		if commentableId == "" {
			commentableId = task.UUID
		}
		
		// Build comments URL with required slugs
		commentsURL := fmt.Sprintf("/comments?commentable_type=issues&commentable_id=%s&company_slug=%s&project_slug=%s", commentableId, workspace, project)
		
		commentsResp, err := client.Get(commentsURL)
		if err != nil {
			fmt.Printf("Could not load comments: %v\n", err)
			fmt.Println()
			return nil
		}
		defer commentsResp.Body.Close()
		
		var commentsResult struct {
			Data []struct {
				Comment   string `json:"comment"`
				CreatedAt *api.DateResource `json:"created_at"`
				User      *struct {
					Name     string `json:"name"`
					Username string `json:"username"`
				} `json:"user"`
			} `json:"data"`
		}
		
		if err := api.DecodeResponse(commentsResp, &commentsResult); err != nil {
			fmt.Printf("%s: %v\n", i18n.T("could_not_parse_comments"), err)
			fmt.Println()
			return nil
		}
		
		if len(commentsResult.Data) == 0 {
			fmt.Println(i18n.T("no_comments_yet"))
			fmt.Println()
			fmt.Println(i18n.Tf("add_comment_hint", map[string]interface{}{"Code": taskCode}))
		} else {
			for _, c := range commentsResult.Data {
				author := i18n.T("unknown")
				if c.User != nil && c.User.Username != "" {
					author = "@" + c.User.Username
				}
				datetime := ""
				if c.CreatedAt != nil {
					datetime = c.CreatedAt.DateTime()
				}
				fmt.Printf("%s (%s):\n", author, datetime)
				fmt.Printf("  %s\n\n", output.StripHTML(c.Comment))
			}
		}
		
		fmt.Println()
		return nil
	}

	// If --timetracking flag is passed, only show time entries and return
	if showTimeTracking {
		fmt.Println()
		
		// Fetch time tracking entries using the same endpoint the app uses
		timeURL := fmt.Sprintf("/time-trackings/no-task/?company_slug=%s&project_slug=%s&task_uuid=%s", workspace, project, task.UUID)
		
		timeResp, err := client.Get(timeURL)
		if err != nil {
			fmt.Printf("%s: %v\n", i18n.T("could_not_load_time_tracking"), err)
			fmt.Println()
			return nil
		}
		defer timeResp.Body.Close()
		
		var timeResult struct {
			Data []struct {
				Comment string `json:"comment"`
				Time    *struct {
					Total           string `json:"total"`
					DurationMinutes int    `json:"duration_minutes"`
					Start           *api.DateResource `json:"start"`
					End             *api.DateResource `json:"end"`
				} `json:"time"`
				User *struct {
					Name     string `json:"name"`
					Username string `json:"username"`
				} `json:"user"`
				IsBillable bool `json:"is_billable"`
			} `json:"data"`
			Stats *struct {
				TotalMinutes int    `json:"total_minutes"`
				Total        string `json:"total"`
			} `json:"stats"`
		}
		
		if err := api.DecodeResponse(timeResp, &timeResult); err != nil {
			fmt.Printf("%s: %v\n", i18n.T("could_not_parse_time_tracking"), err)
			fmt.Println()
			return nil
		}
		
		if len(timeResult.Data) == 0 {
			fmt.Println(i18n.T("no_time_entries_yet"))
			fmt.Println()
			fmt.Println(i18n.Tf("start_tracking_hint", map[string]interface{}{"Code": taskCode}))
		} else {
			// Show total first
			if timeResult.Stats != nil && timeResult.Stats.Total != "" {
				fmt.Printf("%s: %s\n\n", i18n.T("total"), timeResult.Stats.Total)
			}
			
			for _, t := range timeResult.Data {
				user := i18n.T("unknown")
				if t.User != nil && t.User.Username != "" {
					user = "@" + t.User.Username
				}
				
				startTime := ""
				if t.Time != nil && t.Time.Start != nil {
					startTime = t.Time.Start.DateTime()
				}
				
				duration := ""
				if t.Time != nil && t.Time.Total != "" {
					duration = t.Time.Total
				}
				
				billable := ""
				if t.IsBillable {
					billable = " [" + i18n.T("billable") + "]"
				}
				
				fmt.Printf("%s (%s) - %s%s\n", user, startTime, duration, billable)
				if t.Comment != "" {
					fmt.Printf("  %s\n", t.Comment)
				}
				fmt.Println()
			}
		}
		
		return nil
	}
	// Status with badges
	statusLine := ""
	if task.Workflow != nil {
		statusLine = task.Workflow.Title
	}
	if task.Settings != nil {
		if task.Settings.IsBlocker {
			statusLine += " [BLOCKER]"
		}
		if task.Settings.IsBug {
			statusLine += " [BUG]"
		}
		if task.Settings.IsDraft {
			statusLine += " [DRAFT]"
		}
	}
	if statusLine != "" {
		fmt.Printf("Status:      %s\n", statusLine)
	}

	// Type and Effort
	if task.Type != nil && task.Type.Title != "" {
		fmt.Printf("Type:        %s\n", task.Type.Title)
	}
	if task.Effort != nil && task.Effort.Title != "" {
		fmt.Printf("Effort:      %s\n", task.Effort.Title)
	}

	// Progress
	if task.Estimative > 0 {
		fmt.Printf("Progress:    %d%%\n", task.Estimative)
	}

	// Time tracking
	if task.EstimatedMinutes > 0 || task.TotalTrackedMinutes > 0 {
		estimated := fmt.Sprintf("%dh%dm", task.EstimatedMinutes/60, task.EstimatedMinutes%60)
		tracked := fmt.Sprintf("%dh%dm", task.TotalTrackedMinutes/60, task.TotalTrackedMinutes%60)
		fmt.Printf("Time:        %s tracked / %s estimated\n", tracked, estimated)
	}

	// Context: Project, Sprint, Board, User Story
	fmt.Println()
	if task.Company != nil && task.Company.Name != "" {
		fmt.Printf("Workspace:   %s\n", task.Company.Name)
	}
	if task.Project != nil && task.Project.Name != "" {
		fmt.Printf("Project:     %s\n", task.Project.Name)
	}
	if task.Sprint != nil && task.Sprint.Title != "" {
		fmt.Printf("Sprint:      %s\n", task.Sprint.Title)
	}
	if task.Board != nil && task.Board.Label != "" {
		fmt.Printf("Board:       %s\n", task.Board.Label)
	}
	if task.UserStory != nil && task.UserStory.Title != "" {
		fmt.Printf("User Story:  %s\n", task.UserStory.Title)
	}

	// Dates
	fmt.Println()
	if task.StartDate != nil && task.StartDate.ISODate() != "" {
		fmt.Printf("Start:       %s\n", task.StartDate.ISODate())
	}
	if task.DueDate != nil && task.DueDate.ISODate() != "" {
		fmt.Printf("Due:         %s\n", task.DueDate.ISODate())
	}
	if task.CompletedDate != nil && task.CompletedDate.ISODate() != "" {
		fmt.Printf("Completed:   %s\n", task.CompletedDate.ISODate())
	}
	if task.CreatedAt != nil && task.CreatedAt.ISODate() != "" {
		fmt.Printf("Created:     %s\n", task.CreatedAt.ISODate())
	}

	// People
	fmt.Println()
	if task.User != nil && task.User.Username != "" {
		fmt.Printf("Creator:     @%s\n", task.User.Username)
	}
	if len(task.Assignees) > 0 {
		assignees := make([]string, len(task.Assignees))
		for i, a := range task.Assignees {
			assignees[i] = "@" + a.Username
		}
		fmt.Printf("Assigned:    %s\n", strings.Join(assignees, ", "))
	}

	// Labels
	if len(task.Labels) > 0 {
		labels := make([]string, len(task.Labels))
		for i, l := range task.Labels {
			labels[i] = l.Title
		}
		fmt.Printf("Labels:      %s\n", strings.Join(labels, ", "))
	}

	// Stats
	if task.Stats != nil {
		stats := []string{}
		if task.Stats.Subtasks > 0 {
			stats = append(stats, fmt.Sprintf("%d subtasks", task.Stats.Subtasks))
		}
		if task.Stats.Checklists > 0 {
			checklistInfo := fmt.Sprintf("%d checklists", task.Stats.Checklists)
			if task.Stats.ChecklistPercentage > 0 {
				checklistInfo += fmt.Sprintf(" (%d%%)", task.Stats.ChecklistPercentage)
			}
			stats = append(stats, checklistInfo)
		}
		if task.Stats.Comments > 0 {
			stats = append(stats, fmt.Sprintf("%d comments", task.Stats.Comments))
		}
		if task.Stats.Attachments > 0 {
			stats = append(stats, fmt.Sprintf("%d attachments", task.Stats.Attachments))
		}
		if len(stats) > 0 {
			fmt.Printf("Stats:       %s\n", strings.Join(stats, ", "))
		}
	}

	// Description
	if task.Description != "" {
		fmt.Println()
		fmt.Println("Description:")
		fmt.Println(output.StripHTML(task.Description))
	}

	// Shortcuts
	fmt.Println()
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Comments:    gitscrum tasks comment %s\n", taskCode)
	fmt.Printf("Timers:      gitscrum tasks view %s --timers\n", taskCode)
	fmt.Printf("Start timer: gitscrum timer start %s\n", taskCode)
	fmt.Printf("Open web:    gitscrum tasks view %s --web\n", taskCode)
	fmt.Println()

	return nil
}

func runTasksViewWeb(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	// Build public task URL: /{company}/task/{project}/{code}
	url := fmt.Sprintf("https://studio.gitscrum.com/%s/task/%s/%s", workspace, project, code)

	fmt.Printf("Opening %s in browser...\n", url)
	return browser.OpenURL(url)
}

// NewCmdTasksCreate creates the tasks create command
func NewCmdTasksCreate(f *factory.Factory) *cobra.Command {
	var description string
	var assignee string
	var taskType string
	var priority string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new task",
		Example: `  gitscrum tasks create "Fix authentication bug"
  gitscrum tasks create "New feature" -a @john --type feature`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksCreate(f, args[0], description, assignee, taskType, priority)
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Task description")
	cmd.Flags().StringVarP(&assignee, "assignee", "a", "", "Assign to user (@username)")
	cmd.Flags().StringVarP(&taskType, "type", "t", "", "Task type (bug, feature, task)")
	cmd.Flags().StringVar(&priority, "priority", "", "Priority (low, medium, high, critical)")

	return cmd
}

func runTasksCreate(f *factory.Factory, title, description, assignee, taskType, priority string) error {
	if err := f.RequireAuth(); err != nil {
		return err
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

	resp, err := client.Post("/tasks", body)
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
		Example: `  gitscrum tasks update a1b2c3d4 --column="In Progress"
  gitscrum tasks update a1b2c3d4 --title="New title"`,
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

	resp, err := client.Patch(fmt.Sprintf("/tasks/by-code/%s", code), body)
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
		Example: "  gitscrum tasks complete a1b2c3d4",
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

	resp, err := client.Patch(fmt.Sprintf("/tasks/by-code/%s", code), body)
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
		Example: "  gitscrum tasks assign a1b2c3d4 @john",
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

	resp, err := client.Post(fmt.Sprintf("/tasks/by-code/%s/assign", code), body)
	sp.Stop()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Printf("Assigned %s to @%s\n", code, username)

	return nil
}

// NewCmdTasksMove creates the tasks move command
func NewCmdTasksMove(f *factory.Factory) *cobra.Command {
	var toProject string

	cmd := &cobra.Command{
		Use:   "move <code>",
		Short: "Move task to another project",
		Example: `  gitscrum tasks move a1b2c3d4 --to=other-project
  gitscrum tasks move a1b2c3d4 --to="My New Project"`,
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
		return err
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
		Example: `  gitscrum tasks duplicate a1b2c3d4
  gitscrum tasks duplicate a1b2c3d4 --to=other-project
  gitscrum tasks duplicate a1b2c3d4 --with-subtasks`,
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
		return err
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
		Example: `  gitscrum tasks subtasks a1b2c3d4`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTasksSubtasks(f, args[0])
		},
	}
}

func runTasksSubtasks(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
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
		status := output.StatusIcon(t.Status.Title)
		assignee := ""
		if t.Assignee.Name != "" {
			assignee = fmt.Sprintf(" @%s", t.Assignee.Name)
		}
		fmt.Printf("  %s [%s] %s%s\n", status, t.Code, t.Title, assignee)
	}

	return nil
}

// NewCmdTasksComment creates the tasks comment command
func NewCmdTasksComment(f *factory.Factory) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "comment <code>",
		Short: "View or add comments to a task",
		Example: `  # List comments
  gitscrum tasks comment a8c25f3d

  # Add a comment
  gitscrum tasks comment a8c25f3d -m "Fixed the bug"
  gitscrum tasks comment a1b2c3d4 -m "Ready for review"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if message != "" {
				return runTasksAddComment(f, args[0], message)
			}
			return runTasksListComments(f, args[0])
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Comment message to add")

	return cmd
}

func runTasksListComments(f *factory.Factory, code string) error {
	// Reuse the same logic from tasks view --comments
	return runTasksView(f, code, true, false)
}

func runTasksAddComment(f *factory.Factory, code, message string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Adding comment...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// First, get the task to retrieve UUID
	isRefCode := len(code) == 8 && regexp.MustCompile(`^[a-fA-F0-9]{8}$`).MatchString(code)
	
	var endpoint string
	if isRefCode {
		endpoint = fmt.Sprintf("/tasks/ref/%s", code)
	} else {
		endpoint = fmt.Sprintf("/tasks/%s", strings.ToLower(code))
	}

	resp, err := client.Get(endpoint)
	if err != nil {
		sp.Stop()
		return fmt.Errorf("failed to fetch task: %w", err)
	}
	defer resp.Body.Close()

	var taskResult struct {
		Data struct {
			UUID    string `json:"uuid"`
			RefCode string `json:"ref_code"`
			Code    string `json:"code"`
			Title   string `json:"title"`
			Project *struct {
				Slug string `json:"slug"`
			} `json:"project"`
			Company *struct {
				Slug string `json:"slug"`
			} `json:"company"`
		} `json:"data"`
	}

	if err := api.DecodeResponse(resp, &taskResult); err != nil {
		sp.Stop()
		return err
	}

	task := taskResult.Data

	// Build comment payload
	payload := map[string]interface{}{
		"commentable_id":   task.UUID,
		"commentable_type": "issues",
		"comment_text":     message,
	}

	// Add project context if available
	queryParams := ""
	if task.Company != nil && task.Project != nil {
		queryParams = fmt.Sprintf("?company_slug=%s&project_slug=%s", task.Company.Slug, task.Project.Slug)
	}

	commentResp, err := client.Post("/comments"+queryParams, payload)
	if err != nil {
		sp.Stop()
		return fmt.Errorf("failed to add comment: %w", err)
	}
	defer commentResp.Body.Close()

	sp.Stop()

	// Display the task code
	taskCode := task.Code
	if taskCode == "" {
		taskCode = task.RefCode
	}

	fmt.Printf("\n✓ Comment added to %s\n", taskCode)
	fmt.Printf("  \"%s\"\n\n", message)

	return nil
}
