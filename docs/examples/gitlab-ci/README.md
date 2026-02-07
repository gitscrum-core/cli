# GitLab CI Examples for GitScrum CLI

This directory contains modular GitLab CI/CD templates for GitScrum integration.
Include only the templates you need in your `.gitlab-ci.yml`.

## Templates

| Template | File | Purpose |
|----------|------|---------|
| **MR Status Sync** | `mr-status-sync.yml` | Update task status on MR events |
| **Link Tasks** | `auto-link-task.yml` | Link commits and MRs to tasks |
| **Pipeline Status** | `pipeline-status.yml` | Sync CI status to tasks |
| **Sprint Report** | `sprint-report.yml` | Scheduled sprint summary |
| **Deploy Tracking** | `deploy-tracking.yml` | Track staging/prod deployments |

## Setup

1. Add variables in **Settings > CI/CD > Variables**:
   - `GITSCRUM_ACCESS_TOKEN`: Your GitScrum API token
   - `GITSCRUM_WORKSPACE`: Workspace slug
   - `GITSCRUM_PROJECT`: Project slug

2. Include templates in your `.gitlab-ci.yml`:

```yaml
include:
  - local: '.gitlab/ci/mr-status-sync.yml'
  - local: '.gitlab/ci/auto-link-task.yml'
```

## Branch Naming Convention

Use task codes in branch names:
- `feature/GS-123-login-form`
- `bugfix/GS-456-fix-crash`
- `hotfix/PROJ-789-security-patch`
