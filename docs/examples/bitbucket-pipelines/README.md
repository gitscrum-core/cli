# Bitbucket Pipelines Examples for GitScrum CLI

This directory contains modular Bitbucket Pipelines configurations for GitScrum integration.
Copy only the templates you need to your `bitbucket-pipelines.yml`.

## Templates

| Template | File | Purpose |
|----------|------|---------|
| **PR Status Sync** | `pr-status-sync.yml` | Update task status on PR events |
| **Link Tasks** | `auto-link-task.yml` | Link commits and PRs to tasks |
| **Pipeline Status** | `pipeline-status.yml` | Sync build status to tasks |
| **Deploy Tracking** | `deploy-tracking.yml` | Track staging/prod deployments |
| **Sprint Report** | `sprint-report.yml` | Scheduled sprint summary |

## Setup

1. Add variables in **Repository Settings > Repository variables**:
   - `GITSCRUM_ACCESS_TOKEN`: Your GitScrum API token
   - `GITSCRUM_WORKSPACE`: Workspace slug
   - `GITSCRUM_PROJECT`: Project slug

2. Copy a template to your `bitbucket-pipelines.yml`

## Branch Naming Convention

Use task codes in branch names:
- `feature/GS-123-login-form`
- `bugfix/GS-456-fix-crash`
- `hotfix/PROJ-789-security-patch`

## Custom Pipelines

Bitbucket supports manual triggers via **custom pipelines**.
Configure in Repository Settings > Pipelines > Schedules for scheduled runs.
