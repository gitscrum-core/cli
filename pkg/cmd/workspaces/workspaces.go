// Package workspaces provides workspace commands for GitScrum CLI
package workspaces

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdWorkspaces creates the workspaces command group
func NewCmdWorkspaces(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspaces [command]",
		Short: "Manage workspaces",
		Long: `View and manage workspaces in GitScrum.

Without a subcommand, lists all accessible workspaces.`,
		Aliases: []string{"workspace", "ws"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspacesList(f)
		},
	}

	cmd.AddCommand(NewCmdWorkspacesView(f))
	cmd.AddCommand(NewCmdWorkspacesStats(f))
	cmd.AddCommand(NewCmdWorkspacesSwitch(f))
	cmd.AddCommand(NewCmdWorkspacesMembers(f))

	return cmd
}

// Workspace represents a workspace
type Workspace struct {
	UUID          string `json:"uuid"`
	Slug          string `json:"slug"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ProjectsCount int    `json:"projects_count"`
	MembersCount  int    `json:"members_count"`
	Plan          string `json:"plan"`
}

func runWorkspacesList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	resp, err := client.Get("/companies")
	if err != nil {
		return err
	}

	var result struct {
		Data []Workspace `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No workspaces found")
		fmt.Println()
		fmt.Println("Create one at: https://gitscrum.com/workspaces/create")
		return nil
	}

	currentWorkspace, _ := f.CurrentWorkspace()

	fmt.Println("WORKSPACES:")
	fmt.Println()

	for _, w := range result.Data {
		marker := "  "
		if w.Slug == currentWorkspace {
			marker = "▶ "
		}
		
		plan := ""
		if w.Plan != "" {
			plan = fmt.Sprintf(" [%s]", strings.ToUpper(w.Plan))
		}
		
		fmt.Printf("%s%s%s\n", marker, w.Name, plan)
		fmt.Printf("     Slug: %s\n", w.Slug)
		fmt.Printf("     %d projects | %d members\n", w.ProjectsCount, w.MembersCount)
		fmt.Println()
	}

	if currentWorkspace == "" {
		fmt.Println("Tip: Set default workspace: gitscrum config set workspace <slug>")
	}

	return nil
}

// NewCmdWorkspacesView shows workspace details
func NewCmdWorkspacesView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view [slug]",
		Short: "View workspace details",
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runWorkspacesView(f, slug)
		},
	}
}

func runWorkspacesView(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	if slug == "" {
		slug, _ = f.CurrentWorkspace()
	}

	if slug == "" {
		return fmt.Errorf("workspace required. Use 'gitscrum workspaces view <slug>'")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/companies/%s", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Workspace `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	w := result.Data

	fmt.Printf("%s\n", w.Name)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("\nSlug: %s\n", w.Slug)
	if w.Plan != "" {
		fmt.Printf("Plan: %s\n", strings.ToUpper(w.Plan))
	}
	if w.Description != "" {
		fmt.Printf("\n%s\n", w.Description)
	}
	fmt.Printf("\nProjects: %d\n", w.ProjectsCount)
	fmt.Printf("Members: %d\n", w.MembersCount)

	return nil
}

// NewCmdWorkspacesStats shows workspace statistics
func NewCmdWorkspacesStats(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats [slug]",
		Short: "Show workspace statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runWorkspacesStats(f, slug)
		},
	}
}

func runWorkspacesStats(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	if slug == "" {
		slug, _ = f.CurrentWorkspace()
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/companies/workspace-stats"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			TotalProjects  int `json:"total_projects"`
			TotalTasks     int `json:"total_tasks"`
			CompletedTasks int `json:"completed_tasks"`
			ActiveSprints  int `json:"active_sprints"`
			TotalMembers   int `json:"total_members"`
			TotalTime      int `json:"total_time_minutes"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	s := result.Data

	fmt.Printf("WORKSPACE STATISTICS: %s\n", slug)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	
	fmt.Printf("Projects:   %d\n", s.TotalProjects)
	fmt.Printf("Tasks:      %d total, %d completed\n", s.TotalTasks, s.CompletedTasks)
	fmt.Printf("Sprints:    %d active\n", s.ActiveSprints)
	fmt.Printf("Members:    %d\n", s.TotalMembers)
	
	hours := s.TotalTime / 60
	fmt.Printf("Total Time: %dh\n", hours)

	return nil
}

// NewCmdWorkspacesSwitch switches to a workspace
func NewCmdWorkspacesSwitch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "switch <slug>",
		Short:   "Switch to a workspace (set as default)",
		Aliases: []string{"use"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWorkspacesSwitch(f, args[0])
		},
	}
}

func runWorkspacesSwitch(f *factory.Factory, slug string) error {
	cfg, err := f.Config()
	if err != nil {
		return err
	}

	cfg.Workspace = slug
	if err := cfg.Save(); err != nil {
		return err
	}

	fmt.Printf("Switched to workspace: %s\n", slug)
	fmt.Println()
	fmt.Println("Now set a project: gitscrum projects switch <slug>")
	
	return nil
}

// NewCmdWorkspacesMembers shows workspace members
func NewCmdWorkspacesMembers(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "members [slug]",
		Short: "List workspace members",
		RunE: func(cmd *cobra.Command, args []string) error {
			slug := ""
			if len(args) > 0 {
				slug = args[0]
			}
			return runWorkspacesMembers(f, slug)
		},
	}
}

func runWorkspacesMembers(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	if slug == "" {
		slug, _ = f.CurrentWorkspace()
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/workspace-members/%s", slug)
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

	// Group by role
	owners := []string{}
	managers := []string{}
	members := []string{}

	for _, m := range result.Data {
		entry := fmt.Sprintf("%s (%s)", m.Name, m.Email)
		switch m.Role {
		case "agency_owner", "owner":
			owners = append(owners, entry)
		case "manager":
			managers = append(managers, entry)
		default:
			members = append(members, entry)
		}
	}

	if len(owners) > 0 {
		fmt.Println("  OWNERS:")
		for _, o := range owners {
			fmt.Printf("     • %s\n", o)
		}
		fmt.Println()
	}

	if len(managers) > 0 {
		fmt.Println("  ⭐ Managers:")
		for _, m := range managers {
			fmt.Printf("     • %s\n", m)
		}
		fmt.Println()
	}

	if len(members) > 0 {
		fmt.Println("  Members:")
		for _, m := range members {
			fmt.Printf("     • %s\n", m)
		}
	}

	return nil
}
