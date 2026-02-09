// Package crm provides CRM dashboard commands for GitScrum CLI
package crm

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdCRM creates the CRM command group
func NewCmdCRM(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "crm [command]",
		Short: "CRM dashboard and analytics",
		Long: `View CRM metrics and client analytics.

Without a subcommand, shows the CRM dashboard overview.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMDashboard(f)
		},
	}

	cmd.AddCommand(NewCmdCRMRevenue(f))
	cmd.AddCommand(NewCmdCRMAtRisk(f))
	cmd.AddCommand(NewCmdCRMPipeline(f))
	cmd.AddCommand(NewCmdCRMProjects(f))
	cmd.AddCommand(NewCmdCRMLeaderboard(f))

	return cmd
}

func runCRMDashboard(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading CRM dashboard...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/overview"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Summary struct {
				TotalClients    int `json:"total_clients"`
				ActiveProjects  int `json:"active_projects"`
				TotalProjects   int `json:"total_projects"`
				ProjectsAtRisk  int `json:"projects_at_risk"`
			} `json:"summary"`
			Invoices struct {
				Total         int     `json:"total"`
				Draft         int     `json:"draft"`
				Pending       int     `json:"pending"`
				Paid          int     `json:"paid"`
				Overdue       int     `json:"overdue"`
				PaidAmount    float64 `json:"paid_amount"`
				PendingAmount float64 `json:"pending_amount"`
				OverdueAmount float64 `json:"overdue_amount"`
			} `json:"invoices"`
			Proposals struct {
				Total           int     `json:"total"`
				Draft           int     `json:"draft"`
				PendingApproval int     `json:"pending_approval"`
				Approved        int     `json:"approved"`
				Rejected        int     `json:"rejected"`
				ExpiringSoon    int     `json:"expiring_soon"`
				ApprovedValue   float64 `json:"approved_value"`
				PendingValue    float64 `json:"pending_value"`
			} `json:"proposals"`
			Alerts struct {
				OverdueInvoices    int `json:"overdue_invoices"`
				ExpiringProposals  int `json:"expiring_proposals"`
				ProjectsAtRisk     int `json:"projects_at_risk"`
				ClientsWithOverdue int `json:"clients_with_overdue"`
			} `json:"alerts"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	d := result.Data

	ws, _ := f.CurrentWorkspace()
	if ws != "" {
		output.Header(fmt.Sprintf("CRM Dashboard (%s)", ws))
	} else {
		output.Header("CRM Dashboard")
	}

	// Summary
	output.SubHeader("Summary")
	output.KeyValuef("Clients", "%d", d.Summary.TotalClients)
	output.KeyValuef("Active Projects", "%d / %d", d.Summary.ActiveProjects, d.Summary.TotalProjects)
	if d.Summary.ProjectsAtRisk > 0 {
		output.Alertf("%d projects at risk", d.Summary.ProjectsAtRisk)
	}

	// Invoices
	output.SubHeader("Invoices")
	output.Successf("Paid: %d ($%.2f)", d.Invoices.Paid, d.Invoices.PaidAmount)
	output.Infof("Pending: %d ($%.2f)", d.Invoices.Pending, d.Invoices.PendingAmount)
	if d.Invoices.Overdue > 0 {
		output.Warningf("Overdue: %d ($%.2f)", d.Invoices.Overdue, d.Invoices.OverdueAmount)
	}

	// Proposals
	output.SubHeader("Proposals")
	output.Successf("Approved: %d ($%.2f)", d.Proposals.Approved, d.Proposals.ApprovedValue)
	output.Infof("Pending: %d ($%.2f)", d.Proposals.PendingApproval, d.Proposals.PendingValue)
	if d.Proposals.ExpiringSoon > 0 {
		output.Warningf("Expiring Soon: %d", d.Proposals.ExpiringSoon)
	}

	// Alerts
	if d.Alerts.OverdueInvoices > 0 || d.Alerts.ExpiringProposals > 0 || d.Alerts.ProjectsAtRisk > 0 {
		output.SubHeader("Alerts")
		if d.Alerts.OverdueInvoices > 0 {
			output.Warningf("%d overdue invoices", d.Alerts.OverdueInvoices)
		}
		if d.Alerts.ExpiringProposals > 0 {
			output.Warningf("%d expiring proposals", d.Alerts.ExpiringProposals)
		}
		if d.Alerts.ProjectsAtRisk > 0 {
			output.Warningf("%d projects at risk", d.Alerts.ProjectsAtRisk)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdCRMRevenue shows revenue pipeline
func NewCmdCRMRevenue(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "revenue",
		Short: "Show revenue pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMRevenue(f)
		},
	}
}

func runCRMRevenue(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading revenue data...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/revenue-pipeline"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			InvoicesSummary map[string]struct {
				Count int     `json:"count"`
				Total float64 `json:"total"`
			} `json:"invoices_summary"`
			ProposalsSummary map[string]struct {
				Count int     `json:"count"`
				Total float64 `json:"total"`
			} `json:"proposals_summary"`
			OverdueInvoices []struct {
				UUID           string  `json:"uuid"`
				Series         string  `json:"series"`
				Client         struct {
					Name string `json:"name"`
				} `json:"client"`
				Amount         float64 `json:"amount"`
				CurrencySymbol string  `json:"currency_symbol"`
				DaysOverdue    int     `json:"days_overdue"`
			} `json:"overdue_invoices"`
			MonthlyRevenue []struct {
				Month string  `json:"month"`
				Total float64 `json:"total"`
				Count int     `json:"count"`
			} `json:"monthly_revenue"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Revenue Pipeline")

	// Invoices Summary
	output.SubHeader("Invoices by Status")
	for status, data := range result.Data.InvoicesSummary {
		output.KeyValuef(strings.ToUpper(status[:1])+status[1:], "%d invoices — $%.2f", data.Count, data.Total)
	}

	// Proposals Summary
	output.SubHeader("Proposals by Status")
	for status, data := range result.Data.ProposalsSummary {
		output.KeyValuef(strings.ToUpper(status[:1])+status[1:], "%d proposals — $%.2f", data.Count, data.Total)
	}

	// Overdue Invoices
	if len(result.Data.OverdueInvoices) > 0 {
		output.SubHeader("Overdue Invoices")
		for _, inv := range result.Data.OverdueInvoices {
			output.Warningf("%s — %s: %s%.2f (%d days overdue)",
				inv.Series, inv.Client.Name, inv.CurrencySymbol, inv.Amount, inv.DaysOverdue)
		}
	}

	// Monthly Revenue
	if len(result.Data.MonthlyRevenue) > 0 {
		output.SubHeader("Monthly Revenue")
		for _, m := range result.Data.MonthlyRevenue {
			output.KeyValuef(m.Month, "$%.2f (%d invoices)", m.Total, m.Count)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdCRMAtRisk shows at-risk clients
func NewCmdCRMAtRisk(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "at-risk",
		Short: "Show clients at risk",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMAtRisk(f)
		},
	}
}

func runCRMAtRisk(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading at-risk clients...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/clients-at-risk"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			ClientsAtRisk []struct {
				UUID  string `json:"uuid"`
				Name  string `json:"name"`
				Email string `json:"email"`
				Risks []struct {
					Type  string `json:"type"`
					Label string `json:"label"`
				} `json:"risks"`
			} `json:"clients_at_risk"`
			Summary struct {
				WithOverdueInvoices   int `json:"with_overdue_invoices"`
				WithStalledProjects   int `json:"with_stalled_projects"`
				WithExpiringProposals int `json:"with_expiring_proposals"`
				TotalAtRisk           int `json:"total_at_risk"`
			} `json:"summary"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Clients at Risk")

	// Summary
	s := result.Data.Summary
	output.KeyValuef("Total at Risk", "%d", s.TotalAtRisk)
	if s.WithOverdueInvoices > 0 {
		output.Warningf("With overdue invoices: %d", s.WithOverdueInvoices)
	}
	if s.WithStalledProjects > 0 {
		output.Warningf("With stalled projects: %d", s.WithStalledProjects)
	}
	if s.WithExpiringProposals > 0 {
		output.Warningf("With expiring proposals: %d", s.WithExpiringProposals)
	}

	if len(result.Data.ClientsAtRisk) == 0 {
		fmt.Println()
		output.Success("No clients at risk")
		return nil
	}

	fmt.Println()
	for _, c := range result.Data.ClientsAtRisk {
		output.Warning(c.Name)
		for _, r := range c.Risks {
			output.Dim("- " + r.Label)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdCRMPipeline shows pending approvals
func NewCmdCRMPipeline(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "pipeline",
		Short: "Show pending approvals pipeline",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMPipeline(f)
		},
	}
}

func runCRMPipeline(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading pipeline...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/pending-approvals"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			All []struct {
				Type   string `json:"type"`
				UUID   string `json:"uuid"`
				Code   string `json:"code"`
				Title  string `json:"title"`
				Client struct {
					Name string `json:"name"`
				} `json:"client"`
				Amount         float64 `json:"amount"`
				CurrencySymbol string  `json:"currency_symbol"`
				DaysWaiting    int     `json:"days_waiting"`
			} `json:"all"`
			Summary struct {
				Total          int     `json:"total"`
				Proposals      int     `json:"proposals"`
				ChangeRequests int     `json:"change_requests"`
				Invoices       int     `json:"invoices"`
				ProposalsValue float64 `json:"proposals_value"`
				InvoicesValue  float64 `json:"invoices_value"`
			} `json:"summary"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Pending Approvals")

	ps := result.Data.Summary
	output.KeyValuef("Total Pending", "%d", ps.Total)
	output.KeyValuef("Proposals", "%d ($%.2f)", ps.Proposals, ps.ProposalsValue)
	output.KeyValuef("Invoices", "%d ($%.2f)", ps.Invoices, ps.InvoicesValue)
	output.KeyValuef("Change Requests", "%d", ps.ChangeRequests)

	if len(result.Data.All) == 0 {
		fmt.Println()
		output.Success("No pending approvals")
		return nil
	}

	output.SubHeader("Pending Items")
	for _, item := range result.Data.All {
		typeLabel := strings.ToUpper(item.Type[:1]) + item.Type[1:]
		title := item.Title
		if title == "" {
			title = item.Code
		}
		output.Bulletf("[%s] %s — %s", typeLabel, title, item.Client.Name)
		if item.Amount > 0 {
			output.Dimf("%s%.2f │ Waiting %d days", item.CurrencySymbol, item.Amount, item.DaysWaiting)
		} else {
			output.Dimf("Waiting %d days", item.DaysWaiting)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdCRMProjects shows projects health
func NewCmdCRMProjects(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "projects",
		Short: "Show projects health overview",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMProjects(f)
		},
	}
}

func runCRMProjects(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading projects health...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/projects-health"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Projects []struct {
				UUID                  string  `json:"uuid"`
				Slug                  string  `json:"slug"`
				Name                  string  `json:"name"`
				ClientName            string  `json:"client_name"`
				BudgetHours           float64 `json:"budget_hours"`
				HoursUsed             float64 `json:"hours_used"`
				BudgetUsagePercentage float64 `json:"budget_usage_percentage"`
				IsOverBudget          bool    `json:"is_over_budget"`
				ProgressPercentage    float64 `json:"progress_percentage"`
				HealthStatus          string  `json:"health_status"`
			} `json:"projects"`
			Summary struct {
				Total      int `json:"total"`
				Healthy    int `json:"healthy"`
				Warning    int `json:"warning"`
				Critical   int `json:"critical"`
				OverBudget int `json:"over_budget"`
				AtRisk     int `json:"at_risk"`
			} `json:"summary"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Projects Health")

	ps := result.Data.Summary
	output.Successf("Healthy: %d", ps.Healthy)
	output.Warningf("Warning: %d", ps.Warning)
	output.Errorf("Critical: %d", ps.Critical)
	if ps.OverBudget > 0 || ps.AtRisk > 0 {
		output.Alertf("Over Budget: %d │ At Risk: %d", ps.OverBudget, ps.AtRisk)
	}

	if len(result.Data.Projects) == 0 {
		fmt.Println()
		output.Empty("No projects found", "")
		return nil
	}

	// Show critical and warning projects
	fmt.Println()
	for _, p := range result.Data.Projects {
		if p.HealthStatus == "critical" || p.HealthStatus == "warning" {
			if p.HealthStatus == "critical" {
				label := p.Name
				if p.ClientName != "" {
					label += " (" + p.ClientName + ")"
				}
				output.Error(label)
			} else {
				label := p.Name
				if p.ClientName != "" {
					label += " (" + p.ClientName + ")"
				}
				output.Warning(label)
			}
			output.Dimf("Progress: %.0f%% │ Budget: %.0f%% used (%.1fh / %.1fh)",
				p.ProgressPercentage, p.BudgetUsagePercentage, p.HoursUsed, p.BudgetHours)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdCRMLeaderboard shows client leaderboard
func NewCmdCRMLeaderboard(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "leaderboard",
		Short: "Show client leaderboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCRMLeaderboard(f)
		},
	}
}

func runCRMLeaderboard(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading leaderboard...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/client-flow/dashboard/leaderboard"
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Clients []struct {
				UUID             string  `json:"uuid"`
				Name             string  `json:"name"`
				TotalRevenue     float64 `json:"total_revenue"`
				PendingRevenue   float64 `json:"pending_revenue"`
				TotalProjects    int     `json:"total_projects"`
				ActiveProjects   int     `json:"active_projects"`
				ReliabilityScore float64 `json:"reliability_score"`
				AvgPaymentDays   float64 `json:"avg_payment_days"`
			} `json:"clients"`
			SortBy string `json:"sort_by"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header("Client Leaderboard")

	if len(result.Data.Clients) == 0 {
		output.Empty("No clients found", "Add clients with the GitScrum web app.")
		return nil
	}

	for i, c := range result.Data.Clients {
		if i >= 10 {
			break
		}
		output.Bulletf("#%d  %s", i+1, c.Name)
		output.Dimf("Revenue: $%.2f │ Projects: %d active / %d total",
			c.TotalRevenue, c.ActiveProjects, c.TotalProjects)
		output.Dimf("Reliability: %.0f%% │ Avg Payment: %.0f days",
			c.ReliabilityScore*100, c.AvgPaymentDays)
	}

	fmt.Println()
	return nil
}
