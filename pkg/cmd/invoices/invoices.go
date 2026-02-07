// Package invoices provides invoice commands for GitScrum CLI
package invoices

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
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

// Invoice represents an invoice
type Invoice struct {
	UUID       string  `json:"uuid"`
	Code       string  `json:"code"`
	Status     string  `json:"status"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	DueDate    string  `json:"due_date"`
	PaidAt     string  `json:"paid_at"`
	Client     struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"client"`
}

func runInvoicesList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/company-invoices"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No invoices found")
		return nil
	}

	fmt.Println("INVOICES:")
	fmt.Println()

	for _, inv := range result.Data {
		status := getInvoiceStatusIcon(inv.Status)
		currency := inv.Currency
		if currency == "" {
			currency = "USD"
		}
		fmt.Printf("  %s %s - %s %.2f\n", status, inv.Code, currency, inv.Amount)
		fmt.Printf("     Client: %s\n", inv.Client.Name)
		fmt.Printf("     Due: %s • Status: %s\n", inv.DueDate, inv.Status)
		fmt.Println()
	}

	return nil
}

func getInvoiceStatusIcon(status string) string {
	switch status {
	case "paid":
		return "[paid]"
	case "sent":
		return "[sent]"
	case "overdue":
		return "[!!]"
	case "draft":
		return "[draft]"
	default:
		return "[-]"
	}
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s", code)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	inv := result.Data
	currency := inv.Currency
	if currency == "" {
		currency = "USD"
	}

	fmt.Printf("Invoice %s\n", inv.Code)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("\nClient: %s\n", inv.Client.Name)
	fmt.Printf("Amount: %s %.2f\n", currency, inv.Amount)
	fmt.Printf("Status: %s %s\n", getInvoiceStatusIcon(inv.Status), inv.Status)
	fmt.Printf("Due Date: %s\n", inv.DueDate)
	if inv.PaidAt != "" {
		fmt.Printf("Paid At: %s\n", inv.PaidAt)
	}

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


	client, err := f.APIClient()
	if err != nil {
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
	if err != nil {
		return err
	}

	var result struct {
		Data Invoice `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Invoice created: %s\n", result.Data.Code)
	fmt.Printf("  Amount: %s %.2f\n", result.Data.Currency, result.Data.Amount)
	fmt.Printf("  Client: %s\n", result.Data.Client.Name)

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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s/send", code)
	_, err = client.Post(path, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Invoice %s sent to client\n", code)
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/company-invoices/%s/paid", code)
	_, err = client.Post(path, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Invoice %s marked as paid\n", code)
	return nil
}
