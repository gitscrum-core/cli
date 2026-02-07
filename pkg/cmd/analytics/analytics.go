// Package analytics provides workspace analytics commands for GitScrum CLI
package analytics

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdAnalytics creates the analytics command group
func NewCmdAnalytics(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analytics [command]",
		Short: "View workspace analytics",
		Long: `View workspace-level analytics and metrics.

Provides quick access to velocity trends, team workload, blockers,
cycle time, and throughput metrics.`,
		Aliases: []string{"stats", "metrics"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(NewCmdVelocity(f))
	cmd.AddCommand(NewCmdWorkload(f))
	cmd.AddCommand(NewCmdBlockers(f))
	cmd.AddCommand(NewCmdCycleTime(f))
	cmd.AddCommand(NewCmdThroughput(f))

	return cmd
}

// NewCmdVelocity shows velocity trends
func NewCmdVelocity(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "velocity",
		Short: "Show velocity trend (last 4 weeks)",
		Long:  `Display weekly velocity trend showing tasks created, completed, and effort points.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVelocity(f)
		},
	}
}

func runVelocity(f *factory.Factory) error {
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

	path := "/companies/analytics/velocity"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			WeeklyTrend []struct {
				Week      string `json:"week"`
				WeekNum   string `json:"week_num"`
				Created   int    `json:"created"`
				Completed int    `json:"completed"`
				Effort    int    `json:"effort_completed"`
				Net       int    `json:"net"`
			} `json:"weekly_trend"`
			ActiveSprints []struct {
				Title         string  `json:"title"`
				Project       string  `json:"project_name"`
				Percent       float64 `json:"percent"`
				DaysRemaining int     `json:"days_remaining"`
			} `json:"active_sprints"`
			QuickStats struct {
				CompletedThisWeek int `json:"completed_this_week"`
				TotalInProgress   int `json:"total_in_progress"`
			} `json:"quick_stats"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("VELOCITY TREND (Last 4 Weeks)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println()

	fmt.Printf("  Completed this week: %d\n", result.Data.QuickStats.CompletedThisWeek)
	fmt.Printf("  In progress: %d\n", result.Data.QuickStats.TotalInProgress)
	fmt.Println()

	if len(result.Data.WeeklyTrend) > 0 {
		fmt.Println("  Week        Created  Completed  Effort  Net")
		fmt.Println("  " + strings.Repeat("-", 48))
		for _, w := range result.Data.WeeklyTrend {
			netSymbol := ""
			if w.Net > 0 {
				netSymbol = "+"
			}
			fmt.Printf("  %-10s  %7d  %9d  %6d  %s%d\n",
				w.Week, w.Created, w.Completed, w.Effort, netSymbol, w.Net)
		}
		fmt.Println()
	}

	if len(result.Data.ActiveSprints) > 0 {
		fmt.Println("ACTIVE SPRINTS:")
		for _, s := range result.Data.ActiveSprints {
			bar := renderProgressBar(s.Percent, 20)
			fmt.Printf("  %s %s %.0f%% (%d days left)\n",
				bar, s.Title, s.Percent, s.DaysRemaining)
		}
	}

	return nil
}

// NewCmdWorkload shows team workload
func NewCmdWorkload(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "workload",
		Short: "Show team workload distribution",
		Long:  `Display task distribution across team members.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkload(f)
		},
	}
}

func runWorkload(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/analytics/workload"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Members []struct {
				Name       string `json:"name"`
				Username   string `json:"username"`
				TotalTasks int    `json:"total_tasks"`
				InProgress int    `json:"in_progress"`
				Overdue    int    `json:"overdue"`
				LoadLevel  string `json:"load_level"`
				IsTracking bool   `json:"is_tracking"`
			} `json:"members"`
			Unassigned int `json:"unassigned"`
			TeamSize   int `json:"team_size"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("TEAM WORKLOAD")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n  Team size: %d members | Unassigned tasks: %d\n\n",
		result.Data.TeamSize, result.Data.Unassigned)

	if len(result.Data.Members) == 0 {
		fmt.Println("  No team members with assigned tasks")
		return nil
	}

	fmt.Println("  Name                 Tasks  In Progress  Overdue  Status")
	fmt.Println("  " + strings.Repeat("-", 55))

	for _, m := range result.Data.Members {
		status := getLoadIndicator(m.LoadLevel)
		tracking := ""
		if m.IsTracking {
			tracking = " [T]"
		}
		name := truncateString(m.Name, 18)
		fmt.Printf("  %-18s  %5d  %11d  %7d  %s%s\n",
			name, m.TotalTasks, m.InProgress, m.Overdue, status, tracking)
	}

	fmt.Println()
	fmt.Println("  Status: [OK] balanced  [!] busy  [!!] overloaded  [T] tracking")

	return nil
}

// NewCmdBlockers shows blocked tasks
func NewCmdBlockers(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "blockers",
		Short: "Show blocked tasks",
		Long:  `Display blocked tasks with days blocked info.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runBlockers(f)
		},
	}
}

func runBlockers(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/analytics/blockers"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Summary struct {
				Total          int     `json:"total"`
				AvgDaysBlocked float64 `json:"avg_days_blocked"`
			} `json:"summary"`
			Tasks []struct {
				Code        string `json:"code"`
				Title       string `json:"title"`
				Project     string `json:"project"`
				DaysBlocked int    `json:"days_blocked"`
				BlockerBy   *struct {
					Name     string `json:"name"`
					Username string `json:"username"`
				} `json:"blocker_by"`
			} `json:"tasks"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("BLOCKED TASKS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n  Total blocked: %d | Avg days blocked: %.1f\n\n",
		result.Data.Summary.Total, result.Data.Summary.AvgDaysBlocked)

	if len(result.Data.Tasks) == 0 {
		fmt.Println("  No blocked tasks")
		return nil
	}

	for _, t := range result.Data.Tasks {
		daysBadge := getDaysBadge(t.DaysBlocked)
		title := truncateString(t.Title, 35)
		fmt.Printf("  %s %s %s\n", t.Code, daysBadge, title)
		blocker := "Unknown"
		if t.BlockerBy != nil {
			blocker = t.BlockerBy.Name
		}
		fmt.Printf("       -> Project: %s | Blocked by: %s\n", t.Project, blocker)
	}

	return nil
}

// NewCmdCycleTime shows cycle time analytics
func NewCmdCycleTime(f *factory.Factory) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "cycle-time",
		Short: "Show average cycle time",
		Long:  `Display average time from work start to completion.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCycleTime(f, days)
		},
	}

	cmd.Flags().IntVar(&days, "days", 30, "Period in days (max 90)")

	return cmd
}

func runCycleTime(f *factory.Factory, days int) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/companies/analytics/cycle-time?days=%d", days)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			AverageHours   float64 `json:"average_hours"`
			TotalCompleted int     `json:"total_completed"`
			PeriodDays     int     `json:"period_days"`
			ByType         []struct {
				Type     string  `json:"type"`
				Count    int     `json:"count"`
				AvgHours float64 `json:"avg_hours"`
			} `json:"by_type"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("CYCLE TIME ANALYTICS")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n  Period: Last %d days\n", result.Data.PeriodDays)
	fmt.Printf("  Tasks completed: %d\n", result.Data.TotalCompleted)
	fmt.Printf("  Average cycle time: %.1f hours (%.1f days)\n\n",
		result.Data.AverageHours, result.Data.AverageHours/24)

	if len(result.Data.ByType) > 0 {
		fmt.Println("  By Issue Type:")
		fmt.Println("  " + strings.Repeat("-", 40))
		fmt.Println("  Type            Count   Avg Hours")
		for _, t := range result.Data.ByType {
			typeName := truncateString(t.Type, 14)
			fmt.Printf("  %-14s  %5d   %9.1f\n", typeName, t.Count, t.AvgHours)
		}
	}

	return nil
}

// NewCmdThroughput shows throughput analytics
func NewCmdThroughput(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "throughput",
		Short: "Show tasks completed per week",
		Long:  `Display weekly throughput (tasks completed).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runThroughput(f)
		},
	}
}

func runThroughput(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/analytics/throughput"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Weekly []struct {
				Week      string `json:"week"`
				Label     string `json:"label"`
				Completed int    `json:"completed"`
				Created   int    `json:"created"`
				Net       int    `json:"net"`
			} `json:"weekly"`
			TotalCompleted int     `json:"total_completed"`
			AveragePerWeek float64 `json:"average_per_week"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Println("THROUGHPUT (Tasks/Week)")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("\n  Total completed: %d | Average/week: %.1f\n\n",
		result.Data.TotalCompleted, result.Data.AveragePerWeek)

	if len(result.Data.Weekly) == 0 {
		fmt.Println("  No data available")
		return nil
	}

	maxCompleted := 0
	for _, w := range result.Data.Weekly {
		if w.Completed > maxCompleted {
			maxCompleted = w.Completed
		}
	}

	for _, w := range result.Data.Weekly {
		bar := renderHorizontalBar(w.Completed, maxCompleted, 30)
		fmt.Printf("  %-10s %s %d\n", w.Week, bar, w.Completed)
	}

	return nil
}

// Helper functions

func renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + strings.Repeat("#", filled) + strings.Repeat("-", empty) + "]"
}

func renderHorizontalBar(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("-", width)
	}
	filled := int(float64(value) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("#", filled) + strings.Repeat("-", width-filled)
}

func getLoadIndicator(level string) string {
	switch level {
	case "overloaded", "high":
		return "[!!]"
	case "busy", "medium":
		return "[!]"
	default:
		return "[OK]"
	}
}

func getDaysBadge(days int) string {
	if days >= 7 {
		return fmt.Sprintf("[!!] %dd", days)
	} else if days >= 3 {
		return fmt.Sprintf("[!] %dd", days)
	}
	return fmt.Sprintf("%dd", days)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
