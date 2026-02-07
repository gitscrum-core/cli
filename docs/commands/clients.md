# Clients

Client management for agencies and freelancers.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum clients` | List all clients |
| `gitscrum clients view <slug>` | View client details |
| `gitscrum clients create <name>` | Create a new client |
| `gitscrum clients stats` | Show client statistics |
| `gitscrum clients projects <slug>` | List client's projects |

---

## Real-World Scenarios

### List Clients

```bash
$ gitscrum clients
CLIENTS:

  [active] TechStart Inc
     Company: TechStart
     john@techstart.com
     3 projects | $45000.00

  [active] Global Bank
     Company: Global Bank Holdings
     contact@globalbank.com
     2 projects | $120000.00

  [at-risk] RetailCo
     Company: Retail Corporation
     1 projects | $28000.00
```

### View Client

```bash
$ gitscrum clients view techstart
TechStart Inc
──────────────────────────────────────────────────

Status: [active] active
Company: TechStart
Email: john@techstart.com
Phone: +1 555-123-4567

Projects: 3
Total Revenue: $45000.00
```

### Create Client

```bash
$ gitscrum clients create "New Client Co" --email client@newco.com --company "New Client Corp"
Client created: New Client Co
  Slug: new-client-co
```

### Client Statistics

```bash
$ gitscrum clients stats
CLIENT STATISTICS
──────────────────────────────────────────────────

Total Clients:  24
Active:         18
At Risk:        4
Churned:        2

Total Revenue:  $450000.00
MRR:            $35000.00
```

### Client Projects

```bash
$ gitscrum clients projects techstart
PROJECTS FOR techstart:

  • Web Application (web-app)
    Budget: $25000.00
  • Mobile App (mobile-app)
    Budget: $15000.00
  • API v2 (api-v2)
    Budget: $5000.00
```

---

## Parameters

### clients create

| Flag | Description |
|:-----|:------------|
| `-e, --email` | Client email |
| `--phone` | Client phone |
| `-c, --company` | Company name |

---

## Tips

- **Quick lookup**: View client details before calls
- **Health check**: Use `stats` for portfolio overview
- **Cross-reference**: Use `projects` to see client's active work
