# Projects

Project listing, details, and management.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum projects` | List all projects |
| `gitscrum projects view SLUG` | View project details |

---

## Real-World Scenarios

### Discovering Your Projects

```bash
$ gitscrum projects
SLUG          NAME                 OPEN TASKS  SPRINTS
backend-api   Backend API          23          Sprint 12
web-app       Web Application      18          Sprint 11
mobile-app    Mobile App           12          Sprint 10
docs          Documentation        8           —
```

### Project Overview

```bash
$ gitscrum projects view backend-api
Backend API

Description:
  Core REST API for the platform. Handles authentication,
  data processing, and third-party integrations.

Stats:
  Tasks:      23 open, 156 completed
  Sprints:    12 total, 1 active
  Team:       5 members

Active Sprint: Sprint 12 — February Release
  Progress: 67% (16 of 24 tasks)
  Ends: Feb 14, 2026 (6 days)

Recent Activity:
  GS-1234  OAuth flow implemented (alice, 2h ago)
  GS-1198  Started pagination fix (bob, 4h ago)
```

### Switching Context

```bash
# Set default project for current session
$ gitscrum config set project backend-api

# Now all commands use this project by default
$ gitscrum tasks
# Shows tasks from backend-api

$ gitscrum sprints current
# Shows current sprint from backend-api
```

### Multi-Project Workflow

```bash
# Work across multiple projects
$ gitscrum tasks -p backend-api --status "in progress"
$ gitscrum tasks -p web-app --status "in progress"

# View combined time report
$ gitscrum timer report --week
This Week — 40h

PROJECT       HOURS
backend-api   25h 30m
web-app       12h 15m
mobile-app    2h 15m
```

---

## Parameters

### projects (list)

| Flag | Description |
|:-----|:------------|
| `-w, --workspace` | Workspace slug |
| `--json` | Output as JSON |
| `-q, --quiet` | Output only slugs |

### projects view

| Flag | Description |
|:-----|:------------|
| `--json` | Output as JSON |

---

## Related Commands

Projects are foundational to other commands:

```bash
# Filter tasks by project
gitscrum tasks -p backend-api

# Filter sprints by project
gitscrum sprints -p web-app

# Filter time report by project
gitscrum timer report --week -p mobile-app
```

---

## Tips

- **Set defaults**: Use `gitscrum config set project <slug>` to avoid typing `-p` every time
- **Repository config**: Add project to `.gitscrum.yml` for repository-level defaults
- **Quick switch**: List projects with `-q` for scripting and quick reference
