# CI/CD Examples

Automation examples for integrating GitScrum CLI with CI/CD platforms.

---

## Quick Start

### Install CLI in Pipeline

```bash
curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh
```

### Authenticate

```bash
export GITSCRUM_ACCESS_TOKEN="your-oauth-access-token"
```

Obtain the access token:
1. Run `gitscrum auth login` locally
2. Copy `access_token` from `~/.gitscrum/token.json`
3. Add as secret in your CI/CD platform

---

## Platform Examples

### GitHub Actions

| Example | Description |
|:--------|:------------|
| [pr-status-sync.yml](github-actions/pr-status-sync.yml) | Sync PR status with task status |
| [auto-link-task.yml](github-actions/auto-link-task.yml) | Auto-link PRs to tasks |
| [ci-status-sync.yml](github-actions/ci-status-sync.yml) | Sync CI status with task |
| [deploy-tracking.yml](github-actions/deploy-tracking.yml) | Track deployments |
| [sprint-automation.yml](github-actions/sprint-automation.yml) | Sprint lifecycle automation |
| [sprint-dashboard.yml](github-actions/sprint-dashboard.yml) | Generate sprint reports |
| [release-notes.yml](github-actions/release-notes.yml) | Generate release notes |
| [issue-sync.yml](github-actions/issue-sync.yml) | Sync GitHub Issues with tasks |
| [stale-reminder.yml](github-actions/stale-reminder.yml) | Alert on stale tasks |
| [blocker-alert.yml](github-actions/blocker-alert.yml) | Alert on blockers |

### GitLab CI

| Example | Description |
|:--------|:------------|
| [mr-status-sync.yml](gitlab-ci/mr-status-sync.yml) | Sync MR status with task status |
| [auto-link-task.yml](gitlab-ci/auto-link-task.yml) | Auto-link MRs to tasks |
| [pipeline-status.yml](gitlab-ci/pipeline-status.yml) | Sync pipeline status with task |
| [deploy-tracking.yml](gitlab-ci/deploy-tracking.yml) | Track deployments |
| [sprint-report.yml](gitlab-ci/sprint-report.yml) | Generate sprint reports |

### Bitbucket Pipelines

| Example | Description |
|:--------|:------------|
| [pr-status-sync.yml](bitbucket-pipelines/pr-status-sync.yml) | Sync PR status with task status |
| [auto-link-task.yml](bitbucket-pipelines/auto-link-task.yml) | Auto-link PRs to tasks |
| [pipeline-status.yml](bitbucket-pipelines/pipeline-status.yml) | Sync pipeline status with task |
| [deploy-tracking.yml](bitbucket-pipelines/deploy-tracking.yml) | Track deployments |
| [sprint-report.yml](bitbucket-pipelines/sprint-report.yml) | Generate sprint reports |

---

## Shell Scripts

| Script | Description |
|:-------|:------------|
| [pr-task-sync.sh](pr-task-sync.sh) | Cross-platform PR/MR to task sync |
| [daily-standup.sh](daily-standup.sh) | Generate daily standup reports |
| [sprint-health.sh](sprint-health.sh) | Sprint health checks |
| [time-report.sh](time-report.sh) | Time tracking reports |
| [blocker-alert.sh](blocker-alert.sh) | Alert on task blockers |
| [task-import.sh](task-import.sh) | Bulk task import |

---

## Git Hooks

| Hook | Description |
|:-----|:------------|
| [prepare-commit-msg](hooks/prepare-commit-msg) | Prepend task code to commits |
| [pre-push](hooks/pre-push) | Validate task status before push |

---

## Environment Variables

| Variable | Description | Required |
|:---------|:------------|:---------|
| `GITSCRUM_ACCESS_TOKEN` | OAuth access token | Yes (CI/CD) |
| `GITSCRUM_WORKSPACE` | Default workspace slug | Recommended |
| `GITSCRUM_PROJECT` | Default project slug | Recommended |
| `GITHUB_TOKEN` | GitHub API token | GitHub integration |
| `GITLAB_TOKEN` | GitLab Private Token | GitLab integration |
| `BITBUCKET_USER` | Bitbucket username | Bitbucket integration |
| `BITBUCKET_TOKEN` | Bitbucket App Password | Bitbucket integration |
| `SLACK_WEBHOOK_URL` | Slack incoming webhook | Notifications |

---

## Common Patterns

### Extract Task Code from Branch

```bash
BRANCH=$(git rev-parse --abbrev-ref HEAD)
TASK_CODE=$(echo "$BRANCH" | grep -oE '[A-Z]+-[0-9]+' | head -1)
```

### Update Task Status

```bash
gitscrum tasks update "$TASK_CODE" --status "in review"
```

### Check Task Exists

```bash
if gitscrum tasks view "$TASK_CODE" &>/dev/null; then
    echo "Task found"
fi
```

### Log Time

```bash
gitscrum timer log "$TASK_CODE" "2h" -d "Code review and testing"
```

---

## Project Configuration

Create `.gitscrum.yml` for repository-level defaults:

```yaml
version: "1"
workspace: my-company
project: my-project

automation:
  on_pr_merge: done
  complete_on_merge: true
```

See [.gitscrum.yml](.gitscrum.yml) for full options.
