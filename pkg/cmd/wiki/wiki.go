// Package wiki provides wiki commands for GitScrum CLI
package wiki

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
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

// Page represents a wiki page (matches WikiResource.php)
type Page struct {
	UUID     string `json:"uuid"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Page     string `json:"page"`
	Children []Page `json:"children,omitempty"`
}

func runWikiList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, project, err := f.RequireWorkspaceAndProject()
	if err != nil {
		return err
	}

	sp := spinner.New("Loading wiki pages...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/wiki/pages?company_slug=%s&project_slug=%s", workspace, project)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		workspace, _ := f.CurrentWorkspace()
		project, _ := f.CurrentProject()
		output.EmptyContext(i18n.T("no_wiki_pages"), workspace, project, i18n.T("create_wiki_hint"))
		return nil
	}

	output.Header("Wiki Pages")

	printPages(result.Data, 0)

	fmt.Println()
	return nil
}

func printPages(pages []Page, indent int) {
	prefix := strings.Repeat("  ", indent)
	for _, p := range pages {
		fmt.Printf("%s  📄 %s\n", prefix, p.Title)
		output.Dimf("%s     slug: %s", prefix, p.Slug)
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

	sp := spinner.New("Loading wiki page...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/wiki/pages/%s", slug)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	p := result.Data

	output.Header(p.Title)

	if p.Page != "" {
		fmt.Println(p.Page)
	}

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

	sp := spinner.New("Creating wiki page...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	body := map[string]interface{}{
		"title": title,
		"page":  content,
	}
	if parent != "" {
		body["parent_slug"] = parent
	}

	path := "/wiki/pages"
	resp, err := client.Post(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Wiki page created: %s", result.Data.Title)
	output.KeyValue("Slug", result.Data.Slug)
	output.Infof("View: gitscrum wiki view %s", result.Data.Slug)

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

	sp := spinner.New("Searching wiki...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/wiki/pages/search?q=%s", query)
	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		output.Empty(fmt.Sprintf("No wiki pages found for \"%s\"", query), "")
		return nil
	}

	output.Header(fmt.Sprintf("Search results for \"%s\"", query))

	for _, p := range result.Data {
		output.Bulletf("📄 %s", p.Title)
		output.Dimf("   slug: %s", p.Slug)
	}

	fmt.Println()
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

	sp := spinner.New("Updating wiki page...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	content, err := os.ReadFile(file)
	if err != nil {
		sp.Stop()
		return fmt.Errorf("failed to read file: %w", err)
	}

	body := map[string]interface{}{
		"page": string(content),
	}

	path := fmt.Sprintf("/wiki/pages/%s", slug)
	resp, err := client.Patch(path, body)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data Page `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Wiki page updated: %s", result.Data.Title)

	return nil
}
