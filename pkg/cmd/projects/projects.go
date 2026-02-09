// Package projects provides project management commands for GitScrum CLI
package projects

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

func NewCmdProjects(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects [command]",
		Short:   "Manage projects",
		Long:    "View and manage projects in your workspace.\n\nWithout a subcommand, lists all projects.",
		Aliases: []string{"project", "proj"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectsList(f)
		},
	}
	cmd.AddCommand(NewCmdProjectsView(f))
	cmd.AddCommand(NewCmdProjectsStats(f))
	cmd.AddCommand(NewCmdProjectsCreate(f))
	cmd.AddCommand(NewCmdProjectsMembers(f))
	cmd.AddCommand(NewCmdProjectsSwitch(f))
	return cmd
}

type Project struct {
	UUID         string `json:"uuid"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	TasksCount   int    `json:"tasks_count"`
	MembersCount int    `json:"members_count"`
	IsActive     bool   `json:"is_active"`
}

func runProjectsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}
	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading projects...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}
	path := fmt.Sprintf("/projects?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		output.EmptyContext(i18n.T("no_projects"), workspace, "", i18n.T("create_project_hint"))
		return nil
	}
	output.Header("Projects")
	currentProject, _ := f.CurrentProject()
	for _, p := range result.Data {
		marker := "  "
		if p.Slug == currentProject {
			marker = "▸ "
		}
		if p.IsActive {
			output.Successf("%s%s", marker, p.Name)
		} else {
			output.Dimf("%s%s (inactive)", marker, p.Name)
		}
		output.Dimf("  %d tasks · %d members", p.TasksCount, p.MembersCount)
	}
	fmt.Println()
	return nil
}

func NewCmdProjectsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view [slug]",
		Short: "View project details",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runProjectsView(f, slug)
		},
	}
}

func runProjectsView(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}
	if slug == "" {
		var err error
		slug, err = f.RequireProject()
		if err != nil {
			return err
		}
	}

	sp := spinner.New("Loading project...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}
	path := fmt.Sprintf("/projects/%s", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	p := result.Data
	output.Header(p.Name)
	if p.Description != "" {
		output.Dim(p.Description)
		fmt.Println()
	}
	output.KeyValue("Slug", p.Slug)
	output.KeyValue("Status", p.Status)
	output.KeyValuef("Tasks", "%d", p.TasksCount)
	output.KeyValuef("Members", "%d", p.MembersCount)
	fmt.Println()
	return nil
}

func NewCmdProjectsStats(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [slug]",
		Short: "Show project statistics",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runProjectsStats(f, slug)
		},
	}
}

func runProjectsStats(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}
	if slug == "" {
		var err error
		slug, err = f.RequireProject()
		if err != nil {
			return err
		}
	}

	sp := spinner.New("Loading statistics...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}
	path := fmt.Sprintf("/projects/%s/stats", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			TotalTasks     int     `json:"total_tasks"`
			CompletedTasks int     `json:"completed_tasks"`
			OpenTasks      int     `json:"open_tasks"`
			OverdueTasks   int     `json:"overdue_tasks"`
			TotalMembers   int     `json:"total_members"`
			TotalSprints   int     `json:"total_sprints"`
			CompletionRate float64 `json:"completion_rate"`
			AvgCycleTime   float64 `json:"avg_cycle_time"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	d := result.Data
	output.Header(fmt.Sprintf("Project Statistics — %s", slug))
	output.KeyValuef("Total Tasks", "%d", d.TotalTasks)
	output.KeyValuef("Completed", "%d", d.CompletedTasks)
	output.KeyValuef("Open", "%d", d.OpenTasks)
	if d.OverdueTasks > 0 {
		output.Warningf("Overdue: %d", d.OverdueTasks)
	}
	output.KeyValuef("Completion Rate", "%.1f%%", d.CompletionRate)
	output.KeyValuef("Avg Cycle Time", "%.1f days", d.AvgCycleTime)
	output.KeyValuef("Members", "%d", d.TotalMembers)
	output.KeyValuef("Sprints", "%d", d.TotalSprints)
	fmt.Println()
	return nil
}

func NewCmdProjectsCreate(f *factory.Factory) *cobra.Command {
	var description string
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectsCreate(f, args[0], description)
		},
	}
	cmd.Flags().StringVarP(&description, "description", "d", "", "Project description")
	return cmd
}

func runProjectsCreate(f *factory.Factory, name, description string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Creating project...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}
	body := map[string]interface{}{
		"name":        name,
		"description": description,
	}
	resp, err := client.Post("/projects", body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Successf("Project created: %s", result.Data.Name)
	output.Dimf("Slug: %s", result.Data.Slug)
	return nil
}

func NewCmdProjectsMembers(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "members [slug]",
		Short: "List project members",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runProjectsMembers(f, slug)
		},
	}
}

func runProjectsMembers(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}
	if slug == "" {
		var err error
		slug, err = f.RequireProject()
		if err != nil {
			return err
		}
	}

	sp := spinner.New("Loading members...")
	sp.Start()
	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}
	path := fmt.Sprintf("/projects/%s/members", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			UUID   string `json:"uuid"`
			Name   string `json:"name"`
			Email  string `json:"email"`
			Role   string `json:"role"`
			Avatar string `json:"avatar"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}
	output.Header(fmt.Sprintf("Members — %s", slug))
	if len(result.Data) == 0 {
		output.Empty("No members", "")
		return nil
	}
	for _, m := range result.Data {
		output.Bulletf("%s (%s) — %s", m.Name, m.Email, m.Role)
	}
	fmt.Println()
	return nil
}

func NewCmdProjectsSwitch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "switch <slug>",
		Short: "Switch to a different project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectsSwitch(f, args[0])
		},
	}
}

func runProjectsSwitch(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.Project = slug
	if err := cfg.Save(); err != nil {
		return err
	}

	output.Successf("Switched to project: %s", slug)
	return nil
}
