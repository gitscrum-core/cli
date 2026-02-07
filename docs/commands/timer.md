# Time Tracking

Timers, time logs, and productivity analytics — all from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum timer` | Show active timer status |
| `gitscrum timer start [CODE]` | Start timer on a task |
| `gitscrum timer stop` | Stop the active timer |
| `gitscrum timer log CODE DURATION` | Log time manually |
| `gitscrum timer report` | View time tracking reports |

---

## Real-World Scenarios

### Starting Your Workday

```bash
# Check if you left a timer running yesterday
$ gitscrum timer
No active timer.

# Start working on your first task
$ gitscrum timer start GS-1234
Timer started for GS-1234 at 09:00

# Or let the CLI detect from your current branch
$ git checkout feature/GS-1234-oauth-flow
$ gitscrum timer start
Timer started for GS-1234 at 09:00
```

### Logging Time After the Fact

```bash
# Forgot to start the timer? Log time manually
$ gitscrum timer log GS-1234 2h30m
Logged 2h 30m to GS-1234

# Add a description for clarity
$ gitscrum timer log GS-1234 45m -d "Code review for PR #42"
Logged 45m to GS-1234: Code review for PR #42

# Log time for a previous day
$ gitscrum timer log GS-1234 4h --date 2026-02-06
Logged 4h to GS-1234 on Feb 6, 2026
```

### End of Day: Stop Timer and Review

```bash
# Stop the active timer
$ gitscrum timer stop
Logged 3h 45m to GS-1234
Total today: 7h 15m

# Review what you worked on today
$ gitscrum timer report --day
Today (Feb 7, 2026) — 7h 15m

TASK      DURATION  DESCRIPTION
GS-1234   3h 45m    Implement OAuth flow
GS-1234   2h 30m    (from manual log)
GS-1198   1h 00m    Fix pagination
```

### Weekly Standup: What Did I Work On?

```bash
$ gitscrum timer report --week
This Week — 32h 15m

PROJECT       HOURS    TASKS
backend-api   18h 30m  GS-1234, GS-1198, GS-1156
web-app       10h 45m  GS-1089, GS-1091
docs          3h 00m   GS-1045

Most time: GS-1234 (12h 30m)
```

### Manager View: Team Report

```bash
$ gitscrum timer report --week --team
Team Time Report — This Week

USER      HOURS    TOP PROJECT
alice     38h 15m  backend-api
bob       35h 30m  web-app
charlie   29h 45m  mobile-app

Total: 103h 30m
```

---

## Duration Format

Time can be specified in multiple formats:

| Format | Example | Meaning |
|:-------|:--------|:--------|
| Hours and minutes | `2h30m` | 2 hours 30 minutes |
| Hours only | `4h` | 4 hours |
| Minutes only | `45m` | 45 minutes |
| Decimal hours | `2.5h` | 2 hours 30 minutes |
| Colon format | `2:30` | 2 hours 30 minutes |

---

## Parameters

### timer start

| Flag | Description |
|:-----|:------------|
| `-d, --description` | What you're working on |

If no task code is provided, the CLI attempts to detect it from the current git branch.

### timer log

| Flag | Description |
|:-----|:------------|
| `-d, --description` | Description of work done |
| `--date` | Date to log time for (YYYY-MM-DD, default: today) |

### timer report

| Flag | Description |
|:-----|:------------|
| `--day` | Today's entries |
| `--week` | This week's entries |
| `--month` | This month's entries |
| `-p, --project` | Filter by project |
| `--team` | Include team members |
| `--json` | Output as JSON |

---

## Workflow

```
1. gitscrum timer                  → check active timer
2. gitscrum timer start [CODE]     → start working
3. ... do your work ...
4. gitscrum timer stop             → log time automatically
5. gitscrum timer report --day     → review your day
```

---

## Tips

- **Git-aware**: If you don't specify a task code, the CLI detects it from your current branch
- **Rounding**: Configure automatic rounding in `.gitscrum.yml` (e.g., round to nearest 15 minutes)
- **Minimum duration**: Set `min_duration` to avoid logging tiny entries
- **Descriptions**: Always add descriptions — they help with invoicing and standup reports

---

## Configuration

In `.gitscrum.yml`:

```yaml
timer:
  # Auto-start timer when switching to task branch
  auto_start: true
  
  # Auto-stop timer when switching away
  auto_stop: true
  
  # Round to nearest 15 minutes
  round_to: 15
  
  # Minimum 5 minutes to log
  min_duration: 5
```

---

## CI/CD Integration

### Log Deployment Time

```yaml
# Track how long deployments take
- name: Log deployment time
  run: |
    gitscrum timer log $TASK_CODE "${{ env.DEPLOY_DURATION }}" \
      -d "Deployment to production"
```
