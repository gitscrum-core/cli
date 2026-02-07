// Package standup provides standup commands for GitScrum CLI
package standup

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdStandup creates the standup command group
func NewCmdStandup(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "standup",
		Short: "Daily standup summary",
		Long: `View and create daily standup reports.

Without a subcommand, shows your standup summary (yesterday, today, blockers).`,
		Aliases: []string{"daily", "stand"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupSummary(f)
		},
	}

	cmd.AddCommand(NewCmdStandupCompleted(f))
	cmd.AddCommand(NewCmdStandupBlockers(f))
	cmd.AddCommand(NewCmdStandupTeam(f))
	cmd.AddCommand(NewCmdStandupDigest(f))

	return cmd
}

func runStandupSummary(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	project, _ := f.CurrentProject()

	if workspace == "" {
		return fmt.Errorf("workspace required. Use -w flag or set default")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	path := fmt.Sprintf("/companies/standup/summary?date=%s", today)
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Standup `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("DAILY STANDUP - %s\n", today)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	if len(result.Data) == 0 {
		fmt.Println("No standup entries for today.")
		fmt.Println()
		fmt.Println("Create one with: gitscrum standup create")
		return nil
	}

	for _, s := range result.Data {
		printStandupEntry(s)
	}

	return nil
}

// Standup represents a standup entry
type Standup struct {
	UUID      string   `json:"uuid"`
	Date      string   `json:"date"`
	Completed []string `json:"completed"`
	Planned   []string `json:"planned"`
	Blockers  []string `json:"blockers"`
	User      struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

func printStandupEntry(s Standup) {
	fmt.Printf("%s\n", s.User.Name)
	
	if len(s.Completed) > 0 {
		fmt.Println("\n  COMPLETED YESTERDAY:")
		for _, item := range s.Completed {
			fmt.Printf("    • %s\n", item)
		}
	}
	
	if len(s.Planned) > 0 {
		fmt.Println("\n  PLANNED FOR TODAY:")
		for _, item := range s.Planned {
			fmt.Printf("    • %s\n", item)
		}
	}
	
	if len(s.Blockers) > 0 {
		fmt.Println("\n  BLOCKERS:")
		for _, item := range s.Blockers {
			fmt.Printf("    [!] %s\n", item)
		}
	}
	
	fmt.Println()
}

// NewCmdStandupCompleted shows completed tasks from yesterday
func NewCmdStandupCompleted(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "completed",
		Short: "Show what was completed yesterday",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupCompleted(f)
		},
	}
}

func runStandupCompleted(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	path := fmt.Sprintf("/companies/standup/completed-yesterday?date=%s", yesterday)
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Code  string `json:"code"`
			Title string `json:"title"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("COMPLETED YESTERDAY (%s):\n\n", yesterday)

	if len(result.Data) == 0 {
		fmt.Println("  No tasks completed yesterday")
		return nil
	}

	for _, t := range result.Data {
		fmt.Printf("  • [%s] %s\n", t.Code, t.Title)
	}

	return nil
}

// NewCmdStandupBlockers lists current blockers
func NewCmdStandupBlockers(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "blockers",
		Short: "List current blockers",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupBlockers(f)
		},
	}
}

func runStandupBlockers(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/standup/blockers"
	if project != "" {
		path += "?project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Code  string `json:"code"`
			Title string `json:"title"`
			Assignee struct {
				Name string `json:"name"`
			} `json:"assignee"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("CURRENT BLOCKERS:")
	fmt.Println()

	if len(result.Data) == 0 {
		fmt.Println("  No blockers")
		return nil
	}

	for _, t := range result.Data {
		assignee := t.Assignee.Name
		if assignee == "" {
			assignee = "unassigned"
		}
		fmt.Printf("  [!] [%s] %s\n", t.Code, t.Title)
		fmt.Printf("      Assigned to: %s\n\n", assignee)
	}

	return nil
}

// NewCmdStandupTeam shows team standup status
func NewCmdStandupTeam(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "team",
		Short: "Show team standup status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupTeam(f)
		},
	}
}

func runStandupTeam(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/standup/team-status"
	if project != "" {
		path += "?project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			UserUUID           string  `json:"user_uuid"`
			UserName           string  `json:"user_name"`
			UserAvatar         string  `json:"user_avatar"`
			TasksInProgress    int     `json:"tasks_in_progress"`
			TasksCompletedToday int    `json:"tasks_completed_today"`
			TasksCompletedWeek  int    `json:"tasks_completed_week"`
			BlockedCount       int     `json:"blocked_count"`
			HoursTrackedToday  float64 `json:"hours_tracked_today"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("TEAM STATUS")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if len(result.Data) == 0 {
		fmt.Println("  No team members found")
		return nil
	}

	for _, m := range result.Data {
		status := "available"
		if m.BlockedCount > 0 {
			status = "blocked"
		} else if m.TasksInProgress > 0 {
			status = "working"
		}

		fmt.Printf("  %s\n", m.UserName)
		fmt.Printf("    Status: %s | In Progress: %d | Done Today: %d | Blocked: %d\n",
			status, m.TasksInProgress, m.TasksCompletedToday, m.BlockedCount)
		if m.HoursTrackedToday > 0 {
			fmt.Printf("    Hours Tracked: %.1fh\n", m.HoursTrackedToday)
		}
		fmt.Println()
	}

	return nil
}

// NewCmdStandupDigest shows weekly digest
func NewCmdStandupDigest(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "digest",
		Short: "Show weekly standup digest",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStandupDigest(f)
		},
	}
}

func runStandupDigest(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/standup/weekly-digest"
	if project != "" {
		path += "?project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			TotalCompleted  int     `json:"total_completed"`
			TotalHours      float64 `json:"total_hours"`
			VelocityChange  float64 `json:"velocity_change"`
			TopContributors []struct {
				Name           string `json:"name"`
				TasksCompleted int    `json:"tasks_completed"`
			} `json:"top_contributors"`
			DailyBreakdown []struct {
				Date      string `json:"date"`
				Completed int    `json:"completed"`
				Created   int    `json:"created"`
				Blocked   int    `json:"blocked"`
			} `json:"daily_breakdown"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("WEEKLY DIGEST")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	fmt.Printf("  Tasks Completed: %d\n", result.Data.TotalCompleted)
	fmt.Printf("  Hours Tracked:   %.1fh\n", result.Data.TotalHours)
	if result.Data.VelocityChange != 0 {
		sign := "+"
		if result.Data.VelocityChange < 0 {
			sign = ""
		}
		fmt.Printf("  Velocity:        %s%.0f%%\n", sign, result.Data.VelocityChange)
	}
	fmt.Println()

	if len(result.Data.TopContributors) > 0 {
		fmt.Println("  TOP CONTRIBUTORS:")
		for i, c := range result.Data.TopContributors {
			if i >= 5 {
				break
			}
			fmt.Printf("    %d. %s (%d tasks)\n", i+1, c.Name, c.TasksCompleted)
		}
		fmt.Println()
	}

	if len(result.Data.DailyBreakdown) > 0 {
		fmt.Println("  DAILY BREAKDOWN:")
		for _, d := range result.Data.DailyBreakdown {
			fmt.Printf("    %s: %d completed, %d created", d.Date, d.Completed, d.Created)
			if d.Blocked > 0 {
				fmt.Printf(", %d blocked", d.Blocked)
			}
			fmt.Println()
		}
	}

	return nil
}



