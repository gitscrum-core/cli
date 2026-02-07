# Configuration Guide

GitScrum CLI configuration options and project settings.

---

## Configuration Locations

| Type | Location | Scope |
|:-----|:---------|:------|
| **Global** | `~/.gitscrum/config.yml` | User-wide defaults |
| **Project** | `.gitscrum.yml` | Repository-specific settings |

Project configuration (`.gitscrum.yml`) overrides global settings when present.

---

## Global Configuration

### Setting Values

```bash
gitscrum config set <key> <value>
```

| Key | Description |
|:-----|:------------|
| `workspace` | Default workspace slug |
| `project` | Default project slug |

**Examples:**
```bash
gitscrum config set workspace my-company
gitscrum config set project backend-api
```

### Viewing Configuration

```bash
gitscrum config
```

**Output:**
```
Workspace: my-company
Project:   backend-api
API URL:   https://api.gitscrum.com
```

---

## Project Configuration (.gitscrum.yml)

Create a `.gitscrum.yml` in your repository root to share settings across your team.

### Quick Setup

```bash
gitscrum init -w my-workspace -p my-project
```

### Full Configuration Reference

```yaml
# GitScrum CLI Project Configuration
# https://gitscrum.com/docs/cli/configuration

version: "1"

# Workspace and project association
workspace: my-company
project: my-project

# Branch naming conventions
branch:
  # Prefix for new branches: feature, bugfix, hotfix, release
  default_prefix: feature
  
  # Include task title in branch name
  include_title: true
  
  # Maximum branch name length
  max_length: 60
  
  # Branch name format
  # Available placeholders: {prefix}, {code}, {title}
  format: "{prefix}/{code}-{title}"

# Timer settings
timer:
  # Auto-start timer when switching to task branch
  auto_start: false
  
  # Auto-stop timer when switching away from task branch
  auto_stop: false
  
  # Round time entries to nearest interval (minutes)
  round_to: 15
  
  # Minimum duration to log (minutes)
  min_duration: 5
  
  # Remind to stop timer after inactivity (minutes, 0 = disabled)
  reminder_after: 120

# Git hooks integration
hooks:
  # Prepend task code to commit messages
  prepend_task_code: true
  
  # Commit message format (first %s = code, second %s = message)
  commit_format: "[%s] %s"
  
  # Validate task code exists before commit
  validate_task: false
  
  # Block commits without task code in branch name
  require_task_branch: false

# Automation rules for CI/CD integration
automation:
  # Update task status when PR is opened
  on_pr_open: in review
  
  # Update task status when PR is merged
  on_pr_merge: done
  
  # Complete task when PR is merged
  complete_on_merge: true
  
  # Auto-link PRs to tasks based on branch name
  link_pr_to_task: true

# Default values for new tasks
defaults:
  # Default task type
  type: task
  
  # Default effort estimate
  effort: 0
  
  # Default priority
  priority: medium
```

---

## Minimal Configuration

For simple setups, only workspace and project are required:

```yaml
version: "1"
workspace: my-company
project: my-project
```

---

## Environment Variables

Override configuration via environment variables:

| Variable | Description |
|:---------|:------------|
| `GITSCRUM_ACCESS_TOKEN` | OAuth access token (for CI/CD) |
| `GITSCRUM_WORKSPACE` | Override default workspace |
| `GITSCRUM_PROJECT` | Override default project |
| `GITSCRUM_API_URL` | Custom API URL (for enterprise) |

**Priority order:**
1. Command-line flags
2. Environment variables
3. Project configuration (`.gitscrum.yml`)
4. Global configuration (`~/.gitscrum/config.yml`)

---

## CI/CD Configuration

For headless environments like CI/CD pipelines:

### Authentication

```bash
export GITSCRUM_ACCESS_TOKEN="your-oauth-access-token"
```

Obtain the access token by:
1. Running `gitscrum auth login` locally
2. Copying the `access_token` from `~/.gitscrum/token.json`
3. Adding it as a secret in your CI/CD platform

### Example GitHub Actions

```yaml
env:
  GITSCRUM_ACCESS_TOKEN: ${{ secrets.GITSCRUM_ACCESS_TOKEN }}
  GITSCRUM_WORKSPACE: my-company
  GITSCRUM_PROJECT: my-project

steps:
  - name: Update task status
    run: gitscrum tasks update $TASK_CODE --status "deployed"
```

---

## Token Storage

OAuth tokens are stored locally:

| Platform | Location |
|:---------|:---------|
| Linux/macOS | `~/.gitscrum/token.json` |
| Windows | `%USERPROFILE%\.gitscrum\token.json` |

Permissions are set to `0600` (owner read/write only).

---

## Configuration Validation

Validate your project configuration:

```bash
gitscrum config validate
```

Checks:
- YAML syntax
- Required fields
- Valid workspace/project references
- Hook configuration
