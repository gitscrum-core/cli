// Package proposals provides proposal commands for GitScrum CLI
package proposals

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
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

// Proposal represents a proposal
type Proposal struct {
	UUID   string  `json:"uuid"`
	Code   string  `json:"code"`
	Title  string  `json:"title"`
	Status string  `json:"status"`
	Amount float64 `json:"amount"`
	Client struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"client"`
	ExpiresAt string `json:"expires_at"`
}

func runProposalsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		return nil
	}


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/proposals"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No proposals found")
		return nil
	}

	fmt.Println("PROPOSALS")
	fmt.Println()

	for _, p := range result.Data {
		status := getProposalStatusIndicator(p.Status)
		fmt.Printf("  %s %s - %s\n", status, p.Code, p.Title)
		fmt.Printf("     Client: %s\n", p.Client.Name)
		if p.Amount > 0 {
			fmt.Printf("     Amount: $%.2f\n", p.Amount)
		}
		fmt.Printf("     Status: %s\n", p.Status)
		fmt.Println()
	}

	return nil
}

func getProposalStatusIndicator(status string) string {
	switch status {
	case "accepted":
		return "[OK]"
	case "sent":
		return "[->]"
	case "declined":
		return "[X]"
	case "draft":
		return "[D]"
	case "expired":
		return "[!]"
	default:
		return "[-]"
	}
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/proposals/%s", code)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	p := result.Data

	fmt.Printf("PROPOSAL %s\n", p.Code)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("\nTitle: %s\n", p.Title)
	fmt.Printf("Client: %s\n", p.Client.Name)
	fmt.Printf("Status: %s %s\n", getProposalStatusIndicator(p.Status), p.Status)
	if p.Amount > 0 {
		fmt.Printf("Amount: $%.2f\n", p.Amount)
	}
	if p.ExpiresAt != "" {
		fmt.Printf("Expires: %s\n", p.ExpiresAt)
	}

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


	client, err := f.APIClient()
	if err != nil {
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
	if err != nil {
		return err
	}

	var result struct {
		Data Proposal `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Proposal created: %s\n", result.Data.Code)
	fmt.Printf("  Title: %s\n", result.Data.Title)
	fmt.Printf("  Client: %s\n", result.Data.Client.Name)

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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/proposals/%s/send", code)
	_, err = client.Post(path, nil)
	if err != nil {
		return err
	}

	fmt.Printf("Proposal %s sent to client\n", code)
	return nil
}

// NewCmdProposalsConvert converts proposal to project
func NewCmdProposalsConvert(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "convert <code>",
		Short: "Convert accepted proposal to project",
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/proposals/%s/convert-to-project", code)
	resp, err := client.Post(path, nil)
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

	fmt.Printf("Proposal converted to project: %s\n", result.Data.Project.Name)
	fmt.Printf("  Project slug: %s\n", result.Data.Project.Slug)

	return nil
}
