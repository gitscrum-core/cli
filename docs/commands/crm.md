# CRM

ClientFlow dashboard for agency owners - strategic view of clients, revenue, and projects.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum crm` | Dashboard overview |
| `gitscrum crm revenue` | Revenue pipeline |
| `gitscrum crm at-risk` | Clients at risk |
| `gitscrum crm pipeline` | Pending approvals |
| `gitscrum crm projects` | Projects health |
| `gitscrum crm leaderboard` | Client leaderboard |

---

## Real-World Scenarios

### Dashboard Overview

```bash
$ gitscrum crm
CRM DASHBOARD
════════════════════════════════════════════════════════════

SUMMARY
  Clients: 24 | Active Projects: 18 | Total Projects: 42

INVOICES
  Paid: 45 ($125,000.00) | Pending: 8 ($32,500.00)
  [!] Overdue: 3 ($8,750.00)

PROPOSALS
  Approved: 12 ($285,000.00) | Pending: 5 ($67,500.00)
  [!] Expiring Soon: 2

ALERTS
  [!] 3 overdue invoices
  [!] 2 expiring proposals
```

### Revenue Pipeline

```bash
$ gitscrum crm revenue
REVENUE PIPELINE
────────────────────────────────────────────────────────────

INVOICES BY STATUS:
  paid:        45 invoices  $125,000.00
  pending:     8 invoices   $32,500.00
  overdue:     3 invoices   $8,750.00

PROPOSALS BY STATUS:
  approved:    12 proposals $285,000.00
  pending:     5 proposals  $67,500.00

OVERDUE INVOICES:
  [!] INV-2024-047 - TechStart Inc: $3,500.00 (15 days overdue)
  [!] INV-2024-051 - Global Bank: $2,750.00 (8 days overdue)

MONTHLY REVENUE:
  2026-01: $42,500.00 (12 invoices)
  2025-12: $38,200.00 (10 invoices)
```

### Clients at Risk

```bash
$ gitscrum crm at-risk
CLIENTS AT RISK
────────────────────────────────────────────────────────────

Total at risk: 4
  With overdue invoices: 2
  With stalled projects: 1
  With expiring proposals: 1

[!] TechStart Inc
    - Overdue invoice: $3,500.00 (15 days)

[!] MedTech Solutions
    - Project stalled: No activity in 14 days
```

### Projects Health

```bash
$ gitscrum crm projects
PROJECTS HEALTH
────────────────────────────────────────────────────────────

Total: 18 | Healthy: 12 | Warning: 4 | Critical: 2
Over Budget: 3 | At Risk: 2

[!!] API Integration (Global Bank)
    Progress: 45% | Budget: 120% used (96.0h / 80.0h)

[!] Web Redesign (TechStart Inc)
    Progress: 68% | Budget: 85% used (42.5h / 50.0h)
```

### Client Leaderboard

```bash
$ gitscrum crm leaderboard
CLIENT LEADERBOARD
────────────────────────────────────────────────────────────

 1. Global Bank
    Revenue: $145,000.00 | Projects: 3 active / 8 total
    Reliability: 95% | Avg Payment: 12 days

 2. TechStart Inc
    Revenue: $87,500.00 | Projects: 2 active / 5 total
    Reliability: 88% | Avg Payment: 18 days

 3. MedTech Solutions
    Revenue: $62,300.00 | Projects: 1 active / 3 total
    Reliability: 92% | Avg Payment: 15 days
```

---

## Access Control

CRM commands require `agency_owner` role. Other roles (manager, developer, client) do not have access to ClientFlow data.

---

## Tips

- **Morning review**: Run `gitscrum crm` to see alerts and outstanding items
- **Collections focus**: Use `gitscrum crm at-risk` to identify clients needing attention
- **Project oversight**: Use `gitscrum crm projects` to catch budget overruns early
- **Revenue tracking**: Use `gitscrum crm revenue` to monitor cash flow
