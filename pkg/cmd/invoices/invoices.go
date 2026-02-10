// Package invoices provides invoice commands for GitScrum CLI
package invoices

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdInvoices creates the invoices command group
func NewCmdInvoices(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "invoices [command]",
		Short: "Manage invoices",
		Long: `View and manage invoices in GitScrum.

Without a subcommand, lists all invoices.`,
		Aliases: []string{"invoice", "inv"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicesList(f)
		},
	}

	cmd.AddCommand(NewCmdInvoicesView(f))
	cmd.AddCommand(NewCmdInvoicesCreate(f))
	cmd.AddCommand(NewCmdInvoicesSend(f))
	cmd.AddCommand(NewCmdInvoicesMarkPaid(f))

	return cmd
}

// Invoice represents an invoice (matches CompanyInvoiceResource.php)
type Invoice struct {
	UUID    string `json:"uuid"`
	RefCode string `json:"ref_code"`
	Series  string `json:"series"`
	Status  struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"status"`
	GrossTotal          int    `json:"gross_total"`
	GrossTotalFormatted string `json:"gross_total_formatted"`
	Currency            struct {
		Symbol string `json:"symbol"`
		Code   string `json:"code"`
	} `json:"currency"`
	PaymentDueAt *api.DateResource `json:"payment_due_at"`
	PaidAt       *api.DateResource `json:"paid_at"`
	Contact      struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
	} `json:"contact"`
}

func runInvoicesList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading invoices...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/company-invoices?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		output.EmptyContext(i18n.T("no_invoices"), workspace, "", i18n.T("create_invoice_hint"))
		return nil
	}

	output.Header("Invoices")

	for _, inv := range result.Data {
		currency := inv.Currency.Code
		if currency == "" {
			currency = "USD"
		}

		// Use ref_code if available, fallback to first 8 chars of UUID
		code := inv.RefCode
		if code == "" && len(inv.UUID) >= 8 {
			code = inv.UUID[:8]
		}

		switch inv.Status.Name {
		case "Paid":
			output.Successf("✓ [%s] %s — %s %s", code, inv.Series, currency, inv.GrossTotalFormatted)
		case "Pending":
			output.Infof("→ [%s] %s — %s %s", code, inv.Series, currency, inv.GrossTotalFormatted)
		default:
			output.Dimf("○ [%s] %s — %s %s (%s)", code, inv.Series, currency, inv.GrossTotalFormatted, inv.Status.Name)
		}
		output.Dimf("   Contact: %s │ Due: %s", inv.Contact.Name, inv.PaymentDueAt.ISODate())
	}

	fmt.Println()
	output.Dimf("View details: gitscrum invoices view <code>")
	return nil
}

// NewCmdInvoicesView shows invoice details
func NewCmdInvoicesView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <code>",
		Short: "View invoice details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicesView(f, args[0])
		},
	}
}

func runInvoicesView(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading invoice...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s", code)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	inv := result.Data
	currency := inv.Currency.Code
	if currency == "" {
		currency = "USD"
	}

	output.Header(fmt.Sprintf("Invoice %s", inv.Series))

	output.KeyValue("Contact", inv.Contact.Name)
	output.KeyValuef("Amount", "%s %s", currency, inv.GrossTotalFormatted)

	switch inv.Status.Name {
	case "Paid":
		output.Success("Status: Paid")
	case "Draft":
		output.Dim("Status: Draft")
	default:
		output.KeyValue("Status", inv.Status.Name)
	}

	output.KeyValue("Due Date", inv.PaymentDueAt.ISODate())
	if inv.PaidAt != nil && inv.PaidAt.ISODate() != "" {
		output.KeyValue("Paid At", inv.PaidAt.ISODate())
	}

	fmt.Println()
	return nil
}

// NewCmdInvoicesCreate creates a new invoice
func NewCmdInvoicesCreate(f *factory.Factory) *cobra.Command {
	var clientSlug, dueDate, currency string
	var amount float64

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new invoice",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicesCreate(f, clientSlug, amount, currency, dueDate)
		},
	}

	cmd.Flags().StringVar(&clientSlug, "client", "", "Client slug (required)")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Invoice amount (required)")
	cmd.Flags().StringVar(&currency, "currency", "USD", "Currency")
	cmd.Flags().StringVar(&dueDate, "due", "", "Due date (YYYY-MM-DD)")
	cmd.MarkFlagRequired("client")
	cmd.MarkFlagRequired("amount")

	return cmd
}

func runInvoicesCreate(f *factory.Factory, clientSlug string, amount float64, currency, dueDate string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Creating invoice...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"client_slug": clientSlug,
		"amount":      amount,
		"currency":    currency,
	}
	if dueDate != "" {
		body["due_date"] = dueDate
	}

	path := "/company-invoices"
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Invoice created: %s", result.Data.Series)
	output.KeyValuef("Amount", "%s %s", result.Data.Currency.Code, result.Data.GrossTotalFormatted)
	output.KeyValue("Contact", result.Data.Contact.Name)

	return nil
}

// NewCmdInvoicesSend sends an invoice
func NewCmdInvoicesSend(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "send <code>",
		Short: "Send invoice to client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicesSend(f, args[0])
		},
	}
}

func runInvoicesSend(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Sending invoice...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s/send", code)
	_, err = client.Post(path, nil)
	sp.Stop()
	if err != nil {
		return err
	}

	output.Successf("Invoice %s sent to client", code)
	return nil
}

// NewCmdInvoicesMarkPaid marks invoice as paid
func NewCmdInvoicesMarkPaid(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "mark-paid <code>",
		Short: "Mark invoice as paid",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInvoicesMarkPaid(f, args[0])
		},
	}
}

func runInvoicesMarkPaid(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Marking as paid...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s/paid", code)
	_, err = client.Post(path, nil)
	sp.Stop()
	if err != nil {
		return err
	}

	output.Successf("Invoice %s marked as paid", code)
	return nil
}
