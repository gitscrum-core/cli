// Package proposals provides proposal commands for GitScrum CLI
package proposals

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdProposals creates the proposals command group
func NewCmdProposals(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "proposals [command]",
		Short: "Manage proposals",
		Long: `View and manage proposals in GitScrum.

Without a subcommand, lists all proposals.`,
		Aliases: []string{"proposal", "prop"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProposalsList(f)
		},
	}

	cmd.AddCommand(NewCmdProposalsView(f))
	cmd.AddCommand(NewCmdProposalsCreate(f))
	cmd.AddCommand(NewCmdProposalsSend(f))
	cmd.AddCommand(NewCmdProposalsConvert(f))

	return cmd
}

// Proposal represents a proposal (matches ProposalResource.php)
type Proposal struct {
	UUID               string            `json:"uuid"`
	RefCode            string            `json:"ref_code"`
	Code               string            `json:"code"`
	Title              string            `json:"title"`
	Status             string            `json:"status"`
	StatusLabel        string            `json:"status_label"`
	TotalValueFormatted string           `json:"total_value_formatted"`
	Currency           string            `json:"currency"`
	ContactCompany     *struct {
		Name string `json:"name"`
		UUID string `json:"uuid"`
	} `json:"contact_company"`
	ClientName         string            `json:"client_name"`
	ExpiresAt          *api.DateResource `json:"expires_at"`
	CreatedAt          *api.DateResource `json:"created_at"`
}

func runProposalsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading proposals...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/proposals?company_slug=%s", workspace)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		output.EmptyContext(i18n.T("no_proposals"), workspace, "", i18n.T("create_proposal_hint"))
		return nil
	}

	output.Header("Proposals")

	for _, p := range result.Data {
		// Use ref_code if available, fallback to Code
		code := p.RefCode
		if code == "" {
			code = p.Code
		}

		switch p.Status {
		case "approved":
			output.Successf("✓ %s — %s", code, p.Title)
		case "rejected":
			output.Errorf("✗ %s — %s", code, p.Title)
		case "sent", "viewed":
			output.Infof("→ %s — %s", code, p.Title)
		case "expired":
			output.Warningf("⚠ %s — %s (expired)", code, p.Title)
		case "converted":
			output.Successf("⇒ %s — %s (converted)", code, p.Title)
		default:
			output.Dimf("○ %s — %s (%s)", code, p.Title, p.StatusLabel)
		}

		clientName := p.ClientName
		if clientName == "" && p.ContactCompany != nil {
			clientName = p.ContactCompany.Name
		}
		details := fmt.Sprintf("   Client: %s", clientName)
		if p.TotalValueFormatted != "" {
			details += fmt.Sprintf(" │ %s %s", p.Currency, p.TotalValueFormatted)
		}
		output.Dim(details)
	}

	fmt.Println()
	output.Dimf("View details: gitscrum proposals view <code>")
	return nil
}

// NewCmdProposalsView shows proposal details
func NewCmdProposalsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <code>",
		Short: "View proposal details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProposalsView(f, args[0])
		},
	}
}

func runProposalsView(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading proposal...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/proposals/%s", code)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	p := result.Data

	output.Header(fmt.Sprintf("Proposal %s", p.Code))

	output.KeyValue("Title", p.Title)
	clientName := p.ClientName
	if clientName == "" && p.ContactCompany != nil {
		clientName = p.ContactCompany.Name
	}
	output.KeyValue("Client", clientName)

	switch p.Status {
	case "approved":
		output.Success("Status: Approved")
	case "rejected":
		output.Error("Status: Rejected")
	case "sent", "viewed":
		output.Info("Status: " + p.StatusLabel)
	case "expired":
		output.Warning("Status: Expired")
	case "converted":
		output.Success("Status: Converted")
	default:
		output.KeyValue("Status", p.StatusLabel)
	}

	if p.TotalValueFormatted != "" {
		output.KeyValuef("Amount", "%s %s", p.Currency, p.TotalValueFormatted)
	}
	if p.ExpiresAt != nil {
		output.KeyValue("Expires", p.ExpiresAt.FormatDate())
	}

	fmt.Println()
	return nil
}

// NewCmdProposalsCreate creates a new proposal
func NewCmdProposalsCreate(f *factory.Factory) *cobra.Command {
	var clientSlug string
	var amount float64

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a new proposal",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProposalsCreate(f, args[0], clientSlug, amount)
		},
	}

	cmd.Flags().StringVar(&clientSlug, "client", "", "Client slug (required)")
	cmd.Flags().Float64Var(&amount, "amount", 0, "Proposal amount")
	cmd.MarkFlagRequired("client")

	return cmd
}

func runProposalsCreate(f *factory.Factory, title, clientSlug string, amount float64) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Creating proposal...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"title":       title,
		"client_slug": clientSlug,
	}
	if amount > 0 {
		body["amount"] = amount
	}

	path := "/proposals"
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Proposal created: %s", result.Data.Code)
	output.KeyValue("Title", result.Data.Title)
	createClientName := result.Data.ClientName
	if createClientName == "" && result.Data.ContactCompany != nil {
		createClientName = result.Data.ContactCompany.Name
	}
	output.KeyValue("Client", createClientName)

	return nil
}

// NewCmdProposalsSend sends a proposal
func NewCmdProposalsSend(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "send <code>",
		Short: "Send proposal to client",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProposalsSend(f, args[0])
		},
	}
}

func runProposalsSend(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Sending proposal...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/proposals/%s/send", code)
	_, err = client.Post(path, nil)
	sp.Stop()
	if err != nil {
		return err
	}

	output.Successf("Proposal %s sent to client", code)
	return nil
}

// NewCmdProposalsConvert converts proposal to project
func NewCmdProposalsConvert(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "convert <code>",
		Short: "Convert approved proposal to project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProposalsConvert(f, args[0])
		},
	}
}

func runProposalsConvert(f *factory.Factory, code string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Converting proposal...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/proposals/%s/convert-to-project", code)
	resp, err := client.Post(path, nil)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Project struct {
				Name string `json:"name"`
				Slug string `json:"slug"`
			} `json:"project"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Proposal converted to project: %s", result.Data.Project.Name)
	output.KeyValue("Slug", result.Data.Project.Slug)

	return nil
}
