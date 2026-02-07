# Analytics

Sprint and project analytics from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum analytics` | Overview of analytics |
| `gitscrum analytics sprint` | Sprint analytics and KPIs |
| `gitscrum analytics velocity` | Team velocity over time |
| `gitscrum analytics time` | Time tracking analytics |

---

## Real-World Scenarios

### Sprint Health Check

```bash
$ gitscrum analytics sprint
Sprint 12 Analytics — February Release

Progress:
  Tasks:    67% complete (16/24)
  Points:   71% complete (37/52)
  
Burndown:
  Ideal:    26 pts remaining
  Actual:   15 pts remaining
  Status:   Ahead of schedule

Velocity:
  Current:  5.3 pts/day
  Average:  4.8 pts/day
  
Blockers:  1 task blocking 3 others
```

### Team Velocity Trend

```bash
$ gitscrum analytics velocity
Team Velocity — Last 6 Sprints

SPRINT    POINTS    COMPLETED    VELOCITY
Sprint 7  45        43           4.3
Sprint 8  48        48           4.8
Sprint 9  52        50           5.0
Sprint 10 50        52           5.2
Sprint 11 55        53           5.3
Sprint 12 52        37*          5.3*

* In progress

Average Velocity: 5.0 pts/sprint
Trend: ↑ Improving (+12% over 6 sprints)
```

### Time Analytics

```bash
$ gitscrum analytics time --period this-month
Time Analytics — February 2026

Total Hours:    312h logged
Billable:       285h (91%)
Non-billable:   27h

By Project:
  backend-api   142h (45%)
  web-app       98h (31%)
  mobile-app    72h (23%)

By Day:
  Mon   █████████ 68h
  Tue   ████████  62h
  Wed   ████████  64h
  Thu   ███████   58h
  Fri   ██████    48h
  Sat   ██        12h
  Sun   ─         0h

Peak Hours: 9am-12pm, 2pm-5pm
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Filter by project |
| `-s, --sprint` | Specific sprint slug |
| `--period` | Time period (this-week, this-month, last-30-days) |
| `--json` | Output as JSON |

---

## Tips

- **Daily standup**: Use `gitscrum analytics sprint` to get quick KPIs
- **Retrospectives**: Compare velocity across sprints
- **Invoicing**: Use time analytics for billing reports
- **JSON export**: Pipe to visualization tools or dashboards
