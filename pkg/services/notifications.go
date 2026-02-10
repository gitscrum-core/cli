package services

import (
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// Notification represents a notification
type Notification struct {
	UUID      string    `json:"uuid"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
}

// NotificationsService handles notification operations
type NotificationsService interface {
	// List returns recent notifications
	List(unreadOnly bool) ([]Notification, error)

	// MarkAsRead marks a notification as read
	MarkAsRead(uuid string) error

	// MarkAllAsRead marks all notifications as read
	MarkAllAsRead() error
}

type notificationsService struct {
	client *api.Client
}

func newNotificationsService(client *api.Client) NotificationsService {
	return &notificationsService{client: client}
}

func (s *notificationsService) List(unreadOnly bool) ([]Notification, error) {
	path := "/notifications"
	if unreadOnly {
		path += "?unread=true"
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Notification `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *notificationsService) MarkAsRead(uuid string) error {
	resp, err := s.client.Post("/notifications/"+uuid+"/mark-read", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *notificationsService) MarkAllAsRead() error {
	resp, err := s.client.Post("/notifications/mark-all-read", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
