# Standup

Team standup analytics and daily status reports.

Standup commands provide analytical views of your team's work - completed tasks, blockers, and weekly metrics. Data is automatically aggregated from tasks and time tracking.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum standup` | Summary view (completed, in progress, blocked) |
| `gitscrum standup completed` | Tasks completed yesterday |
| `gitscrum standup blockers` | Current blocked tasks |
| `gitscrum standup team` | Team member status and workload |
| `gitscrum standup digest` | Weekly productivity metrics |

---

## Real-World Scenarios

### Daily Standup Summary

```bash
$ gitscrum standup
DAILY STANDUP - 2026-02-07
──────────────────────────────────────────────────

  Completed yesterday:    12 tasks
  In progress:            8 tasks
  Blocked:                2 tasks
  Hours tracked:          24.5h
```

### What Was Completed Yesterday

```bash
$ gitscrum standup completed
COMPLETED YESTERDAY (2026-02-06):

  [GS-1234] Implement OAuth flow
  [GS-1198] Code review for pagination
  [GS-1156] Fix database migration
```

### Check Current Blockers

```bash
$ gitscrum standup blockers
CURRENT BLOCKERS:

  [!] [GS-1156] Waiting on API spec
      Assigned to: bob

  [!] [GS-1178] Blocked by external dependency
      Assigned to: alice
```

### Team Status

```bash
$ gitscrum standup team
TEAM STATUS
────────────────────────────────────────────────────────────

  Alice
    Status: working | In Progress: 3 | Done Today: 2 | Blocked: 0
    Hours Tracked: 4.5h

  Bob
    Status: blocked | In Progress: 1 | Done Today: 0 | Blocked: 1
    Hours Tracked: 2.0h

  Charlie
    Status: available | In Progress: 0 | Done Today: 4 | Blocked: 0
```

### Weekly Digest

```bash
$ gitscrum standup digest
WEEKLY DIGEST
──────────────────────────────────────────────────

  Tasks Completed: 45
  Hours Tracked:   128.5h
  Velocity:        +15%

  TOP CONTRIBUTORS:
    1. Alice (15 tasks)
    2. Bob (12 tasks)
    3. Charlie (10 tasks)

  DAILY BREAKDOWN:
    2026-02-03: 8 completed, 5 created
    2026-02-04: 10 completed, 7 created
    2026-02-05: 9 completed, 4 created
    2026-02-06: 12 completed, 6 created
    2026-02-07: 6 completed, 8 created
```

---

## Scope Filtering

Commands support both workspace-level and project-level views:

```bash
# Workspace-level (all projects)
$ gitscrum standup

# Project-level (filter by project)
$ gitscrum standup --project my-app
$ gitscrum standup team --project my-app
$ gitscrum standup digest --project my-app
```

---

## Tips

- **Morning routine**: Run `gitscrum standup` before your daily meeting
- **Focus on blockers**: Use `gitscrum standup blockers` to identify and resolve impediments
- **Team visibility**: Use `gitscrum standup team` to see who might need help
- **Weekly review**: Use `gitscrum standup digest` for sprint retrospectives
