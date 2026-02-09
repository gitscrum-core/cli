// Package workspaces provides workspace commands for GitScrum CLI
package workspaces

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
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
		return err
	}

	sp := spinner.New("Loading workspaces...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	resp, err := client.Get("/companies")
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Workspace `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		output.Empty(i18n.T("no_workspaces"), i18n.T("create_workspace_hint"))
		return nil
	}

	currentWorkspace, _ := f.CurrentWorkspace()

	output.Header("Workspaces")

	for _, w := range result.Data {
		name := w.Name
		if w.Plan != "" {
			name += fmt.Sprintf(" [%s]", strings.ToUpper(w.Plan))
		}

		if w.Slug == currentWorkspace {
			output.Successf("▶ %s", name)
		} else {
			output.Infof("  %s", name)
		}
		output.Dimf("  Slug: %s │ %d projects │ %d members", w.Slug, w.ProjectsCount, w.MembersCount)
	}

	if currentWorkspace == "" {
		fmt.Println()
		output.Info("Set default workspace: gitscrum config set workspace <slug>")
	}

	fmt.Println()
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
		_, err := f.RequireWorkspace()
		if err != nil {
			return err
		}
	}

	sp := spinner.New("Loading workspace...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/companies/%s", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Workspace `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	w := result.Data

	output.Header(w.Name)

	output.KeyValue("Slug", w.Slug)
	if w.Plan != "" {
		output.KeyValue("Plan", strings.ToUpper(w.Plan))
	}
	if w.Description != "" {
		fmt.Println()
		output.Dim(w.Description)
	}
	fmt.Println()
	output.KeyValuef("Projects", "%d", w.ProjectsCount)
	output.KeyValuef("Members", "%d", w.MembersCount)

	fmt.Println()
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

	sp := spinner.New("Loading statistics...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/companies/workspace-stats"
	resp, err := client.Get(path)
	sp.Stop()
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

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	s := result.Data

	output.Header(fmt.Sprintf("Workspace Statistics: %s", slug))

	output.KeyValuef("Projects", "%d", s.TotalProjects)
	output.KeyValuef("Tasks", "%d total, %d completed", s.TotalTasks, s.CompletedTasks)
	output.KeyValuef("Sprints", "%d active", s.ActiveSprints)
	output.KeyValuef("Members", "%d", s.TotalMembers)

	hours := s.TotalTime / 60
	output.KeyValuef("Total Time", "%dh", hours)

	fmt.Println()
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

	output.Successf("Switched to workspace: %s", slug)
	output.Infof("Now set a project: gitscrum projects switch <slug>")

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

	sp := spinner.New("Loading members...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/workspace-members/%s", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			UUID  string `json:"uuid"`
			Name  string `json:"name"`
			Email string `json:"email"`
			Role  string `json:"role"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header(fmt.Sprintf("Members of %s", slug))

	if len(result.Data) == 0 {
		output.Empty("No members found", "")
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
		output.SubHeader("Owners")
		for _, o := range owners {
			output.Bulletf("👑 %s", o)
		}
	}

	if len(managers) > 0 {
		output.SubHeader("Managers")
		for _, m := range managers {
			output.Bulletf("⭐ %s", m)
		}
	}

	if len(members) > 0 {
		output.SubHeader("Members")
		for _, m := range members {
			output.Bullet(m)
		}
	}

	fmt.Println()
	return nil
}
