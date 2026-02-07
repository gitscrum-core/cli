// Package chat provides chat/discussions commands for GitScrum CLI
package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/gitscrum-core/cli/pkg/api"
	"github.com/gitscrum-core/cli/pkg/cmd/factory"
)

// NewCmdChat creates the chat command group
func NewCmdChat(f *factory.Factory) *cobra.Command {
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
				return runChatView(f, args[0])
			}
			return runChatSend(f, args[0], strings.Join(args[1:], " "))
		},
	}

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
		Content   string `json:"content"`
		CreatedAt string `json:"created_at"`
		User      struct {
			Name string `json:"name"`
		} `json:"user"`
	} `json:"last_message"`
}

// Message represents a chat message
type Message struct {
	UUID      string `json:"uuid"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Name   string `json:"name"`
		Avatar string `json:"avatar"`
	} `json:"user"`
}

func runChatList(f *factory.Factory) error {
	if err := f.RequireAuth(); err != nil {
		fmt.Println("error: not authenticated")
		fmt.Println("  run 'gitscrum auth login' to authenticate")
		return nil
	}

	workspace, _ := f.CurrentWorkspace()
	project, _ := f.CurrentProject()

	if workspace == "" {
		return fmt.Errorf("workspace required")
	}

	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := "/discussions/channels"
	if project != "" {
		path += "?project=" + project
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

	if len(result.Data) == 0 {
		fmt.Println("No channels found")
		return nil
	}

	fmt.Println("CHANNELS")
	fmt.Println()

	for _, c := range result.Data {
		unread := ""
		if c.UnreadCount > 0 {
			unread = fmt.Sprintf(" (%d unread)", c.UnreadCount)
		}
		fmt.Printf("  #%s%s\n", c.Name, unread)
		if c.Description != "" {
			fmt.Printf("    %s\n", c.Description)
		}
		if c.LastMessage.Content != "" {
			preview := c.LastMessage.Content
			if len(preview) > 50 {
				preview = preview[:50] + "..."
			}
			fmt.Printf("    %s: %s\n", c.LastMessage.User.Name, preview)
		}
		fmt.Println()
	}

	return nil
}

func runChatView(f *factory.Factory, channel string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	channel = strings.TrimPrefix(channel, "#")


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/discussions/channels/%s/messages", channel)
	resp, err := client.Get(path)
	if err != nil {
		return err
	}

	var result struct {
		Data []Message `json:"data"`
	}
	if err := api.DecodeResponse(resp, &result); err != nil {
		return err
	}

	fmt.Printf("#%s\n", channel)
	fmt.Println(strings.Repeat("-", 50))
	fmt.Println()

	if len(result.Data) == 0 {
		fmt.Println("  No messages yet")
		return nil
	}

	for i := len(result.Data) - 1; i >= 0; i-- {
		m := result.Data[i]
		timestamp := formatTimestamp(m.CreatedAt)
		fmt.Printf("  [%s] %s:\n", timestamp, m.User.Name)
		fmt.Printf("    %s\n\n", m.Content)
	}

	return nil
}

func runChatSend(f *factory.Factory, channel, message string) error {
	if err := f.RequireAuth(); err != nil {
		return err
	}

	channel = strings.TrimPrefix(channel, "#")


	client, err := f.APIClient()
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"content": message,
	}

	path := fmt.Sprintf("/discussions/channels/%s/messages", channel)
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

	fmt.Printf("Message sent to #%s\n", channel)
	return nil
}

func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	return t.Format("15:04")
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
		fmt.Println("No unread messages")
		return nil
	}

	fmt.Println("UNREAD MESSAGES")
	fmt.Println()

	for _, item := range result.Data {
		fmt.Printf("  #%s (%d new)\n", item.Channel.Name, len(item.Messages))
		for _, m := range item.Messages {
			preview := m.Content
			if len(preview) > 60 {
				preview = preview[:60] + "..."
			}
			fmt.Printf("    * %s: %s\n", m.User.Name, preview)
		}
		fmt.Println()
	}

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
