# Sprints

Sprint management, analytics, and burndown charts from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum sprints` | List all sprints |
| `gitscrum sprints current` | Current sprint with KPIs |
| `gitscrum sprints view SLUG` | Sprint details |
| `gitscrum sprints burndown` | ASCII burndown chart |
| `gitscrum sprints create` | Create a new sprint |
| `gitscrum sprints close SLUG` | Close a sprint |

---

## Real-World Scenarios

### Daily Standup: How's The Sprint Going?

```bash
$ gitscrum sprints current
Sprint 12 — "February Release"
Feb 1 - Feb 14, 2026 (Day 8 of 10)

Progress: 67% complete
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━ 67%

Tasks:  24 total | 16 done | 5 in progress | 3 todo
Points: 52 total | 35 done | 12 in progress | 5 todo

Velocity: 4.4 pts/day
Status: On track
```

### Visualize the Burndown

```bash
$ gitscrum sprints burndown
Sprint 12 Burndown

52 │●
44 │  ●
36 │    ●
28 │      ●───●
20 │            ●
12 │              ●
 4 │                ●
 0 └────────────────────────
     1  2  3  4  5  6  7  8

Ideal:     26 pts remaining
Actual:    17 pts remaining
Status:    Ahead of schedule (+9 pts)
```

### Sprint Planning: Create Next Sprint

```bash
$ gitscrum sprints create -n "Sprint 13" --start 2026-02-15 --end 2026-02-28
Created: Sprint 13 (Feb 15 - Feb 28)

# Add tasks to the new sprint
$ gitscrum tasks update GS-1301 --sprint sprint-13
$ gitscrum tasks update GS-1302 --sprint sprint-13
```

### Sprint Review: Closing Out

```bash
# Review sprint completion
$ gitscrum sprints view sprint-12
Sprint 12 — "February Release"

Completed: 22 of 24 tasks (92%)
Points:    47 of 52 (90%)

Incomplete:
  GS-1198  Fix pagination (In Progress, 3 pts)
  GS-1201  Add caching layer (Todo, 5 pts)

# Close the sprint
$ gitscrum sprints close sprint-12
Sprint 12 closed.
2 incomplete tasks moved to backlog.
```

### Historical Analysis

```bash
# List recent sprints
$ gitscrum sprints
SLUG        TITLE              STATUS   DATES              PROGRESS
sprint-12   February Release   Active   Feb 1 - Feb 14     67%
sprint-11   January Release    Closed   Jan 15 - Jan 28    100%
sprint-10   Q4 Wrap-up         Closed   Jan 1 - Jan 14     95%
```

---

## Parameters

### sprints (list)

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Filter by project |
| `--status` | Filter: active, closed, planned |
| `--limit` | Maximum results |
| `--json` | Output as JSON |

### sprints create

| Flag | Description |
|:-----|:------------|
| `-n, --name` | Sprint name (required) |
| `--start` | Start date YYYY-MM-DD (required) |
| `--end` | End date YYYY-MM-DD (required) |
| `-p, --project` | Project slug |

### sprints burndown

| Flag | Description |
|:-----|:------------|
| `--width` | Chart width (default: 40) |
| `--json` | Output raw burndown data |

---

## Workflow

```
1. gitscrum sprints                    → list sprints
2. gitscrum sprints current            → check progress
3. gitscrum sprints burndown           → visualize
4. gitscrum sprints create             → plan next sprint
5. gitscrum tasks update --sprint      → add tasks to sprint
6. gitscrum sprints close              → wrap up
```

---

## Tips

- **Quick check**: Use `gitscrum sprints current` in your morning routine
- **Team standup**: Pipe burndown to Slack with `gitscrum sprints burndown --json`
- **Velocity tracking**: Compare burndown across sprints to understand team velocity
- **JSON mode**: Use `--json` for integrating with dashboards and reporting tools

---

## CI/CD Integration

### Daily Sprint Report to Slack

```yaml
# .github/workflows/sprint-report.yml
on:
  schedule:
    - cron: '0 9 * * 1-5'  # Mon-Fri 9am

jobs:
  report:
    runs-on: ubuntu-latest
    steps:
      - name: Send sprint status
        run: |
          STATUS=$(gitscrum sprints current --json)
          curl -X POST $SLACK_WEBHOOK \
            -d "{\"text\": \"Sprint Status:\n$(echo $STATUS | jq -r '.progress')% complete\"}"
```

### Sprint Close Automation

```yaml
# Auto-close sprint when all tasks are done
- name: Check sprint completion
  run: |
    REMAINING=$(gitscrum sprints current --json | jq '.tasks.todo + .tasks.in_progress')
    if [ "$REMAINING" -eq 0 ]; then
      gitscrum sprints close current
    fi
```
