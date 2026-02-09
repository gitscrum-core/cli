package services

import (
	"github.com/gitscrum-core/cli/pkg/api"
)

// TeamStandupSummary represents the standup summary stats
type TeamStandupSummary struct {
	DoneTasks    int     `json:"done_tasks"`
	ActiveTasks  int     `json:"active_tasks"`
	BlockedTasks int     `json:"blocked_tasks"`
	TrackedHours float64 `json:"tracked_hours"`
}

// StandupTask represents a task in standup context
type StandupTask struct {
	UUID      string `json:"uuid"`
	Code      string `json:"code"`
	Title     string `json:"title"`
	Status    string `json:"workflow_name"`
	Assignees []User `json:"users"`
}

// TeamMemberStatus represents a team member's standup status
type TeamMemberStatus struct {
	User          User    `json:"user"`
	ActiveTasks   int     `json:"active_tasks"`
	DoneTasks     int     `json:"done_tasks"`
	BlockedTasks  int     `json:"blocked_tasks"`
	TrackedHours  float64 `json:"tracked_hours"`
	IsOnline      bool    `json:"is_online"`
}

// WeeklyDigest represents weekly standup metrics
type WeeklyDigest struct {
	Days []DayDigest `json:"days"`
}

// DayDigest represents a single day's metrics
type DayDigest struct {
	Date         string `json:"date"`
	DoneTasks    int    `json:"done_count"`
	ActiveTasks  int    `json:"active_count"`
	BlockedTasks int    `json:"blocked_count"`
}

// Contributor represents a contributor's activity
type Contributor struct {
	User         User    `json:"user"`
	TasksDone    int     `json:"tasks_done"`
	TrackedHours float64 `json:"tracked_hours"`
}

// StandupService handles team standup operations
type StandupService interface {
	// Summary returns aggregated standup stats for workspace
	Summary(workspaceSlug string) (*TeamStandupSummary, error)

	// CompletedYesterday returns tasks completed yesterday
	CompletedYesterday(workspaceSlug string) ([]StandupTask, error)

	// Blockers returns current blocking tasks
	Blockers(workspaceSlug string) ([]StandupTask, error)

	// TeamStatus returns team members' status
	TeamStatus(workspaceSlug string) ([]TeamMemberStatus, error)

	// WeeklyDigest returns weekly metrics
	WeeklyDigest(workspaceSlug string) (*WeeklyDigest, error)

	// Contributors returns contributors for period
	Contributors(workspaceSlug, period string) ([]Contributor, error)
}

type standupService struct {
	client *api.Client
}

func newStandupService(client *api.Client) StandupService {
	return &standupService{client: client}
}

func (s *standupService) Summary(workspaceSlug string) (*TeamStandupSummary, error) {
	path := "/companies/standup/summary?company_slug=" + workspaceSlug

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data TeamStandupSummary `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *standupService) CompletedYesterday(workspaceSlug string) ([]StandupTask, error) {
	path := "/companies/standup/completed-yesterday?company_slug=" + workspaceSlug

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []StandupTask `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *standupService) Blockers(workspaceSlug string) ([]StandupTask, error) {
	path := "/companies/standup/blockers?company_slug=" + workspaceSlug

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []StandupTask `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *standupService) TeamStatus(workspaceSlug string) ([]TeamMemberStatus, error) {
	path := "/companies/standup/team-status?company_slug=" + workspaceSlug

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []TeamMemberStatus `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}

func (s *standupService) WeeklyDigest(workspaceSlug string) (*WeeklyDigest, error) {
	path := "/companies/standup/weekly-digest?company_slug=" + workspaceSlug

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data WeeklyDigest `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *standupService) Contributors(workspaceSlug, period string) ([]Contributor, error) {
	path := "/companies/standup/contributors?company_slug=" + workspaceSlug
	if period != "" {
		path += "&period=" + period
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data []Contributor `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return result.Data, nil
}
