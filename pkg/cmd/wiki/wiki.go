// Package wiki provides wiki commands for GitScrum CLI
package wiki

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdWiki creates the wiki command group
func NewCmdWiki(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wiki [command]",
		Short: "Project wiki and documentation",
		Long: `View and manage wiki pages.

Without a subcommand, lists wiki pages.`,
		Aliases: []string{"docs", "doc"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWikiList(f)
		},
	}

	cmd.AddCommand(NewCmdWikiView(f))
	cmd.AddCommand(NewCmdWikiCreate(f))
	cmd.AddCommand(NewCmdWikiSearch(f))
	cmd.AddCommand(NewCmdWikiEdit(f))

	return cmd
}

// Page represents a wiki page
type Page struct {
	UUID      string `json:"uuid"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	UpdatedAt string `json:"updated_at"`
	Author    struct {
		Name string `json:"name"`
	} `json:"author"`
	Children []Page `json:"children,omitempty"`
}

func runWikiList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}



	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/wiki/pages"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Println("No wiki pages yet")
		fmt.Println()
		fmt.Println("Create one with: gitscrum wiki create \"Getting Started\"")
		return nil
	}

	fmt.Println("WIKI PAGES:")
	fmt.Println()

	printPages(result.Data, 0)

	return nil
}

func printPages(pages []Page, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, p := range pages {
		fmt.Printf("%s- %s\n", prefix, p.Title)
		fmt.Printf("%s   slug: %s\n", prefix, p.Slug)
		if len(p.Children) > 0 {
			printPages(p.Children, indent+1)
		}
	}
}

// NewCmdWikiView views a wiki page
func NewCmdWikiView(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "view <slug>",
		Short: "View a wiki page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWikiView(f, args[0])
		},
	}
}

func runWikiView(f *factory.Factory, slug string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/wiki/pages/%s", slug)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	p := result.Data

	fmt.Printf("PAGE: %s\n", p.Title)
	fmt.Println(strings.Repeat("─", 60))
	fmt.Println()
	fmt.Println(p.Content)
	fmt.Println()
	fmt.Println(strings.Repeat("─", 60))
	fmt.Printf("Last updated by %s\n", p.Author.Name)

	return nil
}

// NewCmdWikiCreate creates a wiki page
func NewCmdWikiCreate(f *factory.Factory) *cobra.Command {
	var file, parent string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a wiki page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWikiCreate(f, args[0], file, parent)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Read content from file")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent page slug")

	return cmd
}

func runWikiCreate(f *factory.Factory, title, file, parent string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	var content string
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		content = string(data)
	} else {
		content = "# " + title + "\n\nContent goes here..."
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"title":   title,
		"content": content,
	}
	if parent != "" {
		body["parent_slug"] = parent
	}

	path := "/wiki/pages"
	resp, err := client.Post(path, body)
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Wiki page created: %s\n", result.Data.Title)
	fmt.Printf("  Slug: %s\n", result.Data.Slug)
	fmt.Printf("  View: gitscrum wiki view %s\n", result.Data.Slug)

	return nil
}

// NewCmdWikiSearch searches wiki pages
func NewCmdWikiSearch(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "search <query>",
		Short: "Search wiki pages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWikiSearch(f, args[0])
		},
	}
}

func runWikiSearch(f *factory.Factory, query string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/wiki/pages/search?q=%s", query)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		fmt.Printf("No wiki pages found for \"%s\"\n", query)
		return nil
	}

	fmt.Printf("Search results for \"%s\":\n\n", query)

	for _, p := range result.Data {
		fmt.Printf("  - %s\n", p.Title)
		fmt.Printf("     Slug: %s\n\n", p.Slug)
	}

	return nil
}

// NewCmdWikiEdit edits a wiki page
func NewCmdWikiEdit(f *factory.Factory) *cobra.Command {
	var file string

	cmd := &cobra.Command{
		Use:   "edit <slug>",
		Short: "Edit a wiki page",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWikiEdit(f, args[0], file)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "Read new content from file")
	cmd.MarkFlagRequired("file")

	return cmd
}

func runWikiEdit(f *factory.Factory, slug, file string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	body := map[string]interface{}{
		"content": string(content),
	}

	path := fmt.Sprintf("/wiki/pages/%s", slug)
	resp, err := client.Patch(path, body)
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("Wiki page updated: %s\n", result.Data.Title)

	return nil
}
