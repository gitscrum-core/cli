package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// ChatMessage represents a chat message
type ChatMessage struct {
	UUID      string    `json:"uuid"`
	Content   string    `json:"content"`
	Author    User      `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

// ChatChannel represents a chat channel
type ChatChannel struct {
	UUID        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // project, direct, team
}

// ChatService handles chat operations
type ChatService interface {
	// ListChannels returns available channels
	ListChannels() ([]ChatChannel, error)

	// ListMessages returns messages from a channel
	ListMessages(channelID string, limit int) ([]ChatMessage, error)

	// SendMessage sends a message to a channel
	SendMessage(channelID, content string) (*ChatMessage, error)
}

type chatService struct {
	client *api.Client
}

func newChatService(client *api.Client) ChatService {
	return &chatService{client: client}
}

func (s *chatService) ListChannels() ([]ChatChannel, error) {
	resp, err := s.client.Get("/chat/channels")
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []ChatChannel `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *chatService) ListMessages(channelID string, limit int) ([]ChatMessage, error) {
	path := "/chat/channels/" + channelID + "/messages"
	if limit > 0 {
		path += "?limit=" + string(rune(limit))
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []ChatMessage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *chatService) SendMessage(channelID, content string) (*ChatMessage, error) {
	body := map[string]interface{}{
		"content": content,
	}

	resp, err := s.client.Post("/chat/channels/"+channelID+"/messages", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data ChatMessage `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}
