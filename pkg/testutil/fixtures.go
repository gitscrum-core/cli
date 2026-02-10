// Package testutil provides shared fixtures for CLI integration testing
package testutil

// TaskResponse represents a task from the API (based on BoardTaskResource)
type TaskResponse struct {
	UUID        string            `json:"uuid"`
	Code        string            `json:"code"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	State       string            `json:"state"`
	Workflow    *WorkflowResponse `json:"workflow"`
	Users       []UserResponse    `json:"users"`
	Settings    TaskSettings      `json:"settings"`
	Stats       TaskStats         `json:"stats"`
	Project     *ProjectCompact   `json:"project"`
	Company     *CompanyCompact   `json:"company"`
	Sprint      *SprintCompact    `json:"sprint"`
	CreatedAt   *DateResponse     `json:"created_at"`
	DueDate     *DateResponse     `json:"due_date"`
}

type WorkflowResponse struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Color string `json:"color"`
}

type TaskSettings struct {
	IsBlocker bool `json:"is_blocker"`
	IsBug     bool `json:"is_bug"`
	IsDraft   bool `json:"is_draft"`
}

type TaskStats struct {
	Votes       int `json:"votes"`
	Comments    int `json:"comments"`
	Attachments int `json:"attachments"`
	Subtasks    int `json:"subtasks"`
}

// UserResponse represents a user from the API
type UserResponse struct {
	UUID     string  `json:"uuid"`
	Username string  `json:"username"`
	Name     string  `json:"name"`
	Email    string  `json:"email"`
	Photo    *string `json:"photo"`
}

// ProjectResponse represents a project from the API (based on ProjectResource)
type ProjectResponse struct {
	ID          int             `json:"id"`
	UUID        string          `json:"uuid"`
	Slug        string          `json:"slug"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Code        string          `json:"code"`
	Logo        string          `json:"logo"`
	Company     *ProjectCompany `json:"company"`
	Stats       ProjectStats    `json:"stats"`
	Settings    ProjectSettings `json:"settings"`
	Status      ProjectStatus   `json:"status"`
	Dates       ProjectDates    `json:"dates"`
	Users       []UserResponse  `json:"users"`
	CreatedAt   *DateResponse   `json:"created_at"`
}

type ProjectCompact struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type CompanyCompact struct {
	Slug string `json:"slug"`
	Name string `json:"name"`
}

type ProjectCompany struct {
	UUID string `json:"uuid"`
	Slug string `json:"slug"`
	Name string `json:"name"`
	Logo string `json:"logo"`
}

type ProjectStats struct {
	Total    int `json:"total"`
	Open     int `json:"open"`
	Progress int `json:"progress"`
	Closed   int `json:"closed"`
}

type ProjectSettings struct {
	UseTimer       bool `json:"use_timer"`
	TaskType       bool `json:"task_type"`
	ShowNumber     bool `json:"show_number"`
	HasSprints     bool `json:"has_sprints"`
	HasWiki        bool `json:"has_wiki"`
	HasUserStories bool `json:"has_user_stories"`
}

type ProjectStatus struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type ProjectDates struct {
	Start *DateResponse `json:"start"`
	Due   *DateResponse `json:"due"`
}

// SprintResponse represents a sprint from the API (based on SprintResource)
type SprintResponse struct {
	ID          int             `json:"id"`
	Code        string          `json:"code"`
	Slug        string          `json:"slug"`
	Title       string          `json:"title"`
	Color       string          `json:"color"`
	Description string          `json:"description"`
	Duration    int             `json:"duration"`
	Timebox     SprintTimebox   `json:"timebox"`
	Stats       SprintStats     `json:"stats"`
	Status      SprintStatus    `json:"status"`
	Project     *ProjectCompact `json:"project"`
	Company     *CompanyCompact `json:"company"`
	CreatedAt   *DateResponse   `json:"created_at"`
}

type SprintCompact struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
}

type SprintTimebox struct {
	Start  *DateResponse `json:"start"`
	Finish *DateResponse `json:"finish"`
}

type SprintStats struct {
	StoryPoints int `json:"story_points"`
	WorkedHours int `json:"worked_hours"`
	TotalTasks  int `json:"total_tasks"`
	ClosedTasks int `json:"closed_tasks"`
	Percentage  int `json:"percentage"`
	Comments    int `json:"comments"`
}

type SprintStatus struct {
	Slug        string `json:"slug"`
	Title       string `json:"title"`
	Color       string `json:"color"`
	Description string `json:"description"`
}

// TimeTrackingResponse represents a time tracking entry
type TimeTrackingResponse struct {
	Comment    string           `json:"comment"`
	Time       TimeTrackingTime `json:"time"`
	Task       *TaskResponse    `json:"task"`
	User       *UserResponse    `json:"user"`
	Billed     bool             `json:"billed"`
	IsBillable bool             `json:"is_billable"`
}

type TimeTrackingTime struct {
	ID              int           `json:"id"`
	Start           *DateResponse `json:"start"`
	End             *DateResponse `json:"end"`
	Total           string        `json:"total"`
	DurationMinutes int           `json:"duration_minutes"`
}

// DateResponse represents a date from the API (based on DateResource)
type DateResponse struct {
	Date     string `json:"date"`
	Timezone string `json:"timezone"`
	Ago      string `json:"ago"`
}

// NotificationResponse represents a notification
type NotificationResponse struct {
	UUID      string        `json:"uuid"`
	Type      string        `json:"type"`
	Title     string        `json:"title"`
	Message   string        `json:"message"`
	ReadAt    *DateResponse `json:"read_at"`
	CreatedAt *DateResponse `json:"created_at"`
}

// APIResponse wraps API responses
type APIResponse struct {
	Data interface{} `json:"data"`
	Meta *APIMeta    `json:"meta,omitempty"`
}

type APIMeta struct {
	CurrentPage int `json:"current_page"`
	PerPage     int `json:"per_page"`
	Total       int `json:"total"`
	LastPage    int `json:"last_page"`
}

// Fixtures provides pre-configured test data
var Fixtures = struct {
	Task        TaskResponse
	Tasks       []TaskResponse
	Project     ProjectResponse
	Projects    []ProjectResponse
	Sprint      SprintResponse
	Sprints     []SprintResponse
	TimeTracker TimeTrackingResponse
	User        UserResponse
}{
	User: UserResponse{
		UUID:     "550e8400-e29b-41d4-a716-446655440001",
		Username: "johndoe",
		Name:     "John Doe",
		Email:    "john@example.com",
		Photo:    nil,
	},
	Task: TaskResponse{
		UUID:        "550e8400-e29b-41d4-a716-446655440000",
		Code:        "GS-123",
		Slug:        "fix-login-issue",
		Title:       "Fix login issue on mobile",
		Description: "Users cannot login on iOS devices",
		State:       "open",
		Workflow: &WorkflowResponse{
			Slug:  "in-progress",
			Title: "In Progress",
			Color: "#3498db",
		},
		Settings: TaskSettings{
			IsBlocker: false,
			IsBug:     true,
			IsDraft:   false,
		},
		Stats: TaskStats{
			Votes:       5,
			Comments:    3,
			Attachments: 2,
			Subtasks:    4,
		},
		Project: &ProjectCompact{
			Slug: "mobile-app",
			Name: "Mobile App",
		},
	},
	Tasks: []TaskResponse{
		{
			UUID:  "550e8400-e29b-41d4-a716-446655440000",
			Code:  "GS-123",
			Slug:  "fix-login-issue",
			Title: "Fix login issue on mobile",
			State: "open",
			Workflow: &WorkflowResponse{
				Slug:  "in-progress",
				Title: "In Progress",
				Color: "#3498db",
			},
		},
		{
			UUID:  "550e8400-e29b-41d4-a716-446655440002",
			Code:  "GS-124",
			Slug:  "add-dark-mode",
			Title: "Add dark mode support",
			State: "open",
			Workflow: &WorkflowResponse{
				Slug:  "todo",
				Title: "To Do",
				Color: "#95a5a6",
			},
		},
		{
			UUID:  "550e8400-e29b-41d4-a716-446655440003",
			Code:  "GS-125",
			Slug:  "optimize-performance",
			Title: "Optimize app performance",
			State: "closed",
			Workflow: &WorkflowResponse{
				Slug:  "done",
				Title: "Done",
				Color: "#27ae60",
			},
		},
	},
	Project: ProjectResponse{
		ID:          1,
		UUID:        "660e8400-e29b-41d4-a716-446655440000",
		Slug:        "mobile-app",
		Name:        "Mobile App",
		Description: "Our flagship mobile application",
		Code:        "MA",
		Company: &ProjectCompany{
			UUID: "770e8400-e29b-41d4-a716-446655440000",
			Slug: "acme-corp",
			Name: "ACME Corporation",
		},
		Stats: ProjectStats{
			Total:    50,
			Open:     20,
			Progress: 15,
			Closed:   15,
		},
		Settings: ProjectSettings{
			UseTimer:   true,
			TaskType:   true,
			ShowNumber: true,
			HasSprints: true,
			HasWiki:    true,
		},
		Status: ProjectStatus{
			Code:  "active",
			Title: "Active",
		},
	},
	Projects: []ProjectResponse{
		{
			ID:   1,
			UUID: "660e8400-e29b-41d4-a716-446655440000",
			Slug: "mobile-app",
			Name: "Mobile App",
			Stats: ProjectStats{
				Total: 50, Open: 20, Progress: 15, Closed: 15,
			},
		},
		{
			ID:   2,
			UUID: "660e8400-e29b-41d4-a716-446655440001",
			Slug: "web-platform",
			Name: "Web Platform",
			Stats: ProjectStats{
				Total: 120, Open: 45, Progress: 30, Closed: 45,
			},
		},
	},
	Sprint: SprintResponse{
		ID:          1,
		Code:        "SPR-1",
		Slug:        "sprint-1-2024",
		Title:       "Sprint 1 - February 2024",
		Color:       "#3498db",
		Description: "First sprint of Q1",
		Duration:    14,
		Timebox: SprintTimebox{
			Start:  &DateResponse{Date: "2024-02-01", Timezone: "UTC", Ago: "6 days ago"},
			Finish: &DateResponse{Date: "2024-02-14", Timezone: "UTC", Ago: "in 8 days"},
		},
		Stats: SprintStats{
			StoryPoints: 21,
			WorkedHours: 45,
			TotalTasks:  15,
			ClosedTasks: 8,
			Percentage:  53,
			Comments:    25,
		},
		Status: SprintStatus{
			Slug:  "sprint-open",
			Title: "Open",
			Color: "079a0d",
		},
		Project: &ProjectCompact{
			Slug: "mobile-app",
			Name: "Mobile App",
		},
	},
	Sprints: []SprintResponse{
		{
			ID:       1,
			Code:     "SPR-1",
			Slug:     "sprint-1-2024",
			Title:    "Sprint 1 - February 2024",
			Duration: 14,
			Stats:    SprintStats{TotalTasks: 15, ClosedTasks: 8, Percentage: 53},
			Status:   SprintStatus{Slug: "sprint-open", Title: "Open"},
		},
		{
			ID:       2,
			Code:     "SPR-2",
			Slug:     "sprint-2-2024",
			Title:    "Sprint 2 - February 2024",
			Duration: 14,
			Stats:    SprintStats{TotalTasks: 20, ClosedTasks: 0, Percentage: 0},
			Status:   SprintStatus{Slug: "sprint-planned", Title: "Planned"},
		},
	},
	TimeTracker: TimeTrackingResponse{
		Comment: "Working on login fix",
		Time: TimeTrackingTime{
			ID:              1,
			Total:           "02:30:00",
			DurationMinutes: 150,
		},
		Billed:     false,
		IsBillable: true,
	},
}
