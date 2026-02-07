// Package notifications provides notification commands for GitScrum CLI
package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
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

// Notification represents a notification
type Notification struct {
	UUID      string `json:"uuid"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Body      string `json:"body"`
	ReadAt    string `json:"read_at"`
	CreatedAt string `json:"created_at"`
	Data      struct {
		TaskCode    string `json:"task_code,omitempty"`
		ProjectSlug string `json:"project_slug,omitempty"`
	} `json:"data"`
}

func runNotifications(f *factory.Factory, unreadOnly bool) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  Run 'gitscrum auth login' to authenticate")
		return nil
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/notifications"
	if unreadOnly {
		path += "?unread=true"
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Notification `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		if unreadOnly {
			fmt.Println("No unread notifications")
		} else {
			fmt.Println("No notifications")
		}
		return nil
	}

	title := "NOTIFICATIONS"
	if unreadOnly {
		title = "UNREAD NOTIFICATIONS"
	}
	fmt.Println(title)
	fmt.Println()

	for _, n := range result.Data {
		icon := "*"
		if n.ReadAt != "" {
			icon = "-"
		}
		
		timestamp := formatRelativeTime(n.CreatedAt)
		
		fmt.Printf("  %s %s\n", icon, n.Title)
		if n.Body != "" {
			body := n.Body
			if len(body) > 70 {
				body = body[:70] + "..."
			}
			fmt.Printf("     %s\n", body)
		}
		fmt.Printf("     %s", timestamp)
		if n.Data.TaskCode != "" {
			fmt.Printf(" • %s", n.Data.TaskCode)
		}
		fmt.Println()
		fmt.Println()
	}

	return nil
}

func formatRelativeTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	
	diff := time.Since(t)
	
	if diff < time.Minute {
		return "just now"
	}
	if diff < time.Hour {
		mins := int(diff.Minutes())
		return fmt.Sprintf("%dm ago", mins)
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		return fmt.Sprintf("%dh ago", hours)
	}
	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		return fmt.Sprintf("%dd ago", days)
	}
	
	return t.Format("Jan 2")
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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/notifications/%s/read", id)
	_, err = client.Post(path, nil)
	if err != nil {
		return err
	}

	fmt.Println("Notification marked as read")
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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	_, err = client.Post("/notifications/read-all", nil)
	if err != nil {
		return err
	}

	fmt.Println("All notifications marked as read")
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

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	_, err = client.Delete("/notifications")
	if err != nil {
		return err
	}

	fmt.Println("All notifications cleared")
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


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/search?q=%s", query)
	if scope != "" {
		path += "&scope=" + scope
	}

	resp, err := client.Get(path)
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

	fmt.Printf("Search results for \"%s\":\n", query)
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	hasResults := false

	if len(result.Data.Tasks) > 0 {
		hasResults = true
		fmt.Println("Tasks:")
		for _, t := range result.Data.Tasks {
			fmt.Printf("   [%s] %s\n", t.Code, t.Title)
		}
		fmt.Println()
	}

	if len(result.Data.Projects) > 0 {
		hasResults = true
		fmt.Println("Projects:")
		for _, p := range result.Data.Projects {
			fmt.Printf("   %s (%s)\n", p.Name, p.Slug)
		}
		fmt.Println()
	}

	if len(result.Data.Wiki) > 0 {
		hasResults = true
		fmt.Println("Wiki:")
		for _, w := range result.Data.Wiki {
			fmt.Printf("   %s (%s)\n", w.Title, w.Slug)
		}
		fmt.Println()
	}

	if !hasResults {
		fmt.Println("  No results found")
	}

	return nil
}
