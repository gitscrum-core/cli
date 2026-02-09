package services

import (
	"fmt"
	"time"

	"github.com/gitscrum-core/cli/pkg/api"
)

// ActiveTimer represents a running time entry
type ActiveTimer struct {
	UUID    string `json:"uuid"`
	Start   string `json:"start"`
	Comment string `json:"comment"`
	Task    Task   `json:"issue"`
}

// TimerReport represents aggregated time data
type TimerReport struct {
	TotalHours   float64         `json:"total_hours"`
	BillableHours float64        `json:"billable_hours"`
	ByProject    []ProjectTime   `json:"by_project"`
	ByDay        []DayTime       `json:"by_day"`
}

// ProjectTime represents time spent on a project
type ProjectTime struct {
	Project Project `json:"project"`
	Hours   float64 `json:"hours"`
}

// DayTime represents time spent on a day
type DayTime struct {
	Date  string  `json:"date"`
	Hours float64 `json:"hours"`
}

// Productivity represents productivity metrics
type Productivity struct {
	Score      float64 `json:"score"`
	TotalHours float64 `json:"total_hours"`
	TasksDone  int     `json:"tasks_done"`
	AvgPerDay  float64 `json:"avg_per_day"`
}

// TimerService handles time tracking operations
type TimerService interface {
	// Status returns the current active timer, if any
	Status() (*ActiveTimer, error)

	// Start begins a new timer for a task
	Start(taskCode, comment string) (*ActiveTimer, error)

	// Stop ends the current timer
	Stop() (*TimeEntry, error)

	// Log adds a manual time entry
	Log(taskCode string, duration time.Duration, comment string) (*TimeEntry, error)

	// Report returns time tracking report
	Report(week, team bool) (*TimerReport, error)

	// Productivity returns productivity metrics
	Productivity(period string) (*Productivity, error)
}

type timerService struct {
	client *api.Client
}

func newTimerService(client *api.Client) TimerService {
	return &timerService{client: client}
}

func (s *timerService) Status() (*ActiveTimer, error) {
	resp, err := s.client.Get("/time-trackings/active")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 204 {
		return nil, nil // No active timer
	}

	var result struct {
		Data ActiveTimer `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *timerService) Start(taskCode, comment string) (*ActiveTimer, error) {
	body := map[string]interface{}{}
	if comment != "" {
		body["comment"] = comment
	}

	// API uses POST /time-trackings/start with issue_code in body
	if taskCode != "" {
		body["issue_code"] = taskCode
	}

	resp, err := s.client.Post("/time-trackings/start", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data ActiveTimer `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *timerService) Stop() (*TimeEntry, error) {
	resp, err := s.client.Post("/time-trackings/stop-all", nil)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data TimeEntry `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *timerService) Log(taskCode string, duration time.Duration, comment string) (*TimeEntry, error) {
	hours := duration.Hours()
	body := map[string]interface{}{
		"duration": hours,
	}
	if comment != "" {
		body["comment"] = comment
	}

	// Use store endpoint with issue_code
	body["issue_code"] = taskCode
	resp, err := s.client.Post("/time-trackings", body)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data TimeEntry `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *timerService) Report(week, team bool) (*TimerReport, error) {
	path := "/time-trackings/reports?"
	if week {
		path += "period=week&"
	}
	if team {
		path += "scope=team&"
	}

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data TimerReport `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

func (s *timerService) Productivity(period string) (*Productivity, error) {
	path := fmt.Sprintf("/time-trackings/productivity?period=%s", period)

	resp, err := s.client.Get(path)
	if err != nil {
		return nil, err
	}
	var result struct {
		Data Productivity `json:"data"`
	}
	if err := handleResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result.Data, nil
}

