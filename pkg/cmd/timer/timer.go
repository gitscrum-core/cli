// Package timer provides time tracking commands for GitScrum CLI
package timer

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdTimer creates the timer command group
func NewCmdTimer(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "timer [command]",
		Short: "Track time on tasks",
		Long: `Manage time tracking for tasks.

Track your work time directly from the command line. Start, stop, and log
time entries against tasks with automatic elapsed time calculation.

Without a subcommand, shows the currently active timer.`,
		Example: `  # Show active timer
  gitscrum timer

  # Start timer on a task
  gitscrum timer start GS-123

  # Start with a note
  gitscrum timer start GS-123 -m "Working on bug fix"

  # Stop current timer
  gitscrum timer stop

  # View today's time log
  gitscrum timer log --today

  # View weekly productivity report
  gitscrum timer productivity --week`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerStatus(f)
		},
	}

	cmd.AddCommand(NewCmdTimerStart(f))
	cmd.AddCommand(NewCmdTimerStop(f))
	cmd.AddCommand(NewCmdTimerLog(f))
	cmd.AddCommand(NewCmdTimerReport(f))
	cmd.AddCommand(NewCmdTimerProductivity(f))

	return cmd
}

// ActiveTimer represents an active time tracking entry
type ActiveTimer struct {
	UUID    string `json:"uuid"`
	Start   string `json:"start"`
	Comment string `json:"comment"`
	Issue   struct {
		UUID   string `json:"uuid"`
		Number int    `json:"number"`
		Title  string `json:"title"`
	} `json:"issue"`
}

func runTimerStatus(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required. Use -w flag or set default")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/time-trackings/active?company_slug=%s", workspace)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data []ActiveTimer `json:"data"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No active timer")
		fmt.Println()
		fmt.Println("Start one with: gitscrum timer start <task-code>")
		return nil
	}

	fmt.Println("ACTIVE TIMERS")
	fmt.Println()

	for _, t := range result.Data {
		start, _ := time.Parse(time.RFC3339, t.Start)
		elapsed := time.Since(start).Round(time.Minute)

		code := fmt.Sprintf("#%d", t.Issue.Number)
		fmt.Printf("* %s  %s\n", code, t.Issue.Title)
		fmt.Printf("  elapsed: %s\n", formatDuration(elapsed))
		if t.Comment != "" {
			fmt.Printf("  note: %s\n", t.Comment)
		}
		fmt.Println()
	}

	fmt.Println("Stop with: gitscrum timer stop")
	return nil
}

// NewCmdTimerStart creates the timer start command
func NewCmdTimerStart(f *factory.Factory) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "start <code>",
		Short: "Start timer for a task",
		Example: `  gitscrum timer start GS-123
  gitscrum timer start GS-123 -m "Working on authentication"`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerStart(f, args[0], message)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Description of work")

	return cmd
}

func runTimerStart(f *factory.Factory, code, message string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// First, get the task by code to get its UUID
	taskResp, err := client.Get(fmt.Sprintf("/time-trackings/resolve-task/%s", code))
	if err != nil {
		return fmt.Errorf("task not found: %s", code)
	}
	defer taskResp.Body.Close()

	if taskResp.StatusCode == 404 {
		return fmt.Errorf("task not found: %s", code)
	}

	var taskResult struct {
		Data struct {
			UUID string `json:"uuid"`
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := decodeResponse(taskResp, &taskResult); err != nil {
		return fmt.Errorf("task not found: %s", code)
	}

	if taskResult.Data.UUID == "" {
		return fmt.Errorf("task not found: %s", code)
	}

	body := map[string]interface{}{
		"task_uuid": taskResult.Data.UUID,
		"comment":   message,
	}

	resp, err := client.Post("/time-trackings/start", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 409 {
		fmt.Printf("warning: timer already running for %s\n", code)
		return nil
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to start timer: HTTP %d", resp.StatusCode)
	}

	displayCode := taskResult.Data.Code
	if displayCode == "" {
		displayCode = code
	}

	fmt.Printf("Timer started for %s\n", displayCode)
	if message != "" {
		fmt.Printf("  note: %s\n", message)
	}
	fmt.Printf("  started at: %s\n", time.Now().Format("15:04"))

	return nil
}

// NewCmdTimerStop creates the timer stop command
func NewCmdTimerStop(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "stop",
		Short:   "Stop all active timers",
		Long:    "Stop all your active timers in the current workspace.",
		Example: "  gitscrum timer stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerStop(f)
		},
	}
}

func runTimerStop(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"company_slug": workspace,
	}

	resp, err := client.Post("/time-trackings/stop-all", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Message string `json:"message"`
		Stopped int    `json:"stopped"`
		Data    []struct {
			TaskCode          string `json:"task_code"`
			TaskTitle         string `json:"task_title"`
			DurationFormatted string `json:"duration_formatted"`
		} `json:"data"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return err
	}

	if result.Stopped == 0 {
		fmt.Println("No active timers to stop")
		return nil
	}

	fmt.Printf("Stopped %d timer(s)\n\n", result.Stopped)

	for _, t := range result.Data {
		fmt.Printf("  %s  %s\n", t.TaskCode, t.TaskTitle)
		fmt.Printf("    duration: %s\n", t.DurationFormatted)
	}

	return nil
}

// NewCmdTimerLog creates the timer log command
func NewCmdTimerLog(f *factory.Factory) *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "log <code> <duration>",
		Short: "Log time manually",
		Long: `Log time manually without using the timer.

Duration format: 1h30m, 2h, 45m`,
		Example: `  gitscrum timer log GS-123 2h30m
  gitscrum timer log GS-123 45m -m "Code review"`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerLog(f, args[0], args[1], message)
		},
	}

	cmd.Flags().StringVarP(&message, "message", "m", "", "Description of work")

	return cmd
}

func runTimerLog(f *factory.Factory, code, duration, message string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	d, err := time.ParseDuration(duration)
	if err != nil {
		return fmt.Errorf("invalid duration format: %s (use format like 1h30m)", duration)
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	taskResp, err := client.Get(fmt.Sprintf("/time-trackings/resolve-task/%s", code))
	if err != nil {
		return fmt.Errorf("task not found: %s", code)
	}
	defer taskResp.Body.Close()

	if taskResp.StatusCode == 404 {
		return fmt.Errorf("task not found: %s", code)
	}

	var taskResult struct {
		Data struct {
			UUID string `json:"uuid"`
			Code string `json:"code"`
		} `json:"data"`
	}
	if err := decodeResponse(taskResp, &taskResult); err != nil {
		return fmt.Errorf("task not found: %s", code)
	}

	if taskResult.Data.UUID == "" {
		return fmt.Errorf("task not found: %s", code)
	}

	end := time.Now()
	start := end.Add(-d)

	body := map[string]interface{}{
		"task_uuid": taskResult.Data.UUID,
		"start":     start.Format(time.RFC3339),
		"end":       end.Format(time.RFC3339),
		"comment":   message,
	}

	resp, err := client.Post("/time-trackings", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to log time: HTTP %d", resp.StatusCode)
	}

	displayCode := taskResult.Data.Code
	if displayCode == "" {
		displayCode = code
	}

	fmt.Printf("Logged %s for %s\n", formatDuration(d), displayCode)
	if message != "" {
		fmt.Printf("  note: %s\n", message)
	}

	return nil
}

// NewCmdTimerReport creates the timer report command
func NewCmdTimerReport(f *factory.Factory) *cobra.Command {
	var week bool
	var team bool

	cmd := &cobra.Command{
		Use:   "report",
		Short: "Show time tracking report",
		Example: `  gitscrum timer report
  gitscrum timer report --week
  gitscrum timer report --team`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerReport(f, week, team)
		},
	}

	cmd.Flags().BoolVar(&week, "week", false, "Show weekly report")
	cmd.Flags().BoolVar(&team, "team", false, "Show team report")

	return cmd
}

func runTimerReport(f *factory.Factory, week, team bool) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	period := "today"
	periodLabel := "Today"
	if week {
		period = "last-7-days"
		periodLabel = "This Week"
	}

	path := fmt.Sprintf("/time-trackings/reports?company_slug=%s&period=%s", workspace, period)
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalSeconds int `json:"total_seconds"`
			Entries      []struct {
				TaskCode    string `json:"task_code"`
				TaskTitle   string `json:"task_title"`
				Seconds     int    `json:"seconds"`
				Description string `json:"description"`
				User        struct {
					Name string `json:"name"`
				} `json:"user"`
			} `json:"entries"`
			ByUser []struct {
				Name    string `json:"name"`
				Seconds int    `json:"seconds"`
			} `json:"by_user,omitempty"`
		} `json:"data"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return err
	}

	if team {
		fmt.Printf("TIME REPORT - %s (Team)\n", periodLabel)
	} else {
		fmt.Printf("TIME REPORT - %s\n", periodLabel)
	}
	fmt.Println()

	totalMinutes := result.Data.TotalSeconds / 60
	totalHours := totalMinutes / 60
	totalMins := totalMinutes % 60
	fmt.Printf("Total: %dh %dm\n\n", totalHours, totalMins)

	if team && len(result.Data.ByUser) > 0 {
		fmt.Println("BY MEMBER:")
		for _, u := range result.Data.ByUser {
			mins := u.Seconds / 60
			h := mins / 60
			m := mins % 60
			fmt.Printf("  %-20s %dh %dm\n", u.Name, h, m)
		}
		fmt.Println()
	}

	if len(result.Data.Entries) > 0 {
		fmt.Println("ENTRIES:")
		for _, e := range result.Data.Entries {
			mins := e.Seconds / 60
			h := mins / 60
			m := mins % 60
			desc := e.Description
			if len(desc) > 30 {
				desc = desc[:30] + "..."
			}
			fmt.Printf("  %s  %dh %dm  %s\n", e.TaskCode, h, m, desc)
		}
	} else {
		fmt.Println("No time entries for this period")
	}

	return nil
}

// NewCmdTimerProductivity creates the productivity command
func NewCmdTimerProductivity(f *factory.Factory) *cobra.Command {
	var period string

	cmd := &cobra.Command{
		Use:   "productivity",
		Short: "Show productivity metrics",
		Long: `Display productivity metrics and insights.

Shows time distribution, focus time, and productivity scores.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTimerProductivity(f, period)
		},
	}

	cmd.Flags().StringVar(&period, "period", "last-7-days", "Time period: today, last-7-days, last-30-days")

	return cmd
}

func runTimerProductivity(f *factory.Factory, period string) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/time-trackings/productivity?company_slug=%s&period=%s", workspace, period)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			TotalSeconds      int     `json:"total_seconds"`
			FocusSeconds      int     `json:"focus_seconds"`
			ProductivityScore float64 `json:"productivity_score"`
			AvgDailyHours     float64 `json:"avg_daily_hours"`
			MostProductiveDay string  `json:"most_productive_day"`
			TasksCompleted    int     `json:"tasks_completed"`
			ByCategory        []struct {
				Name    string  `json:"name"`
				Seconds int     `json:"seconds"`
				Percent float64 `json:"percent"`
			} `json:"by_category"`
		} `json:"data"`
	}

	if err := decodeResponse(resp, &result); err != nil {
		return err
	}

	d := result.Data

	fmt.Println("PRODUCTIVITY METRICS")
	fmt.Println("====================")
	fmt.Println()

	totalMins := d.TotalSeconds / 60
	totalH := totalMins / 60
	totalM := totalMins % 60
	focusMins := d.FocusSeconds / 60
	focusH := focusMins / 60
	focusM := focusMins % 60

	fmt.Printf("Total time:       %dh %dm\n", totalH, totalM)
	fmt.Printf("Focus time:       %dh %dm\n", focusH, focusM)
	fmt.Printf("Avg daily:        %.1fh\n", d.AvgDailyHours)
	fmt.Printf("Tasks completed:  %d\n", d.TasksCompleted)
	fmt.Println()

	scoreBar := renderProgressBar(d.ProductivityScore, 100, 20)
	fmt.Printf("Score: %.0f%% %s\n", d.ProductivityScore, scoreBar)
	fmt.Println()

	if d.MostProductiveDay != "" {
		fmt.Printf("Most productive: %s\n", d.MostProductiveDay)
		fmt.Println()
	}

	if len(d.ByCategory) > 0 {
		fmt.Println("TIME BY CATEGORY:")
		for _, c := range d.ByCategory {
			bar := renderProgressBar(c.Percent, 100, 15)
			mins := c.Seconds / 60
			h := mins / 60
			m := mins % 60
			fmt.Printf("  %-15s %s %dh %dm (%.0f%%)\n", c.Name, bar, h, m, c.Percent)
		}
	}

	return nil
}

// Helper functions

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

func renderProgressBar(value, max float64, width int) string {
	if max == 0 {
		max = 100
	}
	percent := value / max
	if percent > 1 {
		percent = 1
	}
	filled := int(percent * float64(width))
	empty := width - filled
	return "[" + repeatChar('#', filled) + repeatChar('-', empty) + "]"
}

func repeatChar(char rune, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += string(char)
	}
	return result
}

func decodeResponse(resp *http.Response, v interface{}) error {
	return api.DecodeResponse(resp, v)
}
