// Package clients provides client CRM commands for GitScrum CLI
package clients

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

// joinWithSeparator joins strings with a separator
func joinWithSeparator(items []string, sep string) string {
	return strings.Join(items, sep)
}

// NewCmdClients creates the clients command group
func NewCmdClients(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clients [command]",
		Short: "Manage clients (CRM)",
		Long: `View and manage clients in GitScrum ClientFlow.

Without a subcommand, lists all clients.`,
		Aliases: []string{"client"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsList(f)
		},
	}

	cmd.AddCommand(NewCmdClientsView(f))
	cmd.AddCommand(NewCmdClientsCreate(f))
	cmd.AddCommand(NewCmdClientsStats(f))
	cmd.AddCommand(NewCmdClientsProjects(f))

	return cmd
}

// Client represents a contact company (matches ContactCompanyResource.php)
type Client struct {
	UUID                string            `json:"uuid"`
	RefCode             string            `json:"ref_code"`
	Name                string            `json:"name"`
	Email               string            `json:"email"`
	Phone               string            `json:"phone"`
	Website             string            `json:"website"`
	ProjectsCount       int               `json:"projects_count"`
	InvoicesCount       int               `json:"invoices_count"`
	ProposalsCount      int               `json:"proposals_count"`
	ChangeRequestsCount int               `json:"change_requests_count"`
	TotalPaid           int               `json:"total_paid"`
	TotalPending        int               `json:"total_pending"`
	HasOverdue          bool              `json:"has_overdue"`
	CreatedAt           *api.DateResource `json:"created_at"`
}

func runClientsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading clients...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/contact-companies/clients?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		output.EmptyContext(i18n.T("no_clients"), workspace, "", i18n.T("create_client_hint"))
		return nil
	}

	formatter := f.Formatter()

	// Build table with columns like tasks
	headers := []string{"Code", "Name", "Email", "Proj", "Prop", "Inv", "Revenue"}
	rows := make([][]string, 0, len(result.Data))

	for _, c := range result.Data {
		// Use ref_code if available, fallback to UUID shorthand
		code := c.RefCode
		if code == "" && len(c.UUID) >= 8 {
			code = c.UUID[:8]
		}

		// Name with overdue warning
		name := c.Name
		if c.HasOverdue {
			name = "⚠ " + name
		}
		if len(name) > 25 {
			name = name[:22] + "..."
		}

		// Email (truncated)
		email := c.Email
		if len(email) > 20 {
			email = email[:17] + "..."
		}

		// Counts
		proj := fmt.Sprintf("%d", c.ProjectsCount)
		prop := fmt.Sprintf("%d", c.ProposalsCount)
		inv := fmt.Sprintf("%d", c.InvoicesCount)

		// Revenue (total paid)
		revenue := "-"
		if c.TotalPaid > 0 {
			revenue = fmt.Sprintf("$%.0f", float64(c.TotalPaid)/100)
		}

		rows = append(rows, []string{code, name, email, proj, prop, inv, revenue})
	}

	formatter.PrintTable(headers, rows)

	fmt.Println()
	output.Dimf("View details: gitscrum clients view <code>")
	fmt.Println()
	return nil
}

// NewCmdClientsView shows client details
func NewCmdClientsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <name-or-uuid>",
		Short: "View client details",
		Long:  `View client details by name or UUID. Use 'gitscrum clients' to list all clients.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsView(f, args[0])
		},
	}
}

func runClientsView(f *factory.Factory, nameOrUUID string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading client...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	// Resolve name to UUID if needed
	clientUUID := nameOrUUID

	// Check if it looks like a UUID (has hyphens or is 32+ chars)
	isLikelyUUID := strings.Contains(nameOrUUID, "-") || len(nameOrUUID) >= 32

	if !isLikelyUUID {
		// Fetch clients list and find by name
		listPath := fmt.Sprintf("/contact-companies/clients?company_slug=%s", workspace)
		listResp, err := client.Get(listPath)
		if err != nil {
			sp.Stop()
			return err
		}

		var listResult struct {
			Data []Client `json:"data"`
		}
		if err := api.DecodeResponse(listResp, &listResult); err != nil {
			sp.Stop()
			return err
		}

		// Search by ref_code, name (case-insensitive), or partial UUID match
		searchLower := strings.ToLower(nameOrUUID)
		var foundClient *Client
		for i, c := range listResult.Data {
			if c.RefCode == nameOrUUID ||
				strings.ToLower(c.Name) == searchLower ||
				strings.HasPrefix(c.UUID, nameOrUUID) {
				foundClient = &listResult.Data[i]
				break
			}
		}

		if foundClient == nil {
			sp.Stop()
			return fmt.Errorf("client '%s' not found. Use 'gitscrum clients' to list all clients", nameOrUUID)
		}
		clientUUID = foundClient.UUID
	}

	path := fmt.Sprintf("/contact-companies/%s?company_slug=%s", clientUUID, workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	c := result.Data

	output.Header(c.Name)

	if c.Email != "" {
		output.KeyValue("Email", c.Email)
	}
	if c.Phone != "" {
		output.KeyValue("Phone", c.Phone)
	}
	if c.Website != "" {
		output.KeyValue("Website", c.Website)
	}

	output.SubHeader("Activity")
	output.KeyValuef("Projects", "%d", c.ProjectsCount)
	output.KeyValuef("Invoices", "%d", c.InvoicesCount)
	output.KeyValuef("Proposals", "%d", c.ProposalsCount)

	output.SubHeader("Financials")
	output.KeyValuef("Total Paid", "$%.2f", float64(c.TotalPaid)/100)
	if c.TotalPending > 0 {
		output.Warningf("Pending: $%.2f", float64(c.TotalPending)/100)
	} else {
		output.KeyValuef("Pending", "$%.2f", float64(c.TotalPending)/100)
	}
	if c.HasOverdue {
		output.Warning("⚠ Has overdue invoices")
	}

	if c.CreatedAt != nil {
		output.KeyValue("Client Since", c.CreatedAt.FormatDate())
	}

	fmt.Println()
	return nil
}

// NewCmdClientsCreate creates a new client
func NewCmdClientsCreate(f *factory.Factory) *cobra.Command {
	var email, phone, company string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsCreate(f, args[0], email, phone, company)
		},
	}

	cmd.Flags().StringVarP(&email, "email", "e", "", "Client email")
	cmd.Flags().StringVar(&phone, "phone", "", "Client phone")
	cmd.Flags().StringVarP(&company, "company", "c", "", "Company name")

	return cmd
}

func runClientsCreate(f *factory.Factory, name, email, phone, company string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Creating client...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"name": name,
	}
	if email != "" {
		body["email"] = email
	}
	if phone != "" {
		body["phone"] = phone
	}
	if company != "" {
		body["company"] = company
	}

	path := "/contact-companies"
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Successf("Client created: %s", result.Data.Name)
	output.KeyValue("UUID", result.Data.UUID)

	return nil
}

// NewCmdClientsStats shows client statistics
func NewCmdClientsStats(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "stats",
		Short: "Show client statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsStats(f)
		},
	}
}

func runClientsStats(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading statistics...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/contact-companies/stats?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	// Match the actual API response format from clientFlowStats
	var result struct {
		Data struct {
			Clients struct {
				Total        int `json:"total"`
				WithInvoices int `json:"with_invoices"`
				WithPending  int `json:"with_pending"`
				WithOverdue  int `json:"with_overdue"`
			} `json:"clients"`
			Invoices struct {
				Total         int `json:"total"`
				Draft         int `json:"draft"`
				Issued        int `json:"issued"`
				Paid          int `json:"paid"`
				Cancelled     int `json:"cancelled"`
				Overdue       int `json:"overdue"`
				TotalAmount   int `json:"total_amount"`
				PaidAmount    int `json:"paid_amount"`
				PendingAmount int `json:"pending_amount"`
				DraftAmount   int `json:"draft_amount"`
			} `json:"invoices"`
			Proposals struct {
				Total         int `json:"total"`
				Draft         int `json:"draft"`
				Sent          int `json:"sent"`
				Viewed        int `json:"viewed"`
				Approved      int `json:"approved"`
				Rejected      int `json:"rejected"`
				Expired       int `json:"expired"`
				Pending       int `json:"pending"`
				ApprovedValue int `json:"approved_value"`
			} `json:"proposals"`
			ChangeRequests struct {
				Total    int `json:"total"`
				Draft    int `json:"draft"`
				Sent     int `json:"sent"`
				Approved int `json:"approved"`
				Rejected int `json:"rejected"`
				Pending  int `json:"pending"`
			} `json:"change_requests"`
			Projects struct {
				Total     int `json:"total"`
				Active    int `json:"active"`
				Completed int `json:"completed"`
				Archived  int `json:"archived"`
			} `json:"projects"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	s := result.Data

	output.Header("Client Statistics")

	// Clients section
	output.SubHeader("Clients")
	output.KeyValuef("Total Clients", "%d", s.Clients.Total)
	if s.Clients.WithInvoices > 0 {
		output.Infof("With Invoices: %d", s.Clients.WithInvoices)
	}
	if s.Clients.WithOverdue > 0 {
		output.Warningf("With Overdue: %d", s.Clients.WithOverdue)
	}

	// Invoices section
	output.SubHeader("Invoices")
	output.KeyValuef("Total", "%d", s.Invoices.Total)
	if s.Invoices.Draft > 0 {
		output.Dimf("Draft: %d ($%.2f)", s.Invoices.Draft, float64(s.Invoices.DraftAmount)/100)
	}
	if s.Invoices.Issued > 0 {
		output.Infof("Pending: %d ($%.2f)", s.Invoices.Issued, float64(s.Invoices.PendingAmount)/100)
	}
	if s.Invoices.Paid > 0 {
		output.Successf("Paid: %d ($%.2f)", s.Invoices.Paid, float64(s.Invoices.PaidAmount)/100)
	}
	if s.Invoices.Overdue > 0 {
		output.Warningf("Overdue: %d", s.Invoices.Overdue)
	}

	// Proposals section
	output.SubHeader("Proposals")
	output.KeyValuef("Total", "%d", s.Proposals.Total)
	if s.Proposals.Pending > 0 {
		output.Infof("Pending: %d", s.Proposals.Pending)
	}
	if s.Proposals.Approved > 0 {
		output.Successf("Approved: %d ($%.2f)", s.Proposals.Approved, float64(s.Proposals.ApprovedValue)/100)
	}

	// Change Requests section
	if s.ChangeRequests.Total > 0 {
		output.SubHeader("Change Requests")
		output.KeyValuef("Total", "%d", s.ChangeRequests.Total)
		if s.ChangeRequests.Pending > 0 {
			output.Infof("Pending: %d", s.ChangeRequests.Pending)
		}
		if s.ChangeRequests.Approved > 0 {
			output.Successf("Approved: %d", s.ChangeRequests.Approved)
		}
	}

	// Projects section
	if s.Projects.Total > 0 {
		output.SubHeader("Projects")
		output.KeyValuef("Total", "%d", s.Projects.Total)
		if s.Projects.Active > 0 {
			output.Successf("Active: %d", s.Projects.Active)
		}
		if s.Projects.Completed > 0 {
			output.Infof("Completed: %d", s.Projects.Completed)
		}
	}

	// Revenue Summary section (using invoice amounts)
	if s.Invoices.TotalAmount > 0 || s.Invoices.PaidAmount > 0 {
		output.SubHeader("Revenue Summary")
		output.Successf("Total Received: $%.2f", float64(s.Invoices.PaidAmount)/100)
		if s.Invoices.PendingAmount > 0 {
			output.Warningf("Pending: $%.2f", float64(s.Invoices.PendingAmount)/100)
		}
		if s.Proposals.ApprovedValue > 0 {
			output.Infof("Approved Proposals: $%.2f", float64(s.Proposals.ApprovedValue)/100)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdClientsProjects lists client projects
func NewCmdClientsProjects(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "projects <slug>",
		Short: "List client projects",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsProjects(f, args[0])
		},
	}
}

func runClientsProjects(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading projects...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/projects?contact_company_uuid=%s", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Name   string  `json:"name"`
			Slug   string  `json:"slug"`
			Status string  `json:"status"`
			Budget float64 `json:"budget"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header(fmt.Sprintf("Projects for %s", slug))

	if len(result.Data) == 0 {
		output.Empty("No projects found", "")
		return nil
	}

	for _, p := range result.Data {
		output.Bulletf("%s (%s)", p.Name, p.Slug)
		if p.Budget > 0 {
			output.Dimf("Budget: $%.2f", p.Budget)
		}
	}

	fmt.Println()
	return nil
}
