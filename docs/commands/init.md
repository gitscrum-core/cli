# Init

Initialize GitScrum in a repository.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum init` | Create `.gitscrum.yml` configuration |

---

## Real-World Scenarios

### Setting Up a New Repository

```bash
$ cd my-project
$ gitscrum init -w acme-corp -p backend-api

Created .gitscrum.yml:
  workspace: acme-corp
  project: backend-api

Commit this file to share settings with your team.
```

### Full Configuration

```bash
$ gitscrum init --full -w acme-corp -p backend-api

Created .gitscrum.yml with all options:
  - Branch naming conventions
  - Timer auto-start/stop
  - Git hooks configuration
  - Automation rules

Review and customize before committing.
```

### Team Onboarding

Once `.gitscrum.yml` is committed, new team members get automatic configuration:

```bash
$ git clone https://github.com/acme/backend-api.git
$ cd backend-api

# Already configured!
$ gitscrum tasks
GS-1234  Implement OAuth flow  In Progress
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `-w, --workspace` | Workspace slug |
| `-p, --project` | Project slug |
| `--minimal` | Create minimal config (workspace + project only) |
| `--full` | Create config with all options |
| `--force` | Overwrite existing `.gitscrum.yml` |

---

## Generated File

### Minimal (default)

```yaml
version: "1"
workspace: acme-corp
project: backend-api
```

### Full

```yaml
version: "1"
workspace: acme-corp
project: backend-api

branch:
  default_prefix: feature
  include_title: true
  max_length: 60

timer:
  auto_start: false
  auto_stop: false
  round_to: 15

hooks:
  prepend_task_code: true
  commit_format: "[%s] %s"

automation:
  on_pr_merge: done
  complete_on_merge: true
```

---

## Tips

- **Commit the file**: `.gitscrum.yml` is meant to be version controlled
- **Start minimal**: Add options as you need them
- **Team alignment**: Everyone uses the same configuration
