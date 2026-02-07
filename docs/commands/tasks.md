# Tasks

Full task lifecycle — create, assign, track, filter, and complete from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum tasks` | List tasks assigned to you |
| `gitscrum tasks view CODE` | View task details |
| `gitscrum tasks create TITLE` | Create a new task |
| `gitscrum tasks update CODE` | Update task properties |
| `gitscrum tasks current` | Show task for current git branch |
| `gitscrum tasks branch CODE` | Create git branch from task |

---

## Real-World Scenarios

### Morning Standup: What's On My Plate?

```bash
$ gitscrum tasks
CODE      TITLE                              STATUS        SPRINT
GS-1234   Implement OAuth flow               In Progress   Sprint 12
GS-1198   Fix pagination on dashboard        Todo          Sprint 12
GS-1156   Add unit tests for auth module     In Review     Sprint 12
```

### Starting Work on a Task

```bash
# Create a properly formatted branch from task
$ gitscrum tasks branch GS-1234
Switched to new branch 'feature/GS-1234-implement-oauth-flow'

# Start your timer automatically
$ gitscrum timer start
Timer started for GS-1234 at 09:15
```

### Quick Task Creation During Development

```bash
# Found a bug? Create a task without leaving terminal
$ gitscrum tasks create "Fix null pointer in user service" --type bug --effort 3
Created: GS-1267 - Fix null pointer in user service

# Assign to yourself and add to current sprint
$ gitscrum tasks update GS-1267 --assignee me --sprint current
Updated GS-1267
```

### Context Switching: What Was I Working On?

```bash
# Came back from a meeting — what branch am I on?
$ gitscrum tasks current
GS-1234: Implement OAuth flow
Status: In Progress | Sprint: Sprint 12 | Effort: 5 pts

Assigned: you
Due: Feb 14, 2026
Sprint: Sprint 12 (Day 5 of 10)

Description:
  Implement OAuth 2.0 Device Flow for CLI authentication...
```

### Filtering Tasks

```bash
# Show only in-progress tasks
$ gitscrum tasks --status "in progress"

# Filter by project
$ gitscrum tasks -p backend-api

# Show blockers across all projects
$ gitscrum tasks --blocker
```

---

## Parameters

### tasks (list)

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Filter by project slug |
| `-s, --sprint` | Filter by sprint slug |
| `--status` | Filter by status (todo, in progress, done) |
| `--blocker` | Show only blockers |
| `--limit` | Maximum results (default: 20) |
| `--json` | Output as JSON |
| `-q, --quiet` | Output only task codes |

### tasks create

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Project slug (required if not set globally) |
| `-d, --description` | Task description (Markdown) |
| `--type` | Task type (bug, feature, chore, etc.) |
| `--effort` | Effort points |
| `--assignee` | Assignee username |
| `--sprint` | Sprint slug |
| `--due` | Due date (YYYY-MM-DD) |

### tasks update

| Flag | Description |
|:-----|:------------|
| `--title` | New title |
| `--status` | New status |
| `--assignee` | New assignee |
| `--effort` | Effort points |
| `--sprint` | Move to sprint |

### tasks branch

| Flag | Description |
|:-----|:------------|
| `--prefix` | Branch prefix (default: feature) |
| `--no-switch` | Create branch but don't switch to it |

---

## Workflow

```
1. gitscrum tasks                  → see your assigned work
2. gitscrum tasks branch GS-XXX    → create branch from task
3. gitscrum timer start            → start tracking time
4. gitscrum tasks update --status  → update progress
5. gitscrum timer stop             → log your time
6. git push                        → push your work
```

---

## Tips

- **Git-aware**: The CLI automatically detects task codes from branch names matching patterns like `feature/GS-1234-description`
- **Current task**: Use `gitscrum tasks current` to quickly see what you're working on based on your current branch
- **Quick filter**: Use `--json` with `jq` for advanced filtering in scripts
- **Quiet mode**: Use `-q` to get only task codes — perfect for piping to other commands

---

## CI/CD Integration

### Update Task Status on PR Merge

```yaml
# GitHub Actions
- name: Complete task
  run: |
    TASK_CODE=$(echo "${{ github.head_ref }}" | grep -oE '[A-Z]+-[0-9]+')
    gitscrum tasks update $TASK_CODE --status "done"
```

### Auto-link Commits to Tasks

Configure in `.gitscrum.yml`:

```yaml
hooks:
  prepend_task_code: true
  commit_format: "[%s] %s"
```
