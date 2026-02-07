# GitHub Actions Examples for GitScrum CLI

This directory contains modular GitHub Actions workflows for GitScrum integration.
Copy only the workflows you need to your `.github/workflows/` directory.

## Workflows

| Workflow | File | Purpose |
|----------|------|---------|
| **PR Lifecycle** | `pr-status-sync.yml` | Update task status on PR events |
| **Link Tasks** | `auto-link-task.yml` | Link commits and PRs to tasks |
| **Sprint Dashboard** | `sprint-dashboard.yml` | Daily sprint summary in GitHub |
| **Release Notes** | `release-notes.yml` | Auto-generate release notes |
| **Blocker Alert** | `blocker-alert.yml` | Scheduled blocker notifications |
| **Issue Sync** | `issue-sync.yml` | Create tasks from GitHub Issues |
| **Deploy Tracking** | `deploy-tracking.yml` | Track staging/prod deployments |
| **Stale Reminder** | `stale-reminder.yml` | Alert for stuck tasks |
| **CI Status** | `ci-status-sync.yml` | Sync build status to tasks |
| **Sprint Automation** | `sprint-automation.yml` | End-of-sprint summary & migration |

## Setup

1. **Add secrets** in Settings > Secrets > Actions:
   - `GITSCRUM_ACCESS_TOKEN` - Your GitScrum API token

2. **Add variables** in Settings > Variables > Actions (optional):
   - `GITSCRUM_WORKSPACE` - Default workspace slug
   - `GITSCRUM_PROJECT` - Default project slug

3. **Copy workflows** to `.github/workflows/`:
   ```bash
   cp docs/examples/github-actions/*.yml .github/workflows/
   ```

## Branch Naming Convention

These workflows expect task codes in branch names:
- `feature/GS-123-login-page`
- `bugfix/GS-456-fix-crash`
- `hotfix/GS-789-security-patch`

Supported patterns: `XX-123`, `XXX-123`, `XXXX-123`, `XXXXX-123`
