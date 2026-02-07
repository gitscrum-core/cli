# Commands Reference

Complete reference for all GitScrum CLI commands.

---

## Global Flags

| Flag | Description |
|:-----|:------------|
| `--json` | Output in JSON format |
| `-q, --quiet` | Quiet mode (minimal output) |
| `--no-color` | Disable colored output |
| `--help` | Show help for command |
| `--version` | Show CLI version |

---

## Authentication

### auth login

Authenticate with GitScrum using OAuth Device Flow.

```bash
gitscrum auth login
```

Opens your browser to complete authentication. Credentials are stored locally in `~/.gitscrum/token.json`.

### auth logout

Remove stored credentials.

```bash
gitscrum auth logout
```

### auth status

Check authentication status.

```bash
gitscrum auth status
```

### auth whoami

Show authenticated user details.

```bash
gitscrum auth whoami
```

**Output:**
```
Logged in as John Doe (john@example.com)
```

---

## Configuration

### config

Show current configuration.

```bash
gitscrum config
```

### config set

Set configuration values.

```bash
gitscrum config set workspace <slug>
gitscrum config set project <slug>
```

**Examples:**
```bash
gitscrum config set workspace my-company
gitscrum config set project backend-api
```

---

## Tasks

### tasks

List tasks assigned to you.

```bash
gitscrum tasks [flags]
```

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Filter by project slug |
| `-s, --sprint` | Filter by sprint |
| `--status` | Filter by status |
| `--limit` | Maximum results (default: 20) |

**Examples:**
```bash
gitscrum tasks
gitscrum tasks -p my-project
gitscrum tasks --status "in progress"
gitscrum tasks --json
```

### tasks view

View task details.

```bash
gitscrum tasks view <code>
```

**Example:**
```bash
gitscrum tasks view GS-1234
```

### tasks create

Create a new task.

```bash
gitscrum tasks create <title> [flags]
```

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Project slug (required if not set globally) |
| `-d, --description` | Task description |
| `--type` | Task type slug |
| `--effort` | Effort points |
| `--assignee` | Assignee username |

**Examples:**
```bash
gitscrum tasks create "Fix login bug"
gitscrum tasks create "Implement OAuth" -p backend --effort 5
```

### tasks update

Update an existing task.

```bash
gitscrum tasks update <code> [flags]
```

| Flag | Description |
|:-----|:------------|
| `--title` | New title |
| `--status` | New status |
| `--effort` | Effort points |

**Example:**
```bash
gitscrum tasks update GS-1234 --status "in review"
```

### tasks current

Show task for current git branch.

```bash
gitscrum tasks current
```

Automatically extracts task code from branch name (e.g., `feature/GS-1234-description`).

### tasks branch

Create a git branch from a task.

```bash
gitscrum tasks branch <code>
```

**Example:**
```bash
gitscrum tasks branch GS-1234
# Creates: feature/GS-1234-implement-user-auth
```

---

## Time Tracking

### timer

Show active timer status.

```bash
gitscrum timer
```

### timer start

Start a timer for a task.

```bash
gitscrum timer start [code]
```

If no code is provided, uses the task from current git branch.

**Examples:**
```bash
gitscrum timer start GS-1234
gitscrum timer start  # Uses current branch task
```

### timer stop

Stop the active timer.

```bash
gitscrum timer stop
```

### timer log

Log time manually.

```bash
gitscrum timer log <code> <duration> [flags]
```

| Flag | Description |
|:-----|:------------|
| `-d, --description` | Description of work |
| `--date` | Date (YYYY-MM-DD, default: today) |

Duration format: `2h`, `30m`, `2h30m`, `1.5h`

**Examples:**
```bash
gitscrum timer log GS-1234 2h30m
gitscrum timer log GS-1234 45m -d "Code review"
gitscrum timer log GS-1234 3h --date 2026-02-06
```

### timer report

View time tracking report.

```bash
gitscrum timer report [flags]
```

| Flag | Description |
|:-----|:------------|
| `--day` | Today's entries |
| `--week` | This week's entries |
| `--month` | This month's entries |
| `-p, --project` | Filter by project |

**Examples:**
```bash
gitscrum timer report --week
gitscrum timer report --month -p my-project
```

---

## Sprints

### sprints

List all sprints.

```bash
gitscrum sprints [flags]
```

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Project slug |
| `--status` | Filter by status (active, closed, planned) |

### sprints current

Show current sprint details with KPIs.

```bash
gitscrum sprints current
```

### sprints view

View sprint details.

```bash
gitscrum sprints view <slug>
```

### sprints burndown

Display ASCII burndown chart.

```bash
gitscrum sprints burndown [slug]
```

**Example output:**
```
Sprint 12 Burndown (Day 5 of 10)
50 │●
40 │  ●
30 │    ●───●
20 │          ●
10 │            ●
 0 └──────────────────
     1  2  3  4  5  6

Remaining: 26 pts | Velocity: 5.2 pts/day
Status: On track
```

### sprints create

Create a new sprint.

```bash
gitscrum sprints create [flags]
```

| Flag | Description |
|:-----|:------------|
| `-n, --name` | Sprint name |
| `--start` | Start date (YYYY-MM-DD) |
| `--end` | End date (YYYY-MM-DD) |

### sprints close

Close a sprint.

```bash
gitscrum sprints close <slug>
```

---

## Projects

### projects

List all projects in workspace.

```bash
gitscrum projects [flags]
```

| Flag | Description |
|:-----|:------------|
| `-w, --workspace` | Workspace slug |

### projects view

View project details.

```bash
gitscrum projects view <slug>
```

---

## Workspaces

### workspaces

List all workspaces.

```bash
gitscrum workspaces
```

---

## Initialization

### init

Initialize GitScrum in current repository.

```bash
gitscrum init [flags]
```

| Flag | Description |
|:-----|:------------|
| `-w, --workspace` | Workspace slug |
| `-p, --project` | Project slug |

Creates a `.gitscrum.yml` configuration file.

---

## Shell Completion

Generate shell completion scripts.

```bash
gitscrum completion bash
gitscrum completion zsh
gitscrum completion fish
gitscrum completion powershell
```

**Installation:**

```bash
# Bash
gitscrum completion bash > /etc/bash_completion.d/gitscrum

# Zsh
gitscrum completion zsh > "${fpath[1]}/_gitscrum"

# Fish
gitscrum completion fish > ~/.config/fish/completions/gitscrum.fish

# PowerShell
gitscrum completion powershell >> $PROFILE
```

---

## Version

Show CLI version.

```bash
gitscrum --version
```
