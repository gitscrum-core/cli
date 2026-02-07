# Command Reference

Detailed documentation for each GitScrum CLI command group.

---

## Command Groups

### Core Project Management

| Command | Description | Docs |
|:--------|:------------|:-----|
| `tasks` | Task lifecycle — create, assign, track, filter, complete | [tasks.md](tasks.md) |
| `timer` | Time tracking — start, stop, log, report | [timer.md](timer.md) |
| `sprints` | Sprint management — analytics, burndown, KPIs | [sprints.md](sprints.md) |
| `projects` | Project listing and details | [projects.md](projects.md) |
| `workspaces` | Workspace management | [workspaces.md](workspaces.md) |

### Analytics & Reports

| Command | Description | Docs |
|:--------|:------------|:-----|
| `analytics` | Sprint and project analytics | [analytics.md](analytics.md) |
| `standup` | Daily standup reports | [standup.md](standup.md) |

### Team Collaboration

| Command | Description | Docs |
|:--------|:------------|:-----|
| `chat` | Team chat and discussions | [chat.md](chat.md) |
| `wiki` | Project wiki pages | [wiki.md](wiki.md) |
| `notifications` | View and manage notifications | [notifications.md](notifications.md) |

### Client & CRM

| Command | Description | Docs |
|:--------|:------------|:-----|
| `clients` | Client management | [clients.md](clients.md) |
| `crm` | Customer relationship management | [crm.md](crm.md) |
| `invoices` | Invoice management and billing | [invoices.md](invoices.md) |
| `proposals` | Project proposals | [proposals.md](proposals.md) |

### Configuration & Setup

| Command | Description | Docs |
|:--------|:------------|:-----|
| `auth` | Authentication — login, logout, status | [auth.md](auth.md) |
| `config` | Configuration — global and project settings | [config.md](config.md) |
| `init` | Initialize GitScrum in a repository | [init.md](init.md) |
| `hooks` | Git hooks integration | [hooks.md](hooks.md) |

---

## Quick Start

```bash
# Authenticate
gitscrum auth login

# Set defaults
gitscrum config set workspace my-company
gitscrum config set project my-project

# Start working
gitscrum tasks current
gitscrum timer start
```

---

## Global Flags

All commands support these flags:

| Flag | Description |
|:-----|:------------|
| `--json` | Output in JSON format |
| `-q, --quiet` | Minimal output (IDs only) |
| `--no-color` | Disable colored output |
| `--help` | Show command help |

---

## Common Workflows

### Developer Daily Routine

```bash
# Morning
gitscrum notifications               # Check what's new
gitscrum tasks                       # What's on my plate?
gitscrum sprints current             # How's the sprint going?
gitscrum tasks branch GS-XXX         # Start working on a task
gitscrum timer start                 # Track my time

# Throughout the day
gitscrum tasks current               # What am I working on?
gitscrum timer                       # How long have I been working?
gitscrum chat                        # Check team messages

# End of day
gitscrum timer stop                  # Log my time
gitscrum timer report --day          # Review my day
gitscrum standup                     # Prepare for tomorrow
```

### Sprint Planning

```bash
# Create new sprint
gitscrum sprints create -n "Sprint 13" --start 2026-03-01 --end 2026-03-14

# Add tasks
gitscrum tasks update GS-XXX --sprint sprint-13

# Review analytics
gitscrum analytics velocity

# Close current sprint
gitscrum sprints close sprint-12
```

### Agency/Freelancer Workflow

```bash
# View clients
gitscrum clients

# Check time for billing
gitscrum timer report --month -p client-project

# Generate invoice
gitscrum invoices create --client techstart --period last-month

# Manage proposals
gitscrum proposals
```

### CI/CD Integration

```bash
# Set up authentication
export GITSCRUM_ACCESS_TOKEN="your-token"

# Update task status
gitscrum tasks update $TASK_CODE --status "in review"

# Log deployment
gitscrum timer log $TASK_CODE 5m -d "Deployment"
```

---

## See Also

- [Configuration Guide](../CONFIGURATION.md) — CLI and project settings
- [CI/CD Examples](../examples/README.md) — GitHub Actions, GitLab CI, Bitbucket
