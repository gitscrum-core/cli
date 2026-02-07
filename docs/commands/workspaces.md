# Workspaces

List and manage workspaces.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum workspaces` | List all workspaces |
| `gitscrum workspaces view SLUG` | View workspace details |

---

## Real-World Scenarios

### Discovering Your Workspaces

```bash
$ gitscrum workspaces
SLUG          NAME             ROLE        PROJECTS
acme-corp     Acme Corp        Owner       12
freelance     Freelance Work   Member      3
opensource    OS Contributions Member      8
```

### Workspace Details

```bash
$ gitscrum workspaces view acme-corp
Acme Corp

Plan:    Business
Members: 15
Projects: 12

Active Sprints:
  Backend API    Sprint 12 (67% complete)
  Web App        Sprint 11 (45% complete)
  Mobile App     Sprint 10 (82% complete)

This Week:
  Tasks completed: 47
  Time logged:     152h
```

### Switching Workspaces

```bash
# Set default workspace
$ gitscrum config set workspace freelance

# Now all commands use this workspace
$ gitscrum projects
# Shows projects from freelance workspace
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `--json` | Output as JSON |
| `-q, --quiet` | Output only slugs |

---

## Tips

- **Multiple workspaces**: You can be a member of many workspaces
- **Set default**: Use `gitscrum config set workspace <slug>` to avoid typing `-w` every time
- **Per-repo config**: Use `.gitscrum.yml` to set workspace per repository
