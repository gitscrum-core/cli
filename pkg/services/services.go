// Package services provides business logic layer for GitScrum CLI
package services

import "github.com/gitscrum-core/cli/pkg/api"

// Services container provides access to all domain services
type Services struct {
	client *api.Client

	// Domain services
	Tasks         TasksService
	Timer         TimerService
	Projects      ProjectsService
	Sprints       SprintsService
	Auth          AuthService
	Analytics     AnalyticsService
	Chat          ChatService
	Clients       ClientsService
	CRM           CRMService
	Invoices      InvoicesService
	Notifications NotificationsService
	Proposals     ProposalsService
	Standup       StandupService
	Wiki          WikiService
	Workspaces    WorkspacesService
}

// New creates all services with the given API client
func New(client *api.Client) *Services {
	s := &Services{client: client}

	// Initialize all services
	s.Tasks = newTasksService(client)
	s.Timer = newTimerService(client)
	s.Projects = newProjectsService(client)
	s.Sprints = newSprintsService(client)
	s.Auth = newAuthService(client)
	s.Analytics = newAnalyticsService(client)
	s.Chat = newChatService(client)
	s.Clients = newClientsService(client)
	s.CRM = newCRMService(client)
	s.Invoices = newInvoicesService(client)
	s.Notifications = newNotificationsService(client)
	s.Proposals = newProposalsService(client)
	s.Standup = newStandupService(client)
	s.Wiki = newWikiService(client)
	s.Workspaces = newWorkspacesService(client)

	return s
}

// Client returns the underlying API client (for custom requests)
func (s *Services) Client() *api.Client {
	return s.client
}
