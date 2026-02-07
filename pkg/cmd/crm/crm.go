// Package crm provides CRM dashboard commands for GitScrum CLI
package crm

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
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
		fmt.Println("error: not authenticated")
		return nil
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/overview"
	resp, err := client.Get(path)
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

	d := result.Data

	fmt.Println("CRM DASHBOARD")
	fmt.Println(strings.Repeat("═", 60))
	fmt.Println()

	// Summary
	fmt.Println("SUMMARY")
	fmt.Printf("  Clients: %d | Active Projects: %d | Total Projects: %d\n",
		d.Summary.TotalClients, d.Summary.ActiveProjects, d.Summary.TotalProjects)
	if d.Summary.ProjectsAtRisk > 0 {
		fmt.Printf("  [!] Projects at Risk: %d\n", d.Summary.ProjectsAtRisk)
	}
	fmt.Println()

	// Invoices
	fmt.Println("INVOICES")
	fmt.Printf("  Paid: %d ($%.2f) | Pending: %d ($%.2f)\n",
		d.Invoices.Paid, d.Invoices.PaidAmount,
		d.Invoices.Pending, d.Invoices.PendingAmount)
	if d.Invoices.Overdue > 0 {
		fmt.Printf("  [!] Overdue: %d ($%.2f)\n", d.Invoices.Overdue, d.Invoices.OverdueAmount)
	}
	fmt.Println()

	// Proposals
	fmt.Println("PROPOSALS")
	fmt.Printf("  Approved: %d ($%.2f) | Pending: %d ($%.2f)\n",
		d.Proposals.Approved, d.Proposals.ApprovedValue,
		d.Proposals.PendingApproval, d.Proposals.PendingValue)
	if d.Proposals.ExpiringSoon > 0 {
		fmt.Printf("  [!] Expiring Soon: %d\n", d.Proposals.ExpiringSoon)
	}
	fmt.Println()

	// Alerts
	if d.Alerts.OverdueInvoices > 0 || d.Alerts.ExpiringProposals > 0 || d.Alerts.ProjectsAtRisk > 0 {
		fmt.Println("ALERTS")
		if d.Alerts.OverdueInvoices > 0 {
			fmt.Printf("  [!] %d overdue invoices\n", d.Alerts.OverdueInvoices)
		}
		if d.Alerts.ExpiringProposals > 0 {
			fmt.Printf("  [!] %d expiring proposals\n", d.Alerts.ExpiringProposals)
		}
		if d.Alerts.ProjectsAtRisk > 0 {
			fmt.Printf("  [!] %d projects at risk\n", d.Alerts.ProjectsAtRisk)
		}
	}

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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/revenue-pipeline"
	resp, err := client.Get(path)
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

	fmt.Println("REVENUE PIPELINE")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Invoices Summary
	fmt.Println("INVOICES BY STATUS:")
	for status, data := range result.Data.InvoicesSummary {
		fmt.Printf("  %-12s %d invoices  $%.2f\n", status+":", data.Count, data.Total)
	}
	fmt.Println()

	// Proposals Summary
	fmt.Println("PROPOSALS BY STATUS:")
	for status, data := range result.Data.ProposalsSummary {
		fmt.Printf("  %-12s %d proposals $%.2f\n", status+":", data.Count, data.Total)
	}
	fmt.Println()

	// Overdue Invoices
	if len(result.Data.OverdueInvoices) > 0 {
		fmt.Println("OVERDUE INVOICES:")
		for _, inv := range result.Data.OverdueInvoices {
			fmt.Printf("  [!] %s - %s: %s%.2f (%d days overdue)\n",
				inv.Series, inv.Client.Name, inv.CurrencySymbol, inv.Amount, inv.DaysOverdue)
		}
		fmt.Println()
	}

	// Monthly Revenue
	if len(result.Data.MonthlyRevenue) > 0 {
		fmt.Println("MONTHLY REVENUE:")
		for _, m := range result.Data.MonthlyRevenue {
			fmt.Printf("  %s: $%.2f (%d invoices)\n", m.Month, m.Total, m.Count)
		}
	}

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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/clients-at-risk"
	resp, err := client.Get(path)
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

	fmt.Println("CLIENTS AT RISK")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	// Summary
	s := result.Data.Summary
	fmt.Printf("Total at risk: %d\n", s.TotalAtRisk)
	if s.WithOverdueInvoices > 0 {
		fmt.Printf("  With overdue invoices: %d\n", s.WithOverdueInvoices)
	}
	if s.WithStalledProjects > 0 {
		fmt.Printf("  With stalled projects: %d\n", s.WithStalledProjects)
	}
	if s.WithExpiringProposals > 0 {
		fmt.Printf("  With expiring proposals: %d\n", s.WithExpiringProposals)
	}
	fmt.Println()

	if len(result.Data.ClientsAtRisk) == 0 {
		fmt.Println("No clients at risk")
		return nil
	}

	for _, c := range result.Data.ClientsAtRisk {
		fmt.Printf("[!] %s\n", c.Name)
		for _, r := range c.Risks {
			fmt.Printf("    - %s\n", r.Label)
		}
		fmt.Println()
	}

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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/pending-approvals"
	resp, err := client.Get(path)
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

	fmt.Println("PENDING APPROVALS")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	s := result.Data.Summary
	fmt.Printf("Total Pending: %d\n", s.Total)
	fmt.Printf("  Proposals: %d ($%.2f)\n", s.Proposals, s.ProposalsValue)
	fmt.Printf("  Invoices: %d ($%.2f)\n", s.Invoices, s.InvoicesValue)
	fmt.Printf("  Change Requests: %d\n", s.ChangeRequests)
	fmt.Println()

	if len(result.Data.All) == 0 {
		fmt.Println("No pending approvals")
		return nil
	}

	fmt.Println("PENDING ITEMS:")
	for _, item := range result.Data.All {
		typeLabel := strings.ToUpper(item.Type[:1]) + item.Type[1:]
		title := item.Title
		if title == "" {
			title = item.Code
		}
		fmt.Printf("  [%s] %s - %s\n", typeLabel, title, item.Client.Name)
		if item.Amount > 0 {
			fmt.Printf("         %s%.2f | Waiting %d days\n", item.CurrencySymbol, item.Amount, item.DaysWaiting)
		} else {
			fmt.Printf("         Waiting %d days\n", item.DaysWaiting)
		}
	}

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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/projects-health"
	resp, err := client.Get(path)
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

	fmt.Println("PROJECTS HEALTH")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	s := result.Data.Summary
	fmt.Printf("Total: %d | Healthy: %d | Warning: %d | Critical: %d\n",
		s.Total, s.Healthy, s.Warning, s.Critical)
	if s.OverBudget > 0 || s.AtRisk > 0 {
		fmt.Printf("Over Budget: %d | At Risk: %d\n", s.OverBudget, s.AtRisk)
	}
	fmt.Println()

	if len(result.Data.Projects) == 0 {
		fmt.Println("No projects found")
		return nil
	}

	// Show critical and warning projects
	for _, p := range result.Data.Projects {
		if p.HealthStatus == "critical" || p.HealthStatus == "warning" {
			icon := "[!]"
			if p.HealthStatus == "critical" {
				icon = "[!!]"
			}
			fmt.Printf("%s %s", icon, p.Name)
			if p.ClientName != "" {
				fmt.Printf(" (%s)", p.ClientName)
			}
			fmt.Println()
			fmt.Printf("    Progress: %.0f%% | Budget: %.0f%% used (%.1fh / %.1fh)\n",
				p.ProgressPercentage, p.BudgetUsagePercentage, p.HoursUsed, p.BudgetHours)
		}
	}

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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/client-flow/dashboard/leaderboard"
	resp, err := client.Get(path)
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

	fmt.Println("CLIENT LEADERBOARD")
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()

	if len(result.Data.Clients) == 0 {
		fmt.Println("No clients found")
		return nil
	}

	for i, c := range result.Data.Clients {
		if i >= 10 {
			break
		}
		fmt.Printf("%2d. %s\n", i+1, c.Name)
		fmt.Printf("    Revenue: $%.2f | Projects: %d active / %d total\n",
			c.TotalRevenue, c.ActiveProjects, c.TotalProjects)
		fmt.Printf("    Reliability: %.0f%% | Avg Payment: %.0f days\n",
			c.ReliabilityScore*100, c.AvgPaymentDays)
		fmt.Println()
	}

	return nil
}
