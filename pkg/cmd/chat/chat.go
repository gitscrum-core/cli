// Package chat provides chat/discussions commands for GitScrum CLI
package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
	"github.com/gitscrum-core/cli/pkg/output"
)

// NewCmdChat creates the chat command group
func NewCmdChat(f *factory.Factory) *cobra.Command {
	var page int
	var limit int

	cmd := &cobra.Command{
		Use:   "chat [channel] [message]",
		Short: "Team discussions and chat",
		Long: `View and participate in team discussions.

Without arguments, lists available channels.
With a channel name, shows recent messages.
With a channel and message, sends a message.`,
		Aliases: []string{"discuss", "discussions"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runChatList(f)
			}
			if len(args) == 1 {
				return runChatView(f, args[0], page, limit)
			}
			return runChatSend(f, args[0], strings.Join(args[1:], " "))
		},
	}

	cmd.Flags().IntVar(&page, "page", 1, "Page number (1 = most recent)")
	cmd.Flags().IntVar(&limit, "limit", 30, "Messages per page (max 100)")

	cmd.AddCommand(NewCmdChatUnread(f))
	cmd.AddCommand(NewCmdChatSend(f))
	cmd.AddCommand(NewCmdChatChannels(f))

	return cmd
}

// Channel represents a chat channel
type Channel struct {
	UUID        string `json:"uuid"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	UnreadCount int    `json:"unread_count"`
	LastMessage struct {
		Content   string            `json:"content"`
		CreatedAt *api.DateResource `json:"created_at"`
		User      struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"last_message"`
}

// Message represents a chat message
type Message struct {
	UUID      string            `json:"uuid"`
	Content   string            `json:"content"`
	CreatedAt *api.DateResource `json:"created_at"`
	User      struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

func runChatList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}
	project, _ := f.CurrentProject()

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/discussions/channels?company_slug=%s", workspace)
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Channel `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	// JSON output if requested
	if f.OutputFormat == output.FormatJSON {
		return f.Formatter().Print(result.Data)
	}

	if len(result.Data) == 0 {
		ws, _ := f.CurrentWorkspace()
		pj, _ := f.CurrentProject()
		hint := "Create a channel in your project settings"
		if ws != "" && pj != "" {
			hint = fmt.Sprintf("Create a channel in %s → %s", ws, pj)
		}
		output.Empty("No channels found", hint)
		return nil
	}

	output.Header("Channels")

	for _, c := range result.Data {
		if c.UnreadCount > 0 {
			output.Warningf("#%s (%d unread)", c.Name, c.UnreadCount)
		} else {
			output.Infof("#%s", c.Name)
		}
		if c.Description != "" {
			output.Dim(c.Description)
		}
		if c.LastMessage.Content != "" {
			preview := output.Truncate(c.LastMessage.Content, 50)
			output.Dimf("%s: %s", c.LastMessage.User.Name, preview)
		}
	}

	fmt.Println()
	return nil
}

func runChatView(f *factory.Factory, channelNameOrSlug string, page, limit int) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}
	project, _ := f.CurrentProject()

	channelNameOrSlug = strings.TrimPrefix(channelNameOrSlug, "#")

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// First, fetch channels to resolve name/slug to UUID
	channelUUID, channelName, err := resolveChannel(client, workspace, project, channelNameOrSlug)
	if err != nil {
		return err
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	// Calculate offset for pagination (page 1 = most recent)
	// We need to fetch more messages for older pages
	// Since API uses cursor-based pagination, we'll fetch all needed and slice
	// For simplicity, we'll request limit messages per page
	offset := (page - 1) * limit

	// Build path with limit - we request enough to cover the page
	// Note: API uses cursor pagination, so we request limit+offset and slice
	totalNeeded := offset + limit
	path := fmt.Sprintf("/discussions/channels/%s/messages?company_slug=%s&limit=%d", channelUUID, workspace, totalNeeded)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Message `json:"data"`
		Meta struct {
			HasMore  bool `json:"has_more"`
			OldestID int  `json:"oldest_id"`
			NewestID int  `json:"newest_id"`
		} `json:"meta"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	// JSON output if requested
	if f.OutputFormat == output.FormatJSON {
		// Slice messages for the current page
		totalMessages := len(result.Data)
		startIdx := offset
		endIdx := offset + limit
		if startIdx >= totalMessages {
			startIdx = totalMessages
		}
		if endIdx > totalMessages {
			endIdx = totalMessages
		}

		pageMessages := result.Data[startIdx:endIdx]
		jsonOutput := struct {
			Channel  string    `json:"channel"`
			Page     int       `json:"page"`
			Limit    int       `json:"limit"`
			Total    int       `json:"total"`
			HasMore  bool      `json:"has_more"`
			Messages []Message `json:"messages"`
		}{
			Channel:  channelName,
			Page:     page,
			Limit:    limit,
			Total:    totalMessages,
			HasMore:  result.Meta.HasMore || endIdx < totalMessages,
			Messages: pageMessages,
		}
		return f.Formatter().Print(jsonOutput)
	}

	// Header with page info
	if page > 1 {
		output.Header(fmt.Sprintf("#%s (Page %d)", channelName, page))
	} else {
		output.Header(fmt.Sprintf("#%s", channelName))
	}

	if len(result.Data) == 0 {
		output.Empty("No messages yet", "Send one with: gitscrum chat send "+channelNameOrSlug+" \"Hello!\"")
		return nil
	}

	// Slice messages for the current page
	// result.Data is ordered newest first from API
	totalMessages := len(result.Data)

	// Calculate start and end indices for this page
	startIdx := offset
	endIdx := offset + limit
	if startIdx >= totalMessages {
		output.Empty("No more messages", fmt.Sprintf("Page %d has no messages. Try: gitscrum chat %s --page %d", page, channelNameOrSlug, page-1))
		return nil
	}
	if endIdx > totalMessages {
		endIdx = totalMessages
	}

	// Get messages for this page (reversed to show oldest first in the page)
	pageMessages := result.Data[startIdx:endIdx]
	for i := len(pageMessages) - 1; i >= 0; i-- {
		m := pageMessages[i]
		timestamp := formatTimestamp(m.CreatedAt)
		output.Infof("[%s] %s:", timestamp, m.User.Name)
		fmt.Printf("    %s\n\n", output.StripHTML(m.Content))
	}

	// Show pagination info
	hasOlderMessages := result.Meta.HasMore || endIdx < totalMessages
	fmt.Println("─────────────────────────────────────")
	fmt.Printf("Page %d • Showing %d messages\n", page, len(pageMessages))

	if page > 1 {
		fmt.Printf("Newer: gitscrum chat %s --page %d\n", channelNameOrSlug, page-1)
	}
	if hasOlderMessages {
		fmt.Printf("Older: gitscrum chat %s --page %d\n", channelNameOrSlug, page+1)
	}
	fmt.Println()

	return nil
}

func runChatSend(f *factory.Factory, channelNameOrSlug, message string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	workspace, err := f.RequireWorkspace()
	if err != nil {
		return err
	}
	project, _ := f.CurrentProject()

	channelNameOrSlug = strings.TrimPrefix(channelNameOrSlug, "#")

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	// First, fetch channels to resolve name/slug to UUID
	channelUUID, channelName, err := resolveChannel(client, workspace, project, channelNameOrSlug)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"content": message,
	}

	path := fmt.Sprintf("/discussions/channels/%s/messages?company_slug=%s", channelUUID, workspace)
	if project != "" {
		path += "&project_slug=" + project
	}
	resp, err := client.Post(path, body)
	if err != nil {
		return err
	}

	var result struct {
		Data Message `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	output.Successf("Message sent to #%s", channelName)
	return nil
}

// resolveChannel fetches channels and resolves a name or slug to the channel UUID
// Returns (uuid, name, error)
func resolveChannel(client *api.Client, workspace, project, channelNameOrSlug string) (string, string, error) {
	path := fmt.Sprintf("/discussions/channels?company_slug=%s", workspace)
	if project != "" {
		path += "&project_slug=" + project
	}

	resp, err := client.Get(path)
	if err != nil {
		return "", "", err
	}

	var result struct {
		Data []Channel `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return "", "", err
	}

	// Try to match by name (case-insensitive), slug, or UUID
	nameToMatch := strings.ToLower(channelNameOrSlug)
	for _, c := range result.Data {
		if strings.ToLower(c.Name) == nameToMatch ||
			strings.ToLower(c.Slug) == nameToMatch ||
			c.UUID == channelNameOrSlug {
			return c.UUID, c.Name, nil
		}
	}

	return "", "", fmt.Errorf("channel '%s' not found", channelNameOrSlug)
}

func formatTimestamp(d *api.DateResource) string {
	if d == nil {
		return ""
	}
	if d.ISO8601 != "" {
		t, err := time.Parse(time.RFC3339, d.ISO8601)
		if err == nil {
			return t.Format("15:04")
		}
	}
	return d.FormatDate()
}

// NewCmdChatUnread shows unread messages
func NewCmdChatUnread(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "unread",
		Short: "Show unread messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChatUnread(f)
		},
	}
}

func runChatUnread(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/discussions/channels?unread=true"
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []struct {
			Channel  Channel   `json:"channel"`
			Messages []Message `json:"messages"`
		} `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	if len(result.Data) == 0 {
		output.Success("No unread messages")
		return nil
	}

	output.Header("Unread Messages")

	for _, item := range result.Data {
		output.Warningf("#%s (%d new)", item.Channel.Name, len(item.Messages))
		for _, m := range item.Messages {
			preview := output.Truncate(m.Content, 60)
			output.Bulletf("%s: %s", m.User.Name, preview)
		}
	}

	fmt.Println()
	return nil
}

// NewCmdChatSend sends a message
func NewCmdChatSend(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:   "send <channel> <message>",
		Short: "Send a message to a channel",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChatSend(f, args[0], strings.Join(args[1:], " "))
		},
	}
}

// NewCmdChatChannels lists channels
func NewCmdChatChannels(f *factory.Factory) *cobra.Command {
	return &cobra.Command{
		Use:     "channels",
		Short:   "List all channels",
		Aliases: []string{"list"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChatList(f)
		},
	}
}
