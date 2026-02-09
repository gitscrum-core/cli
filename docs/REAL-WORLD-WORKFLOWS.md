# Real-World Workflows

Practical scenarios showing GitScrum CLI in action. Copy, paste, adapt.

---

## Watch in Action

[![asciicast](https://asciinema.org/a/gitscrum-workflow.svg)](https://asciinema.org/a/gitscrum-workflow)

> Record your own: `asciinema rec gitscrum-demo.cast`

---

## Developer Daily Workflow

### Morning Startup (2 min)

```bash
# What needs my attention?
$ gitscrum notifications
3 new notifications:
  • @alice mentioned you in GS-1234
  • GS-1198 moved to "In Review"
  • Sprint 12 ends in 2 days

# What's on my plate today?
$ gitscrum tasks
CODE      TITLE                         STATUS        EFFORT
GS-1234   Implement OAuth flow          In Progress   5 pts
GS-1198   Fix pagination bug            Todo          2 pts
GS-1201   Add caching layer             Todo          8 pts

# How's the sprint going?
$ gitscrum sprints current
Sprint 12 — Day 8 of 10
Progress: ━━━━━━━━━━━━━━━━━━░░░░ 67%
Remaining: 18 pts | Velocity: 5.2 pts/day
Status: On track ✓
```

### Start Working on a Task

```bash
# Create branch and start timer in one flow
$ gitscrum tasks branch GS-1234
Switched to branch 'feature/GS-1234-implement-oauth-flow'

$ gitscrum timer start
Timer started for GS-1234 at 09:15

# Check what you're working on anytime
$ gitscrum timer
Active: GS-1234 — Implement OAuth flow
Running: 2h 15m
```

### Context Switching

```bash
# Stop current timer before switching
$ gitscrum timer stop
Logged 2h 15m to GS-1234

# Switch to urgent bug
$ git checkout -b fix/GS-1199-critical-bug
$ gitscrum timer start GS-1199
Timer started for GS-1199 at 11:30

# Quick fix, back to main task
$ gitscrum timer stop
Logged 25m to GS-1199

$ git checkout feature/GS-1234-implement-oauth-flow
$ gitscrum timer start
Timer started for GS-1234 at 11:55
```

### End of Day

```bash
# Stop any running timer
$ gitscrum timer stop
Logged 3h 45m to GS-1234
Total today: 6h 25m

# Quick review
$ gitscrum timer report --day
TODAY — Feb 7, 2026 — 6h 25m

TASK      DURATION  DESCRIPTION
GS-1234   6h 00m    Implement OAuth flow
GS-1199   0h 25m    Critical bug fix

# Prepare for standup
$ gitscrum standup
DAILY SUMMARY
  Completed: 1 task
  In Progress: 1 task (GS-1234)
  Blocked: 0
```

---

## PR Workflow

### Create PR with Task Link

```bash
# Your branch already has the task code
$ git push origin feature/GS-1234-implement-oauth-flow

# Update task status
$ gitscrum tasks update GS-1234 --status "in review"
Task GS-1234 updated to "In Review"

# The PR title can reference the task
# Title: [GS-1234] Implement OAuth flow
```

### After PR Merged

```bash
# Mark task complete
$ gitscrum tasks update GS-1234 --status done
Task GS-1234 marked as Done

# Log any final time
$ gitscrum timer log GS-1234 30m -d "Code review fixes"
Logged 30m to GS-1234
```

---

## Sprint Planning

### Start New Sprint

```bash
# Review team capacity
$ gitscrum standup team
TEAM STATUS
  alice: 32h available, 8 pts in backlog
  bob: 24h available, 5 pts in backlog
  charlie: 40h available, 0 pts assigned

# Create sprint
$ gitscrum sprints create -n "Sprint 13" --start 2026-03-01 --end 2026-03-14
Created: Sprint 13 (Mar 1-14)

# Assign backlog items
$ gitscrum tasks update GS-1301 --sprint sprint-13
$ gitscrum tasks update GS-1302 --sprint sprint-13
$ gitscrum tasks update GS-1303 --sprint sprint-13
```

### Daily Sprint Check

```bash
# Quick burndown
$ gitscrum sprints burndown
52 │●
44 │  ●
36 │    ●───●
28 │          ●
20 │            ●  ← You are here
12 │
 0 └──────────────────
    1  2  3  4  5  6

Remaining: 20 pts | Target: 18 pts
Status: Slightly behind (-2 pts)

# Who needs help?
$ gitscrum standup blockers
BLOCKERS (2):
  [!] GS-1156 — Waiting on API spec (bob)
  [!] GS-1178 — External dependency (alice)
```

---

## Agency & Freelancer

### Time for Invoicing

```bash
# Monthly time report by client/project
$ gitscrum timer report --month -p client-acme
JANUARY 2026 — Acme Corp — 45h 30m

TASK          HOURS    DESCRIPTION
ACME-201      18h 30m  Dashboard implementation
ACME-198      12h 00m  API integration
ACME-156      8h 00m   Bug fixes
ACME-misc     7h 00m   Meetings & support

Billable: 45h 30m × $150 = $6,825.00
```

### Client Deliverables

```bash
# What did we deliver this month?
$ gitscrum tasks --project acme-portal --status done --since 2026-01-01
15 tasks completed | 67 points delivered

# Generate summary
$ gitscrum sprints view sprint-q1-acme --json | jq '.completed'
```

---

## CI/CD Automation

### GitHub Actions: Auto-Update Task on PR

```yaml
# .github/workflows/task-sync.yml
name: Sync Task Status
on:
  pull_request:
    types: [opened, closed, merged]

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - name: Extract task code
        id: task
        run: |
          TASK=$(echo "${{ github.head_ref }}" | grep -oE '[A-Z]+-[0-9]+')
          echo "code=$TASK" >> $GITHUB_OUTPUT

      - name: Update task status
        if: steps.task.outputs.code
        run: |
          curl -sL https://cli.gitscrum.com/install.sh | sh
          
          if [ "${{ github.event.action }}" == "opened" ]; then
            gitscrum tasks update ${{ steps.task.outputs.code }} --status "in review"
          elif [ "${{ github.event.pull_request.merged }}" == "true" ]; then
            gitscrum tasks update ${{ steps.task.outputs.code }} --status "done"
          fi
        env:
          GITSCRUM_ACCESS_TOKEN: ${{ secrets.GITSCRUM_TOKEN }}
```

### Deployment Tracking

```bash
# In your deploy script
DEPLOY_START=$(date +%s)

# ... deploy commands ...

DEPLOY_END=$(date +%s)
DURATION=$((DEPLOY_END - DEPLOY_START))
MINUTES=$((DURATION / 60))

gitscrum timer log $TASK_CODE "${MINUTES}m" -d "Production deployment"
```

---

## Recording Tips (Asciinema)

```bash
# Install
brew install asciinema  # macOS
sudo apt install asciinema  # Ubuntu

# Record a workflow demo
asciinema rec -t "GitScrum Daily Workflow" demo.cast

# Play it back
asciinema play demo.cast

# Upload to share
asciinema upload demo.cast
```

**Pro tips:**
- Use `--idle-time-limit 2` to cap pauses
- Add `export PS1="$ "` for clean prompt
- Script your demo beforehand
- Use `pv` for realistic typing: `echo "gitscrum tasks" | pv -qL 10`
