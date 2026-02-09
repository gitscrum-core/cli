package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// AnalyticsData represents analytics metrics
type AnalyticsData struct {
	TasksCreated   int               `json:"tasks_created"`
	TasksCompleted int               `json:"tasks_completed"`
	TimeSpent      float64           `json:"time_spent"`
	Velocity       float64           `json:"velocity"`
	ByStatus       map[string]int    `json:"by_status"`
	ByMember       []MemberAnalytics `json:"by_member"`
}

// MemberAnalytics represents per-user analytics
type MemberAnalytics struct {
	User           User    `json:"user"`
	TasksCompleted int     `json:"tasks_completed"`
	TimeSpent      float64 `json:"time_spent"`
}

// AnalyticsService handles analytics operations
type AnalyticsService interface {
	// ProjectAnalytics returns analytics for a project
	ProjectAnalytics(project string, period string) (*AnalyticsData, error)

	// TeamAnalytics returns team-wide analytics
	TeamAnalytics(period string) (*AnalyticsData, error)
}

type analyticsService struct {
	client *api.Client
}

func newAnalyticsService(client *api.Client) AnalyticsService {
	return &analyticsService{client: client}
}

func (s *analyticsService) ProjectAnalytics(project, period string) (*AnalyticsData, error) {
	path := "/analytics/projects/" + project + "?period=" + period

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data AnalyticsData `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *analyticsService) TeamAnalytics(period string) (*AnalyticsData, error) {
	path := "/analytics/team?period=" + period

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data AnalyticsData `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

