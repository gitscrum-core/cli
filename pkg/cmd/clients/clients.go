// Package clients provides client CRM commands for GitScrum CLI
package clients

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

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

// Client represents a client
type Client struct {
	UUID         string  `json:"uuid"`
	Slug         string  `json:"slug"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	Company      string  `json:"company"`
	Status       string  `json:"status"`
	TotalRevenue float64 `json:"total_revenue"`
	Projects     int     `json:"projects_count"`
}

func runClientsList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/contact-companies/clients"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No clients found")
		fmt.Println()
		fmt.Println("Add one with: gitscrum clients create \"Client Name\"")
		return nil
	}

	fmt.Println("CLIENTS:")
	fmt.Println()

	for _, c := range result.Data {
		status := getStatusIcon(c.Status)
		fmt.Printf("  %s %s\n", status, c.Name)
		if c.Company != "" {
			fmt.Printf("     Company: %s\n", c.Company)
		}
		if c.Email != "" {
			fmt.Printf("     %s\n", c.Email)
		}
		fmt.Printf("     %d projects", c.Projects)
		if c.TotalRevenue > 0 {
			fmt.Printf(" | $%.2f", c.TotalRevenue)
		}
		fmt.Println()
		fmt.Println()
	}

	return nil
}

func getStatusIcon(status string) string {
	switch status {
	case "active":
		return "[active]"
	case "at-risk":
		return "[at-risk]"
	case "churned":
		return "[churned]"
	case "prospect":
		return "[prospect]"
	default:
		return "[-]"
	}
}

// NewCmdClientsView shows client details
func NewCmdClientsView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <slug>",
		Short: "View client details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClientsView(f, args[0])
		},
	}
}

func runClientsView(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/contact-companies/%s", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	c := result.Data

	fmt.Printf("%s\n", c.Name)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Printf("\nStatus: %s %s\n", getStatusIcon(c.Status), c.Status)
	if c.Company != "" {
		fmt.Printf("Company: %s\n", c.Company)
	}
	if c.Email != "" {
		fmt.Printf("Email: %s\n", c.Email)
	}
	if c.Phone != "" {
		fmt.Printf("Phone: %s\n", c.Phone)
	}
	fmt.Printf("\nProjects: %d\n", c.Projects)
	fmt.Printf("Total Revenue: $%.2f\n", c.TotalRevenue)

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


	client, err := f.APIClient()
	if err != nil {
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
	if err != nil {
		return err
	}

	var result struct {
		Data Client `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Client created: %s\n", result.Data.Name)
	fmt.Printf("  Slug: %s\n", result.Data.Slug)

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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/contact-companies/stats"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			TotalClients  int     `json:"total_clients"`
			ActiveClients int     `json:"active_clients"`
			AtRisk        int     `json:"at_risk"`
			Churned       int     `json:"churned"`
			TotalRevenue  float64 `json:"total_revenue"`
			MRR           float64 `json:"mrr"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	s := result.Data

	fmt.Println("CLIENT STATISTICS")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()
	fmt.Printf("Total Clients:  %d\n", s.TotalClients)
	fmt.Printf("Active:         %d\n", s.ActiveClients)
	fmt.Printf("At Risk:        %d\n", s.AtRisk)
	fmt.Printf("Churned:        %d\n", s.Churned)
	fmt.Println()
	fmt.Printf("Total Revenue:  $%.2f\n", s.TotalRevenue)
	fmt.Printf("MRR:            $%.2f\n", s.MRR)

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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/projects?contact_company_uuid=%s", slug)
	resp, err := client.Get(path)
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

	fmt.Printf("PROJECTS FOR %s:\n\n", slug)

	if len(result.Data) == 0 {
		fmt.Println("  No projects found")
		return nil
	}

	for _, p := range result.Data {
		fmt.Printf("  • %s (%s)\n", p.Name, p.Slug)
		if p.Budget > 0 {
			fmt.Printf("    Budget: $%.2f\n", p.Budget)
		}
	}

	return nil
}
