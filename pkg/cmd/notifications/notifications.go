// Package notifications provides notification commands for GitScrum CLI
package notifications

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/i18n"
	"github.com/gitscrum-core/cli/pkg/output"
	"github.com/gitscrum-core/cli/pkg/spinner"
)

// NewCmdNotifications creates the notifications command
func NewCmdNotifications(f *factory.Factory) *cobra.Command {
	var unreadOnly bool

	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "View notifications",
		Long: `View your notifications.

Use --unread to show only unread notifications.`,
		Aliases: []string{"notif", "n"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotifications(f, unreadOnly)
		},
	}

	cmd.Flags().BoolVar(&unreadOnly, "unread", false, "Show only unread notifications")

	cmd.AddCommand(NewCmdNotificationsRead(f))
	cmd.AddCommand(NewCmdNotificationsReadAll(f))
	cmd.AddCommand(NewCmdNotificationsClear(f))

	return cmd
}

// FeedUser represents a user in a feed notification
type FeedUser struct {
	UUID     string `json:"uuid"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

// Notification represents a feed notification (FeedUserResource)
type Notification struct {
	UUID      string            `json:"uuid"`
	User      FeedUser          `json:"user"`
	Resource  string            `json:"resource"`
	Action    string            `json:"action"`
	Message   string            `json:"message"`
	CreatedAt *api.DateResource `json:"created_at"`
}

func runNotifications(f *factory.Factory, unreadOnly bool) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Loading notifications...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := "/feeds/notifications?page=1"
	if unreadOnly {
		path += "&unread=true"
	}

	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data []Notification `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		if unreadOnly {
			output.Success(i18n.T("no_unread_notifications"))
		} else {
			workspace, _ := f.CurrentWorkspace()
			output.EmptyContext(i18n.T("no_notifications"), workspace, "", "")
		}
		return nil
	}

	title := "Notifications"
	if unreadOnly {
		title = "Unread Notifications"
	}
	output.Header(title)

	for _, n := range result.Data {
		timestamp := ""
		if n.CreatedAt != nil {
			timestamp = n.CreatedAt.FormatDate()
		}

		output.Warningf("● %s", n.Message)

		meta := ""
		if n.User.Name != "" {
			meta = n.User.Name
		}
		if n.Action != "" {
			if meta != "" {
				meta += " • "
			}
			meta += n.Action
		}
		if timestamp != "" {
			if meta != "" {
				meta += " • "
			}
			meta += timestamp
		}
		if meta != "" {
			output.Dim("  " + meta)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdNotificationsRead marks a notification as read
func NewCmdNotificationsRead(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "read <id>",
		Short: "Mark notification as read",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotificationsRead(f, args[0])
		},
	}
}

func runNotificationsRead(f *factory.Factory, id string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Marking as read...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/feeds/notifications/%s/read", id)
	_, err = client.Post(path, nil)
	sp.Stop()
	if err != nil {
		return err
	}

	output.Success("Notification marked as read")
	return nil
}

// NewCmdNotificationsReadAll marks all notifications as read
func NewCmdNotificationsReadAll(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "read-all",
		Short: "Mark all notifications as read",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotificationsReadAll(f)
		},
	}
}

func runNotificationsReadAll(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Marking all as read...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	_, err = client.Post("/feeds/notifications/read-all", nil)
	sp.Stop()
	if err != nil {
		return err
	}

	output.Success("All notifications marked as read")
	return nil
}

// NewCmdNotificationsClear clears all notifications
func NewCmdNotificationsClear(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all notifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNotificationsClear(f)
		},
	}
}

func runNotificationsClear(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Clearing notifications...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	_, err = client.Delete("/feeds/notifications")
	sp.Stop()
	if err != nil {
		return err
	}

	output.Success("All notifications cleared")
	return nil
}

// NewCmdSearch creates the global search command
func NewCmdSearch(f *factory.Factory) *cobra.Command {
	var scope string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Global search across GitScrum",
		Long: `Search for tasks, projects, wiki pages, and more.

Use --scope to limit search to a specific type.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSearch(f, args[0], scope)
		},
	}

	cmd.Flags().StringVar(&scope, "scope", "", "Limit search to: tasks, projects, wiki, users")

	return cmd
}

func runSearch(f *factory.Factory, query, scope string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	sp := spinner.New("Searching...")
	sp.Start()

	client, err := f.APIClient()
	if err != nil {
		sp.Stop()
		return err
	}

	path := fmt.Sprintf("/search?q=%s", query)
	if scope != "" {
		path += "&scope=" + scope
	}

	resp, err := client.Get(path)
	sp.Stop()
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			Tasks []struct {
				Code  string `json:"code"`
				Title string `json:"title"`
			} `json:"tasks"`
			Projects []struct {
				Slug string `json:"slug"`
				Name string `json:"name"`
			} `json:"projects"`
			Wiki []struct {
				Slug  string `json:"slug"`
				Title string `json:"title"`
			} `json:"wiki"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	output.Header(fmt.Sprintf("Search results for \"%s\"", query))

	hasResults := false

	if len(result.Data.Tasks) > 0 {
		hasResults = true
		output.SubHeader("Tasks")
		for _, t := range result.Data.Tasks {
			output.Bulletf("[%s] %s", t.Code, t.Title)
		}
	}

	if len(result.Data.Projects) > 0 {
		hasResults = true
		output.SubHeader("Projects")
		for _, p := range result.Data.Projects {
			output.Bulletf("%s (%s)", p.Name, p.Slug)
		}
	}

	if len(result.Data.Wiki) > 0 {
		hasResults = true
		output.SubHeader("Wiki")
		for _, w := range result.Data.Wiki {
			output.Bulletf("📄 %s (%s)", w.Title, w.Slug)
		}
	}

	if !hasResults {
		output.Empty("No results found", "")
	}

	fmt.Println()
	return nil
}
