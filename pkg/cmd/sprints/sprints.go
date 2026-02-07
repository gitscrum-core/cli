// Package sprints provides sprint commands for GitScrum CLI
package sprints

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdSprints creates the sprints command group
func NewCmdSprints(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sprints [command]",
		Short: "Manage sprints",
		Long: `View and manage sprints in GitScrum.

Sprints are time-boxed iterations for organizing work. View current sprint,
check burndown charts, and track team velocity.

Without a subcommand, lists sprints for the current project.`,
		Example: `  # List all sprints
  gitscrum sprints

  # View current active sprint
  gitscrum sprints current

  # View sprint details
  gitscrum sprints view sprint-1

  # View sprint burndown chart
  gitscrum sprints burndown sprint-1

  # View sprint statistics
  gitscrum sprints stats sprint-1

  # Create a new sprint
  gitscrum sprints create -n "Sprint 5" --start 2024-03-01 --end 2024-03-14`,
		Aliases: []string{"sprint", "sp"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSprintsList(f)
		},
	}

	cmd.AddCommand(NewCmdSprintsCurrent(f))
	cmd.AddCommand(NewCmdSprintsView(f))
	cmd.AddCommand(NewCmdSprintsCreate(f))
	cmd.AddCommand(NewCmdSprintsBurndown(f))
	cmd.AddCommand(NewCmdSprintsStats(f))

	return cmd
}

func runSprintsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	project, _ := f.CurrentProject()

	if workspace == "" || project == "" {
		return fmt.Errorf("workspace and project required. Use -w and -p flags or set defaults")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/sprints"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No sprints found")
		return nil
	}

	fmt.Printf("Sprints in %s/%s:\n\n", workspace, project)
	
	for _, s := range result.Data {
		status := getSprintStatus(s)
		fmt.Printf("  %s %s\n", status, s.Title)
		fmt.Printf("    %s -> %s\n", s.DateStart, s.DateEnd)
		if s.TotalTasks > 0 {
			completed := float64(s.CompletedTasks) / float64(s.TotalTasks) * 100
			fmt.Printf("    %d/%d tasks (%.0f%%)\n", s.CompletedTasks, s.TotalTasks, completed)
		}
		fmt.Println()
	}

	return nil
}

// Sprint represents a sprint
type Sprint struct {
	UUID           string `json:"uuid"`
	Slug           string `json:"slug"`
	Title          string `json:"title"`
	DateStart      string `json:"date_start"`
	DateEnd        string `json:"date_end"`
	Status         string `json:"status"`
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
	TotalPoints    int    `json:"total_points"`
	CompletedPoints int   `json:"completed_points"`
}

func getSprintStatus(s Sprint) string {
	now := time.Now()
	start, _ := time.Parse("2006-01-02", s.DateStart)
	end, _ := time.Parse("2006-01-02", s.DateEnd)

	if now.Before(start) {
		return "[pending]"
	}
	if now.After(end) {
		return "[done]"
	}
	return "[active]"
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
		fmt.Println("error: not authenticated")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	project, _ := f.CurrentProject()

	if workspace == "" || project == "" {
		return fmt.Errorf("workspace and project required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/sprints?current=true"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return fmt.Errorf("no active sprint found")
	}

	s := result.Data

	fmt.Printf("CURRENT SPRINT: %s\n", s.Title)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("\nPeriod: %s -> %s\n", s.DateStart, s.DateEnd)
	
	// Calculate days remaining
	end, _ := time.Parse("2006-01-02", s.DateEnd)
	daysLeft := int(time.Until(end).Hours() / 24)
	if daysLeft > 0 {
		fmt.Printf("Days remaining: %d\n", daysLeft)
	} else {
		fmt.Println("Sprint ended")
	}

	fmt.Println()
	
	// Progress
	if s.TotalTasks > 0 {
		progress := float64(s.CompletedTasks) / float64(s.TotalTasks) * 100
		bar := renderProgressBar(progress, 30)
		fmt.Printf("Tasks:  %s %.0f%% (%d/%d)\n", bar, progress, s.CompletedTasks, s.TotalTasks)
	}
	
	if s.TotalPoints > 0 {
		progress := float64(s.CompletedPoints) / float64(s.TotalPoints) * 100
		bar := renderProgressBar(progress, 30)
		fmt.Printf("Points: %s %.0f%% (%d/%d)\n", bar, progress, s.CompletedPoints, s.TotalPoints)
	}

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
		Use:   "view <slug>",
		Short: "View sprint details",
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/sprints/%s", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	s := result.Data
	fmt.Printf("Sprint: %s\n", s.Title)
	fmt.Printf("Period: %s → %s\n", s.DateStart, s.DateEnd)
	fmt.Printf("Status: %s\n", getSprintStatus(s))
	fmt.Printf("Tasks: %d/%d completed\n", s.CompletedTasks, s.TotalTasks)
	fmt.Printf("Points: %d/%d completed\n", s.CompletedPoints, s.TotalPoints)

	return nil
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"title":      title,
		"date_start": startDate,
		"date_end":   endDate,
	}

	path := "/sprints"
	resp, err := client.Post(path, body)
	if err != nil {
		return err
	}

	var result struct {
		Data Sprint `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Sprint created: %s\n", result.Data.Title)
	fmt.Printf("  Period: %s → %s\n", result.Data.DateStart, result.Data.DateEnd)

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


	client, err := f.APIClient()
	if err != nil {
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
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Days []struct {
				Date     string `json:"date"`
				Ideal    int    `json:"ideal"`
				Actual   int    `json:"actual"`
			} `json:"days"`
			TotalPoints int `json:"total_points"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return fmt.Errorf("failed to get burndown data: %w", err)
	}

	// Render ASCII chart
	fmt.Println("SPRINT BURNDOWN")
	fmt.Println(strings.Repeat("─", 60))
	
	maxPoints := result.Data.TotalPoints
	if maxPoints == 0 {
		fmt.Println("No points data available")
		return nil
	}

	for _, day := range result.Data.Days {
		idealBar := int(float64(day.Ideal) / float64(maxPoints) * 40)
		actualBar := int(float64(day.Actual) / float64(maxPoints) * 40)
		
		fmt.Printf("%s │", day.Date[5:10]) // MM-DD
		fmt.Printf(" %s", strings.Repeat("░", idealBar))
		fmt.Println()
		fmt.Printf("       │")
		fmt.Printf(" %s %d pts", strings.Repeat("█", actualBar), day.Actual)
		fmt.Println()
	}

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

	fmt.Println("SPRINT STATISTICS")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Println("  Coming soon: velocity, cycle time, scope changes...")
	fmt.Println()
	fmt.Println("  Use 'gitscrum sprints current' for basic KPIs")

	return nil
}
