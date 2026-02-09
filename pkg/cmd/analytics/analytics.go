// Package analytics provides workspace analytics commands for GitScrum CLI
package analytics

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
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
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading velocity data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/analytics/velocity?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Velocity Trend (Last 4 Weeks)")

	output.KeyValuef("Completed this week", "%d", result.Data.QuickStats.CompletedThisWeek)
	output.KeyValuef("In progress", "%d", result.Data.QuickStats.TotalInProgress)

	if len(result.Data.WeeklyTrend) > 0 {
		output.SubHeader("Weekly Breakdown")
		fmt.Println("  Week        Created  Completed  Effort  Net")
		fmt.Println("  " + strings.Repeat("─", 48))
		for _, w := range result.Data.WeeklyTrend {
			netSymbol := ""
			if w.Net > 0 {
				netSymbol = "+"
			}
			fmt.Printf("  %-10s  %7d  %9d  %6d  %s%d\n",
				w.Week, w.Created, w.Completed, w.Effort, netSymbol, w.Net)
		}
	}

	if len(result.Data.ActiveSprints) > 0 {
		output.SubHeader("Active Sprints")
		for _, s := range result.Data.ActiveSprints {
			bar := renderProgressBar(s.Percent, 20)
			output.Bulletf("%s %s %.0f%% (%d days left)", bar, s.Title, s.Percent, s.DaysRemaining)
		}
	}

	fmt.Println()
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
		return err
	}

	if _, err := f.RequireWorkspace(); err != nil {
		return err
	}

	sp := spinner.New("Loading workload data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/companies/analytics/workload"
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Team Workload")

	output.KeyValuef("Team size", "%d members", result.Data.TeamSize)
	output.KeyValuef("Unassigned tasks", "%d", result.Data.Unassigned)

	if len(result.Data.Members) == 0 {
		fmt.Println()
		output.Empty("No team members with assigned tasks", "")
		return nil
	}

	fmt.Println()
	fmt.Println("  Name                 Tasks  In Progress  Overdue  Status")
	fmt.Println("  " + strings.Repeat("─", 55))

	for _, m := range result.Data.Members {
		status := getLoadIndicator(m.LoadLevel)
		tracking := ""
		if m.IsTracking {
			tracking = " [T]"
		}
		name := output.Truncate(m.Name, 18)
		fmt.Printf("  %-18s  %5d  %11d  %7d  %s%s\n",
			name, m.TotalTasks, m.InProgress, m.Overdue, status, tracking)
	}

	fmt.Println()
	output.Dim("Status: [OK] balanced  [!] busy  [!!] overloaded  [T] tracking")

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
		return err
	}

	if _, err := f.RequireWorkspace(); err != nil {
		return err
	}

	sp := spinner.New("Loading blockers...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/companies/analytics/blockers"
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Blocked Tasks")

	output.KeyValuef("Total blocked", "%d", result.Data.Summary.Total)
	output.KeyValuef("Avg days blocked", "%.1f", result.Data.Summary.AvgDaysBlocked)

	if len(result.Data.Tasks) == 0 {
		fmt.Println()
		output.Success("No blocked tasks")
		return nil
	}

	fmt.Println()
	for _, t := range result.Data.Tasks {
		title := output.Truncate(t.Title, 35)
		if t.DaysBlocked >= 7 {
			output.Errorf("%s  %dd blocked  %s", t.Code, t.DaysBlocked, title)
		} else if t.DaysBlocked >= 3 {
			output.Warningf("%s  %dd blocked  %s", t.Code, t.DaysBlocked, title)
		} else {
			output.Infof("%s  %dd blocked  %s", t.Code, t.DaysBlocked, title)
		}
		blocker := "Unknown"
		if t.BlockerBy != nil {
			blocker = t.BlockerBy.Name
		}
		output.Dimf("Project: %s │ Blocked by: %s", t.Project, blocker)
	}

	fmt.Println()
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
		return err
	}

	if _, err := f.RequireWorkspace(); err != nil {
		return err
	}

	sp := spinner.New("Loading cycle time data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/analytics/cycle-time?days=%d", days)
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Cycle Time Analytics")

	output.KeyValuef("Period", "Last %d days", result.Data.PeriodDays)
	output.KeyValuef("Tasks completed", "%d", result.Data.TotalCompleted)
	output.KeyValuef("Average cycle time", "%.1f hours (%.1f days)",
		result.Data.AverageHours, result.Data.AverageHours/24)

	if len(result.Data.ByType) > 0 {
		output.SubHeader("By Issue Type")
		fmt.Println("  Type            Count   Avg Hours")
		fmt.Println("  " + strings.Repeat("─", 40))
		for _, t := range result.Data.ByType {
			typeName := output.Truncate(t.Type, 14)
			fmt.Printf("  %-14s  %5d   %9.1f\n", typeName, t.Count, t.AvgHours)
		}
	}

	fmt.Println()
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
		return err
	}

	if _, err := f.RequireWorkspace(); err != nil {
		return err
	}

	sp := spinner.New("Loading throughput data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/companies/analytics/throughput"
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Throughput (Tasks/Week)")

	output.KeyValuef("Total completed", "%d", result.Data.TotalCompleted)
	output.KeyValuef("Average/week", "%.1f", result.Data.AveragePerWeek)

	if len(result.Data.Weekly) == 0 {
		fmt.Println()
		output.Empty("No data available", "")
		return nil
	}

	maxCompleted := 0
	for _, w := range result.Data.Weekly {
		if w.Completed > maxCompleted {
			maxCompleted = w.Completed
		}
	}

	fmt.Println()
	for _, w := range result.Data.Weekly {
		bar := renderHorizontalBar(w.Completed, maxCompleted, 30)
		fmt.Printf("  %-10s %s %d\n", w.Week, bar, w.Completed)
	}

	fmt.Println()
	return nil
}

// Helper functions

func renderProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	empty := width - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func renderHorizontalBar(value, max, width int) string {
	if max == 0 {
		return strings.Repeat("░", width)
	}
	filled := int(float64(value) / float64(max) * float64(width))
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
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
