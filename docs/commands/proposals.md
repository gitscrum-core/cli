# Proposals

Create and manage project proposals.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum proposals` | List all proposals |
| `gitscrum proposals view <uuid>` | View proposal details |
| `gitscrum proposals create <title>` | Create new proposal |
| `gitscrum proposals send <uuid>` | Send proposal to client |
| `gitscrum proposals convert <uuid>` | Convert accepted proposal to project |

---

## Real-World Scenarios

### List Proposals

```bash
$ gitscrum proposals
PROPOSALS

  [->] PRP-2026-015 - Phase 2 Development
     Client: Global Bank
     Amount: $120000.00
     Status: sent

  [D] PRP-2026-014 - MVP Development
     Client: NewStartup
     Amount: $45000.00
     Status: draft

  [OK] PRP-2026-013 - E-commerce Redesign
     Client: RetailCo
     Amount: $65000.00
     Status: accepted
```

### View Proposal

```bash
$ gitscrum proposals view a1b2c3d4
PROPOSAL PRP-2026-015
--------------------------------------------------

Title: Phase 2 Development
Client: Global Bank
Status: [->] sent
Amount: $120000.00
Expires: 2026-02-28
```

### Create Proposal

```bash
$ gitscrum proposals create "Mobile App Development" --client newstartup --amount 80000
Proposal created: PRP-2026-016
  Title: Mobile App Development
  Client: NewStartup
```

### Send to Client

```bash
$ gitscrum proposals send a1b2c3d4
Proposal PRP-2026-016 sent to client
```

### Convert to Project

```bash
$ gitscrum proposals convert a1b2c3d4
Proposal converted to project: Mobile App Development
  Project slug: mobile-app-development
```

---

## Parameters

### proposals create

| Flag | Description |
|:-----|:------------|
| `--client` | Client slug (required) |
| `--amount` | Proposal amount |

---

## Tips

- **Pipeline review**: Use `gitscrum crm pipeline` for pending proposals
- **Follow up**: Check sent proposals status
- **Convert fast**: Use `convert` after client accepts to create project immediately
