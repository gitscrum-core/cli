// Package projects provides project commands for GitScrum CLI
package projects

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdProjects creates the projects command group
func NewCmdProjects(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects [command]",
		Short: "Manage projects",
		Long: `View and manage projects in GitScrum.

List projects, view details, check statistics, and manage team members.
Projects organize your tasks, sprints, and wiki within a workspace.

Without a subcommand, lists projects in the current workspace.`,
		Example: `  # List all projects
  gitscrum projects

  # View project details
  gitscrum projects view my-project

  # View project statistics
  gitscrum projects stats my-project

  # List project members
  gitscrum projects members my-project

  # Create a new project
  gitscrum projects create -n "My Project" -d "Description"

  # Switch default project
  gitscrum projects switch my-project`,
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

// Project represents a project
type Project struct {
	UUID        string `json:"uuid"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	TotalTasks  int    `json:"total_tasks"`
	OpenTasks   int    `json:"open_tasks"`
	Members     int    `json:"members_count"`
}

func runProjectsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required. Use -w flag or set default with 'gitscrum config set workspace'")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/projects"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No projects found in this workspace")
		return nil
	}

	currentProject, _ := f.CurrentProject()

	fmt.Printf("Projects in %s:\n\n", workspace)

	for _, p := range result.Data {
		marker := "  "
		if p.Slug == currentProject {
			marker = "▶ "
		}
		
		status := "*"
		if p.Status == "archived" {
			status = "[archived]"
		}
		
		fmt.Printf("%s%s %s\n", marker, status, p.Name)
		fmt.Printf("     Slug: %s\n", p.Slug)
		if p.TotalTasks > 0 {
			fmt.Printf("     %d tasks (%d open)\n", p.TotalTasks, p.OpenTasks)
		}
		if p.Members > 0 {
			fmt.Printf("     %d members\n", p.Members)
		}
		fmt.Println()
	}

	return nil
}

// NewCmdProjectsView shows project details
func NewCmdProjectsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view [slug]",
		Short: "View project details",
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

	workspace, _ := f.CurrentWorkspace()
	if slug == "" {
		slug, _ = f.CurrentProject()
	}

	if workspace == "" || slug == "" {
		return fmt.Errorf("workspace and project required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/projects/%s", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	p := result.Data

	fmt.Printf("%s\n", p.Name)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("\nSlug: %s\n", p.Slug)
	fmt.Printf("Status: %s\n", p.Status)
	if p.Description != "" {
		fmt.Printf("\n%s\n", p.Description)
	}
	fmt.Printf("\nTasks: %d total, %d open\n", p.TotalTasks, p.OpenTasks)
	fmt.Printf("Members: %d\n", p.Members)

	return nil
}

// NewCmdProjectsStats shows project statistics
func NewCmdProjectsStats(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [slug]",
		Short: "Show project statistics",
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
		slug, _ = f.CurrentProject()
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/projects/%s/stats", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			TotalTasks     int `json:"total_tasks"`
			CompletedTasks int `json:"completed_tasks"`
			OpenTasks      int `json:"open_tasks"`
			TotalTime      int `json:"total_time_minutes"`
			ActiveSprints  int `json:"active_sprints"`
			TotalMembers   int `json:"total_members"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	s := result.Data

	fmt.Printf("PROJECT STATISTICS: %s\n", slug)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	
	if s.TotalTasks > 0 {
		progress := float64(s.CompletedTasks) / float64(s.TotalTasks) * 100
		fmt.Printf("Tasks:      %d total, %d completed (%.0f%%)\n", s.TotalTasks, s.CompletedTasks, progress)
	}
	fmt.Printf("Open:       %d tasks\n", s.OpenTasks)
	
	hours := s.TotalTime / 60
	mins := s.TotalTime % 60
	fmt.Printf("Total Time: %dh %dm\n", hours, mins)
	
	fmt.Printf("Sprints:    %d active\n", s.ActiveSprints)
	fmt.Printf("Members:    %d\n", s.TotalMembers)

	return nil
}

// NewCmdProjectsCreate creates a new project
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

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"name": name,
	}
	if description != "" {
		body["description"] = description
	}

	path := "/projects"
	resp, err := client.Post(path, body)
	if err != nil {
		return err
	}

	var result struct {
		Data Project `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Project created: %s\n", result.Data.Name)
	fmt.Printf("  Slug: %s\n", result.Data.Slug)
	fmt.Println()
	fmt.Printf("Set as default: gitscrum config set project %s\n", result.Data.Slug)

	return nil
}

// NewCmdProjectsMembers shows project members
func NewCmdProjectsMembers(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "members [slug]",
		Short: "List project members",
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
		slug, _ = f.CurrentProject()
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/project-members/%s/members", slug)
	resp, err := client.Get(path)
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

	fmt.Printf("MEMBERS OF %s:\n\n", slug)

	if len(result.Data) == 0 {
		fmt.Println("  No members found")
		return nil
	}

	for _, m := range result.Data {
		role := ""
		if m.Role != "" {
			role = fmt.Sprintf(" (%s)", m.Role)
		}
		fmt.Printf("  • %s%s\n", m.Name, role)
		fmt.Printf("    %s\n\n", m.Email)
	}

	return nil
}

// NewCmdProjectsSwitch switches to a project
func NewCmdProjectsSwitch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "switch <slug>",
		Short:   "Switch to a project (set as default)",
		Aliases: []string{"use"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProjectsSwitch(f, args[0])
		},
	}
}

func runProjectsSwitch(f *factory.Factory, slug string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.Project = slug
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Switched to project: %s\n", slug)
	return nil
}
