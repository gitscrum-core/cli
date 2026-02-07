# Wiki

Access project wiki pages from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum wiki` | List wiki pages |
| `gitscrum wiki view SLUG` | View wiki page content |
| `gitscrum wiki search QUERY` | Search wiki |

---

## Real-World Scenarios

### Finding Documentation

```bash
$ gitscrum wiki
PROJECT WIKI — Backend API

SLUG              TITLE                   UPDATED
architecture      System Architecture     Feb 5, 2026
api-reference     API Reference           Feb 3, 2026
deployment        Deployment Guide        Jan 28, 2026
onboarding        Developer Onboarding    Jan 15, 2026
troubleshooting   Troubleshooting Guide   Jan 10, 2026
```

### Quick Reference

```bash
$ gitscrum wiki view api-reference
API Reference

Authentication:
  All endpoints require Bearer token in Authorization header.
  Token obtained via POST /auth/login

Endpoints:
  GET  /api/v1/users         List users
  GET  /api/v1/users/:id     Get user
  POST /api/v1/users         Create user
  ...

Rate Limits:
  100 requests/minute per token
```

### Search

```bash
$ gitscrum wiki search "database connection"
Search Results — "database connection"

1. deployment.md
   "...set DATABASE_URL environment variable for database connection..."

2. troubleshooting.md
   "...if database connection fails, check firewall rules..."

3. architecture.md
   "...uses connection pooling for database connection management..."
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `-p, --project` | Project slug |
| `--json` | Output as JSON |
| `--markdown` | Raw markdown output |

---

## Tips

- **Quick lookups**: Use wiki from terminal when coding
- **Pipe to less**: `gitscrum wiki view deployment | less` for long pages
- **Search first**: Use search before browsing for faster results
