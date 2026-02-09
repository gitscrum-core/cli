// Package standup provides standup commands for GitScrum CLI
package standup

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

func NewCmdStandup(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "standup",
		Short:   "Daily standup summary",
		Long:    "View and create daily standup reports.\n\nWithout a subcommand, shows your standup summary (yesterday, today, blockers).",
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

func runStandupSummary(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}
	if _, err := f.RequireWorkspace(); err != nil {
		return err
	}
	project, _ := f.CurrentProject()

	sp := spinner.New("Loading standup...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		sp.Stop()
		return err
	}

	today := time.Now().Format("2006-01-02")
	path := fmt.Sprintf("/companies/standup/summary?company_slug=%s&date=%s", workspace, today)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Standup `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header(fmt.Sprintf("Daily Standup — %s", today))
	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		project, _ := f.CurrentProject()
		output.EmptyContext(i18n.T("no_standup_entries"), workspace, project, i18n.T("create_standup_hint"))
		return nil
	}
	for _, s := range result.Data {
		printStandupEntry(s)
	}
	return nil
}

func printStandupEntry(s Standup) {
	output.SubHeader(s.User.Name)
	if len(s.Completed) > 0 {
		output.Success("Completed Yesterday")
		for _, item := range s.Completed {
			output.Bullet(item)
		}
	}
	if len(s.Planned) > 0 {
		output.Info("Planned for Today")
		for _, item := range s.Planned {
			output.Bullet(item)
		}
	}
	if len(s.Blockers) > 0 {
		output.Warning("Blockers")
		for _, item := range s.Blockers {
			output.Alert(item)
		}
	}
	fmt.Println()
}

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
	sp := spinner.New("Loading completed tasks...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		sp.Stop()
		return err
	}

	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	path := fmt.Sprintf("/companies/standup/completed-yesterday?company_slug=%s&date=%s", workspace, yesterday)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	sp.Stop()
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
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Header(fmt.Sprintf("Completed Yesterday (%s)", yesterday))
	if len(result.Data) == 0 {
		output.Empty("No tasks completed yesterday", "")
		return nil
	}
	for _, t := range result.Data {
		output.Successf("[%s] %s", t.Code, t.Title)
	}
	fmt.Println()
	return nil
}

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
	sp := spinner.New("Loading blockers...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/standup/blockers?company_slug=%s", workspace)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Code     string `json:"code"`
			Title    string `json:"title"`
			Assignee struct {
				Name string `json:"name"`
			} `json:"assignee"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Header("Current Blockers")
	if len(result.Data) == 0 {
		output.Success("No blockers")
		return nil
	}
	for _, t := range result.Data {
		assignee := t.Assignee.Name
		if assignee == "" {
			assignee = "unassigned"
		}
		output.Warningf("[%s] %s", t.Code, t.Title)
		output.Dimf("Assigned to: %s", assignee)
	}
	fmt.Println()
	return nil
}

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
	sp := spinner.New("Loading team status...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/standup/team-status?company_slug=%s", workspace)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			UserUUID            string  `json:"user_uuid"`
			UserName            string  `json:"user_name"`
			UserAvatar          string  `json:"user_avatar"`
			TasksInProgress     int     `json:"tasks_in_progress"`
			TasksCompletedToday int     `json:"tasks_completed_today"`
			TasksCompletedWeek  int     `json:"tasks_completed_week"`
			BlockedCount        int     `json:"blocked_count"`
			HoursTrackedToday   float64 `json:"hours_tracked_today"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Header("Team Status")
	if len(result.Data) == 0 {
		output.Empty("No team members found", "")
		return nil
	}
	for _, m := range result.Data {
		if m.BlockedCount > 0 {
			output.Warningf("%s — blocked", m.UserName)
		} else if m.TasksInProgress > 0 {
			output.Successf("%s — working", m.UserName)
		} else {
			output.Infof("%s — available", m.UserName)
		}
		output.Dimf("In Progress: %d │ Done Today: %d │ Blocked: %d",
			m.TasksInProgress, m.TasksCompletedToday, m.BlockedCount)
		if m.HoursTrackedToday > 0 {
			output.Dimf("Hours Tracked: %.1fh", m.HoursTrackedToday)
		}
	}
	fmt.Println()
	return nil
}

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
	sp := spinner.New("Loading weekly digest...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/standup/weekly-digest?company_slug=%s", workspace)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	sp.Stop()
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
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Header("Weekly Digest")
	output.KeyValuef("Tasks Completed", "%d", result.Data.TotalCompleted)
	output.KeyValuef("Hours Tracked", "%.1fh", result.Data.TotalHours)
	if result.Data.VelocityChange != 0 {
		sign := "+"
		if result.Data.VelocityChange < 0 {
			sign = ""
		}
		if result.Data.VelocityChange > 0 {
			output.Successf("Velocity: %s%.0f%%", sign, result.Data.VelocityChange)
		} else {
			output.Warningf("Velocity: %s%.0f%%", sign, result.Data.VelocityChange)
		}
	}
	if len(result.Data.TopContributors) > 0 {
		output.SubHeader("Top Contributors")
		for i, c := range result.Data.TopContributors {
			if i >= 5 {
				break
			}
			output.Bulletf("#%d  %s (%d tasks)", i+1, c.Name, c.TasksCompleted)
		}
	}
	if len(result.Data.DailyBreakdown) > 0 {
		output.SubHeader("Daily Breakdown")
		for _, d := range result.Data.DailyBreakdown {
			details := fmt.Sprintf("%d completed, %d created", d.Completed, d.Created)
			if d.Blocked > 0 {
				details += fmt.Sprintf(", %d blocked", d.Blocked)
			}
			output.KeyValue(d.Date, details)
		}
	}
	fmt.Println()
	return nil
}
