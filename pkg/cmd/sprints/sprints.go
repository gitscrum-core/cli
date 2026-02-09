// Package sprints provides sprint commands for GitScrum CLI
package sprints

import (
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdSprints creates the sprints command group
func NewCmdSprints(f *factory.Factory) *cobra.Command {
	var page int

	cmd := &cobra.Command{
		Use:   "sprints [command]",
		Short: "Manage sprints",
		Long: `View and manage sprints in GitScrum.

Sprints are time-boxed iterations for organizing work. View current sprint,
check burndown charts, and track team velocity.

Without a subcommand, lists sprints for the current project.`,
		Example: `  # List all sprints
  gitscrum sprints

  # List sprints (page 2)
  gitscrum sprints --page 2

  # View current active sprint
  gitscrum sprints current

  # View sprint details
  gitscrum sprints view [sprint-slug]

  # View sprint burndown chart
  gitscrum sprints burndown [sprint-slug]

  # View sprint statistics
  gitscrum sprints stats [sprint-slug]

  # Create a new sprint
  gitscrum sprints create -n "Sprint 5" --start 2024-03-01 --end 2024-03-14`,
		Aliases: []string{"sprint", "sp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintsList(f, page)
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "page number for pagination")

	cmd.AddCommand(NewCmdSprintsCurrent(f))
	cmd.AddCommand(NewCmdSprintsView(f))
	cmd.AddCommand(NewCmdSprintsCreate(f))
	cmd.AddCommand(NewCmdSprintsBurndown(f))
	cmd.AddCommand(NewCmdSprintsStats(f))

	return cmd
}

func runSprintsList(f *factory.Factory, page int) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading sprints...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/sprints?company_slug=%s&project_slug=%s&page=%d", workspace, project, page)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data        []Sprint `json:"data"`
		Total       int      `json:"total"`
		Count       int      `json:"count"`
		PerPage     int      `json:"per_page"`
		CurrentPage int      `json:"current_page"`
		TotalPages  int      `json:"total_pages"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		output.EmptyContext(i18n.T("no_sprints"), workspace, project, i18n.T("create_sprint_hint"))
		return nil
	}

	// Build table with slug column for use in subcommands
	formatter := f.Formatter()
	headers := []string{i18n.T("col_code"), i18n.T("col_slug"), i18n.T("col_title"), i18n.T("col_timebox"), i18n.T("col_status"), i18n.T("col_progress")}
	rows := make([][]string, 0, len(result.Data))

	for _, s := range result.Data {
		// Progress from stats
		totalTasks := getIntValue(s.Stats.TotalTasks)
		closedTasks := getIntValue(s.Stats.ClosedTasks)
		percentage := getIntValue(s.Stats.Percentage)
		progress := fmt.Sprintf("%d/%d (%d%%)", closedTasks, totalTasks, percentage)
		if totalTasks == 0 {
			progress = "-"
		}

		rows = append(rows, []string{
			s.Code,
			s.Slug,
			truncate(s.Title, 25),
			s.Timebox,
			s.Status.Title,
			progress,
		})
	}

	formatter.PrintTable(headers, rows)

	// Show pagination info if more than one page
	if result.TotalPages > 1 {
		fmt.Printf("\nPage %d of %d (Total: %d sprints)\n", result.CurrentPage, result.TotalPages, result.Total)
	}

	return nil
}

// truncate truncates a string to max length with ellipsis
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// SprintStatus represents a sprint status from the API
type SprintStatus struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

// DateResource represents a date from the API
type DateResource struct {
	DateForHumans string `json:"date_for_humans"`
	ISO8601       string `json:"iso8601"`
	Timestamp     int64  `json:"timestamp"`
}

// SprintTimeboxDetail represents the timebox object from detail API (contains DateResource objects)
type SprintTimeboxDetail struct {
	Start  *DateResource `json:"start"`
	Finish *DateResource `json:"finish"`
}

// String returns a formatted period string
func (t SprintTimeboxDetail) String() string {
	if t.Start != nil && t.Finish != nil && t.Start.ISO8601 != "" && t.Finish.ISO8601 != "" {
		// Parse and format the dates
		startTime, err1 := time.Parse(time.RFC3339, t.Start.ISO8601)
		finishTime, err2 := time.Parse(time.RFC3339, t.Finish.ISO8601)
		if err1 == nil && err2 == nil {
			return startTime.Format("2006-01-02") + " - " + finishTime.Format("2006-01-02")
		}
	}
	return ""
}

// SprintStats represents sprint statistics (using interface{} for flexible types)
type SprintStats struct {
	WorkedHours interface{} `json:"worked_hours"`
	TotalTasks  interface{} `json:"total_tasks"`
	ClosedTasks interface{} `json:"closed_tasks"`
	Percentage  interface{} `json:"percentage"`
	StoryPoints interface{} `json:"story_points"`
	Comments    interface{} `json:"comments"`
}

// GetInt safely gets an int value from interface{}
func getIntValue(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case int64:
		return int(val)
	case string:
		var result int
		fmt.Sscanf(val, "%d", &result)
		return result
	default:
		return 0
	}
}

// getStringValue safely gets a string value from interface{}
func getStringValue(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == 0 {
			return "0"
		}
		return fmt.Sprintf("%.1f", val)
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	default:
		return ""
	}
}

// Sprint represents a sprint for list endpoints (timebox is a string)
type Sprint struct {
	ID       int          `json:"id"`
	Code     string       `json:"code"`
	Slug     string       `json:"slug"`
	Title    string       `json:"title"`
	Timebox  string       `json:"timebox"`
	Duration int          `json:"duration"`
	Status   SprintStatus `json:"status"`
	Stats    SprintStats  `json:"stats"`
}

// SprintDetail represents a single sprint from detail endpoint (timebox is an object)
type SprintDetail struct {
	ID       int                 `json:"id"`
	Code     string              `json:"code"`
	Slug     string              `json:"slug"`
	Title    string              `json:"title"`
	Timebox  SprintTimeboxDetail `json:"timebox"`
	Duration int                 `json:"duration"`
	Status   SprintStatus        `json:"status"`
	Stats    SprintStats         `json:"stats"`
}

func getSprintStatus(s Sprint) string {
	if s.Status.Title != "" {
		return s.Status.Title
	}
	return "Pending"
}

// NewCmdSprintsCurrent shows the active sprint
func NewCmdSprintsCurrent(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current/active sprint",
		Long:  `Show the currently active sprint with KPIs and progress.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintsCurrent(f)
		},
	}
}

func runSprintsCurrent(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	if _, _, err := f.RequireWorkspaceAndProject(); err != nil {
		return err
	}

	sp := spinner.New("Loading current sprint...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/sprints?current=true"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return fmt.Errorf("no active sprint found")
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	s := result.Data

	output.Header(fmt.Sprintf("Current Sprint: %s", s.Title))

	output.KeyValue("Period", s.Timebox)

	// Calculate days remaining from Timebox (format: "YYYY-MM-DD - YYYY-MM-DD")
	parts := strings.Split(s.Timebox, " - ")
	if len(parts) == 2 {
		end, _ := time.Parse("2006-01-02", parts[1])
		daysLeft := int(time.Until(end).Hours() / 24)
		if daysLeft > 0 {
			output.KeyValuef("Days remaining", "%d", daysLeft)
		} else {
			output.Warning("Sprint ended")
		}
	}

	// Progress
	totalTasks := getIntValue(s.Stats.TotalTasks)
	closedTasks := getIntValue(s.Stats.ClosedTasks)
	storyPoints := getIntValue(s.Stats.StoryPoints)
	if totalTasks > 0 {
		progress := float64(closedTasks) / float64(totalTasks) * 100
		bar := renderProgressBar(progress, 30)
		output.KeyValuef("Tasks", "%s %.0f%% (%d/%d)", bar, progress, closedTasks, totalTasks)
	}

	if storyPoints > 0 {
		output.KeyValuef("Story Points", "%d", storyPoints)
	}

	fmt.Println()
	return nil
}

func renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

// NewCmdSprintsView shows a specific sprint
func NewCmdSprintsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view [sprint-slug]",
		Short: "View sprint details with KPIs",
		Long:  `Display detailed sprint information including KPIs, progress, and stats.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintsView(f, args[0])
		},
	}
}

func runSprintsView(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading sprint...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/sprints/%s?company_slug=%s&project_slug=%s", slug, workspace, project)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data SprintDetail `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	s := result.Data

	// Extract stats
	totalTasks := getIntValue(s.Stats.TotalTasks)
	closedTasks := getIntValue(s.Stats.ClosedTasks)
	percentage := getIntValue(s.Stats.Percentage)
	storyPoints := getIntValue(s.Stats.StoryPoints)
	workedHours := getStringValue(s.Stats.WorkedHours)

	// Fetch tasks for this sprint
	var tasks []SprintTask
	tasksPath := fmt.Sprintf("/tasks/?company_slug=%s&project_slug=%s&sprint_slug=%s&per_page=50", workspace, project, slug)
	tasksResp, err := client.Get(tasksPath)
	if err == nil {
		var tasksResult struct {
			Data []SprintTask `json:"data"`
		}
		if api.DecodeResponse(tasksResp, &tasksResult) == nil {
			tasks = tasksResult.Data
		}
	}

	// === CLEAN GITHUB CLI-STYLE OUTPUT ===
	fmt.Println()
	
	// Header line
	statusColor := pterm.FgYellow
	switch strings.ToLower(s.Status.Title) {
	case "active", "in progress":
		statusColor = pterm.FgGreen
	case "done", "completed":
		statusColor = pterm.FgBlue
	}
	
	fmt.Printf("%s  %s\n", 
		pterm.Bold.Sprintf("%s", s.Title),
		pterm.NewStyle(statusColor).Sprint(strings.ToUpper(s.Status.Title)))
	fmt.Printf("%s → %s  (%dd)\n\n", 
		getDatePart(s.Timebox.Start), 
		getDatePart(s.Timebox.Finish),
		s.Duration)

	// Stats in two columns
	fmt.Printf("Tasks: %s/%s (%d%%)    Points: %s    Worked: %sh\n\n",
		pterm.Green(fmt.Sprintf("%d", closedTasks)),
		fmt.Sprintf("%d", totalTasks),
		percentage,
		pterm.Cyan(fmt.Sprintf("%d", storyPoints)),
		workedHours)

	// Group tasks by workflow state
	backlog := []SprintTask{}
	inProgress := []SprintTask{}
	review := []SprintTask{}
	done := []SprintTask{}
	
	for _, t := range tasks {
		switch t.Workflow.State {
		case 0:
			backlog = append(backlog, t)
		case 2:
			inProgress = append(inProgress, t)
		case 3:
			review = append(review, t)
		case 1:
			done = append(done, t)
		default:
			backlog = append(backlog, t)
		}
	}

	// Print tasks by column - simple list format
	printTaskGroup("TODO", len(backlog), backlog, pterm.LightBlue)
	printTaskGroup("IN PROGRESS", len(inProgress), inProgress, pterm.Yellow)
	printTaskGroup("REVIEW", len(review), review, pterm.Magenta)
	printTaskGroup("DONE", len(done), done, pterm.Green)

	// Hint
	fmt.Printf("\n%s\n", pterm.Gray("Use 'gitscrum tasks view <ref_code>' to see task details"))

	return nil
}

// printTaskGroup prints a group of tasks with a header
func printTaskGroup(title string, count int, tasks []SprintTask, colorFn func(a ...interface{}) string) {
	fmt.Printf("%s %s\n", colorFn(title), pterm.Gray(fmt.Sprintf("(%d)", count)))
	
	if len(tasks) == 0 {
		fmt.Println(pterm.Gray("  (empty)"))
		return
	}
	
	for i, t := range tasks {
		if i >= 5 {
			fmt.Printf("  %s\n", pterm.Gray(fmt.Sprintf("... and %d more", len(tasks)-5)))
			break
		}
		
		// Simple format: #ref_code  title  @owner
		owner := ""
		if len(t.Users) > 0 {
			owner = pterm.Gray("@" + t.Users[0].Username)
		}
		
		title := t.Title
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		
		fmt.Printf("  %s  %s  %s\n", 
			pterm.Cyan(t.GetIdentifier()),
			title,
			owner)
	}
}

// padTwoColumns formats two strings in a left-right layout
func padTwoColumns(left, right string, width int) string {
	leftVisible := pterm.RemoveColorFromString(left)
	rightVisible := pterm.RemoveColorFromString(right)
	padding := width - len(leftVisible) - len(rightVisible)
	if padding < 1 {
		padding = 1
	}
	return left + strings.Repeat(" ", padding) + right
}

// renderProgressBar3 creates a progress bar with Unicode blocks
func renderProgressBar3(percentage int, width int) string {
	filled := (percentage * width) / 100
	if filled < 0 {
		filled = 0
	}
	if filled > width {
		filled = width
	}
	empty := width - filled
	
	bar := pterm.Green(strings.Repeat("█", filled)) + pterm.Gray(strings.Repeat("░", empty))
	return bar
}

// formatKanbanTask formats a single task for Kanban view
func formatKanbanTask(t SprintTask, width int) string {
	owner := "-"
	if len(t.Users) > 0 {
		owner = "@" + truncate(t.Users[0].Username, 8)
	}
	
	sp := getStringValue(t.Effort)
	if sp == "" || sp == "0" {
		sp = "-"
	}
	
	title := t.Title
	maxTitle := width - 50
	if maxTitle < 20 {
		maxTitle = 20
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-3] + "..."
	}
	
	// Format: #CODE  Title...  SP: X  @owner
	// Uses cli_code (8 chars) as primary identifier for CLI
	code := pterm.Cyan("#" + t.GetIdentifier())
	return fmt.Sprintf("%s  %-*s  SP: %-3s  %s", code, maxTitle, title, sp, pterm.Gray(owner))
}

// padToWidth pads a string to the specified width
func padToWidth(s string, width int) string {
	// Calculate visible length (without ANSI codes)
	visibleLen := pterm.RemoveColorFromString(s)
	if len(visibleLen) >= width {
		return s
	}
	padding := width - len(visibleLen)
	return s + strings.Repeat(" ", padding)
}

// SprintTask represents a task in a sprint
type SprintTask struct {
	Code     string `json:"code"`
	RefCode  string `json:"ref_code"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Effort   interface{} `json:"effort"`
	Priority interface{} `json:"priority"`
	Workflow struct {
		Title string `json:"title"`
		State int    `json:"state"`
	} `json:"workflow"`
	Users []struct {
		Username string `json:"username"`
	} `json:"users"`
}

// GetIdentifier returns the best identifier for CLI/API usage
// Priority: ref_code (always 8 chars) > code > slug
func (t SprintTask) GetIdentifier() string {
	if t.RefCode != "" {
		return t.RefCode
	}
	if t.Code != "" {
		return t.Code
	}
	return t.Slug
}

// getDatePart extracts just the date from a DateResource
func getDatePart(d *DateResource) string {
	if d == nil || d.ISO8601 == "" {
		return "-"
	}
	t, err := time.Parse(time.RFC3339, d.ISO8601)
	if err != nil {
		return "-"
	}
	return t.Format("2006-01-02")
}

// renderProgressBar2 creates a progress bar with filled/empty chars
func renderProgressBar2(percentage int, width int) string {
	filled := (percentage * width) / 100
	empty := width - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += pterm.Green("▉")
	}
	for i := 0; i < empty; i++ {
		bar += pterm.Gray("░")
	}
	return bar
}

// renderTasksTable renders the tasks overview table
func renderTasksTable(sprintCode string, tasks []SprintTask) {
	title := fmt.Sprintf("Tasks Overview (%s)", sprintCode)
	
	// Build table data
	tableData := pterm.TableData{
		{"ID", "Title", "Owner", "Status", "Est", "Sp"},
	}
	
	for _, t := range tasks {
		// Get owner
		owner := "-"
		if len(t.Users) > 0 {
			owner = "@" + t.Users[0].Username
		}
		
		// Get status styled
		status := strings.ToUpper(t.Workflow.Title)
		if len(status) > 8 {
			status = status[:8]
		}
		statusStyled := status
		switch t.Workflow.State {
		case 0: // TODO
			statusStyled = pterm.LightBlue(status)
		case 2: // In Progress
			statusStyled = pterm.Yellow(status)
		case 1: // Done
			statusStyled = pterm.Green(status)
		}
		
		// Get effort and story points
		effort := getStringValue(t.Effort)
		if effort == "" || effort == "0" {
			effort = "-"
		} else {
			effort = effort + "h"
		}
		
		// Truncate title
		taskTitle := t.Title
		if len(taskTitle) > 32 {
			taskTitle = taskTitle[:29] + "..."
		}
		
		tableData = append(tableData, []string{
			t.Code,
			taskTitle,
			truncate(owner, 12),
			statusStyled,
			effort,
			"-",
		})
	}
	
	pterm.DefaultBox.
		WithTitle(title).
		WithTitleTopLeft().
		WithBoxStyle(pterm.NewStyle(pterm.FgWhite)).
		Println("")
	
	pterm.DefaultTable.
		WithHasHeader(true).
		WithData(tableData).
		Render()
}

// renderTasksTableWithWidth renders the tasks overview table with consistent width
func renderTasksTableWithWidth(sprintCode string, tasks []SprintTask, boxWidth int) {
	title := fmt.Sprintf("Tasks Overview (%s)", sprintCode)
	
	// Calculate dynamic column widths based on boxWidth
	titleWidth := boxWidth - 50 // Leave room for ID, Owner, Status, Est, Sp
	if titleWidth < 20 {
		titleWidth = 20
	}
	
	// Build table header with proper spacing
	separator := strings.Repeat("─", boxWidth-6)
	header := fmt.Sprintf("%-8s %-*s %-12s %-8s %-4s %-4s", "ID", titleWidth, "Title", "Owner", "Status", "Est", "Sp")
	
	// Build rows
	var rows []string
	for _, t := range tasks {
		// Get owner
		owner := "-"
		if len(t.Users) > 0 {
			owner = "@" + t.Users[0].Username
		}
		owner = truncate(owner, 11)
		
		// Get status styled
		status := t.Workflow.Title
		if len(status) > 8 {
			status = status[:8]
		}
		statusStyled := status
		switch t.Workflow.State {
		case 0: // TODO
			statusStyled = pterm.LightBlue(status)
		case 2: // In Progress
			statusStyled = pterm.Yellow(status)
		case 1: // Done
			statusStyled = pterm.Green(status)
		}
		
		// Get effort
		effort := getStringValue(t.Effort)
		if effort == "" || effort == "0" {
			effort = "-"
		} else {
			effort = effort + "h"
		}
		
		// Truncate title
		taskTitle := t.Title
		if len(taskTitle) > titleWidth {
			taskTitle = taskTitle[:titleWidth-3] + "..."
		}
		
		row := fmt.Sprintf("%-8s %-*s %-12s %-8s %-4s %-4s", 
			t.Code, 
			titleWidth, taskTitle, 
			owner,
			statusStyled,
			effort,
			"-")
		rows = append(rows, padToWidth(row, boxWidth-4))
	}
	
	// Build content
	content := padToWidth(header, boxWidth-4) + "\n" + separator + "\n" + strings.Join(rows, "\n")
	
	pterm.DefaultBox.
		WithTitle(title).
		WithTitleTopLeft().
		WithBoxStyle(pterm.NewStyle(pterm.FgWhite)).
		Println(content)
}

// renderKanbanBoard creates an ASCII art Kanban board visualization
func renderKanbanBoard(todo, inProgress, done, total int) {
	if total == 0 {
		return
	}

	// Calculate percentages and bar widths
	barWidth := 15
	todoWidth := (todo * barWidth) / max(1, total)
	wipWidth := (inProgress * barWidth) / max(1, total)
	doneWidth := (done * barWidth) / max(1, total)

	// Ensure at least 1 char if there are any tasks
	if todo > 0 && todoWidth == 0 {
		todoWidth = 1
	}
	if inProgress > 0 && wipWidth == 0 {
		wipWidth = 1
	}
	if done > 0 && doneWidth == 0 {
		doneWidth = 1
	}

	// Create mini bars
	todoBar := strings.Repeat("█", todoWidth) + strings.Repeat("░", barWidth-todoWidth)
	wipBar := strings.Repeat("█", wipWidth) + strings.Repeat("░", barWidth-wipWidth)
	doneBar := strings.Repeat("█", doneWidth) + strings.Repeat("░", barWidth-doneWidth)

	// Build Kanban columns
	content := fmt.Sprintf(`
    ┌───────────────────┬───────────────────┬───────────────────┐
    │    📋 TO DO       │   🔄 IN PROGRESS  │     ✅ DONE       │
    ├───────────────────┼───────────────────┼───────────────────┤
    │                   │                   │                   │
    │   %s%s%s   │   %s%s%s   │   %s%s%s   │
    │                   │                   │                   │
    │     %s %d tasks     │     %s %d tasks     │     %s %d tasks     │
    │                   │                   │                   │
    └───────────────────┴───────────────────┴───────────────────┘`,
		pterm.Gray(todoBar[:5]), pterm.LightBlue(todoBar[5:10]), pterm.Gray(todoBar[10:]),
		pterm.Gray(wipBar[:5]), pterm.Yellow(wipBar[5:10]), pterm.Gray(wipBar[10:]),
		pterm.Gray(doneBar[:5]), pterm.Green(doneBar[5:10]), pterm.Gray(doneBar[10:]),
		pterm.LightBlue(""), todo, pterm.Yellow(""), inProgress, pterm.Green(""), done,
	)

	pterm.DefaultBox.WithTitle(pterm.Cyan("📊 Kanban Board")).WithTitleTopCenter().WithBoxStyle(pterm.NewStyle(pterm.FgCyan)).Println(content)
}

// getProgressColored returns a colored percentage string
func getProgressColored(percentage int) string {
	pctStr := fmt.Sprintf("%d%%", percentage)
	if percentage >= 80 {
		return pterm.Green(pctStr)
	} else if percentage >= 50 {
		return pterm.Yellow(pctStr)
	} else if percentage >= 25 {
		return pterm.LightYellow(pctStr)
	}
	return pterm.Red(pctStr)
}

// renderVisualProgress creates a visual progress bar
func renderVisualProgress(percentage int, width int) string {
	filled := (percentage * width) / 100
	empty := width - filled
	
	bar := ""
	for i := 0; i < filled; i++ {
		bar += pterm.Green("█")
	}
	for i := 0; i < empty; i++ {
		bar += pterm.Gray("░")
	}
	return bar
}

// getStatusStyled returns a styled status string
func getStatusStyled(status string) string {
	switch strings.ToLower(status) {
	case "active", "in progress":
		return pterm.Green(status)
	case "done", "completed":
		return pterm.Blue(status)
	case "standby", "pending":
		return pterm.Yellow(status)
	default:
		return pterm.Gray(status)
	}
}

// NewCmdSprintsCreate creates a new sprint
func NewCmdSprintsCreate(f *factory.Factory) *cobra.Command {
	var startDate, endDate string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new sprint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintsCreate(f, args[0], startDate, endDate)
		},
	}

	cmd.Flags().StringVar(&startDate, "start", "", "Start date (YYYY-MM-DD)")
	cmd.Flags().StringVar(&endDate, "end", "", "End date (YYYY-MM-DD)")
	cmd.MarkFlagRequired("start")
	cmd.MarkFlagRequired("end")

	return cmd
}

func runSprintsCreate(f *factory.Factory, title, startDate, endDate string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Creating sprint...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"title":      title,
		"date_start": startDate,
		"date_end":   endDate,
	}

	path := "/sprints"
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Successf("Sprint created: %s", result.Data.Title)
	output.KeyValue("Period", result.Data.Timebox)

	return nil
}

// NewCmdSprintsBurndown shows burndown chart
func NewCmdSprintsBurndown(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "burndown [slug]",
		Short: "Show burndown chart (ASCII)",
		Long:  `Display an ASCII burndown chart for the current or specified sprint.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runSprintsBurndown(f, slug)
		},
	}
}

func runSprintsBurndown(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading burndown data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Get burndown data
	path := "/sprints"
	if slug != "" {
		path += "/" + slug
	} else {
		path += "/current"
	}
	path += "/burndown"

	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Days []struct {
				Date   string `json:"date"`
				Ideal  int    `json:"ideal"`
				Actual int    `json:"actual"`
			} `json:"days"`
			TotalPoints int `json:"total_points"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return fmt.Errorf("failed to get burndown data: %w", err)
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Sprint Burndown")

	maxPoints := result.Data.TotalPoints
	if maxPoints == 0 {
		output.Empty("No points data available", "")
		return nil
	}

	for _, day := range result.Data.Days {
		idealBar := int(float64(day.Ideal) / float64(maxPoints) * 40)
		actualBar := int(float64(day.Actual) / float64(maxPoints) * 40)

		fmt.Printf("  %s │ %s\n", day.Date[5:10], strings.Repeat("░", idealBar))
		fmt.Printf("         │ %s %d pts\n", strings.Repeat("█", actualBar), day.Actual)
	}

	fmt.Println()
	return nil
}

// NewCmdSprintsStats shows sprint statistics
func NewCmdSprintsStats(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [slug]",
		Short: "Show sprint statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runSprintsStats(f, slug)
		},
	}
}

func runSprintsStats(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	output.Header("Sprint Statistics")
	output.Info("Coming soon: velocity, cycle time, scope changes...")
	fmt.Println()
	output.Dim("Use 'gitscrum sprints current' for basic KPIs")

	return nil
}
