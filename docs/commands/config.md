# Configuration

CLI configuration and project settings.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum config` | Show current configuration |
| `gitscrum config set KEY VALUE` | Set configuration value |
| `gitscrum init` | Initialize project configuration |

---

## Real-World Scenarios

### Initial Setup After Login

```bash
$ gitscrum auth login
✓ Logged in successfully!

# Set your default workspace
$ gitscrum config set workspace acme-corp

# Set your default project
$ gitscrum config set project backend-api

# Verify configuration
$ gitscrum config
Workspace: acme-corp
Project:   backend-api
API URL:   https://api.gitscrum.com
```

### Setting Up a Repository

```bash
$ cd my-project
$ gitscrum init -w acme-corp -p backend-api

Created .gitscrum.yml:
  workspace: acme-corp
  project: backend-api

This file can be committed to share settings with your team.
```

### Team Configuration

Everyone cloning the repository gets the same defaults:

```bash
$ git clone https://github.com/acme/backend-api.git
$ cd backend-api

# .gitscrum.yml is already configured
$ gitscrum tasks
# Shows tasks from the correct project
```

### Viewing All Settings

```bash
$ gitscrum config
Configuration:

Global (~/.gitscrum/config.yml):
  Workspace: acme-corp
  Project:   (not set)
  API URL:   https://api.gitscrum.com

Project (.gitscrum.yml):
  Workspace: acme-corp
  Project:   backend-api
  Branch:    feature/{code}-{title}
  Hooks:     prepend_task_code=true

Effective:
  Workspace: acme-corp
  Project:   backend-api
```

---

## Configuration Files

### Global Configuration

Location: `~/.gitscrum/config.yml`

```yaml
workspace: acme-corp
project: ""
api_url: https://api.gitscrum.com
```

### Project Configuration

Location: `.gitscrum.yml` (repository root)

```yaml
version: "1"
workspace: acme-corp
project: backend-api

branch:
  default_prefix: feature
  include_title: true
  max_length: 60
  format: "{prefix}/{code}-{title}"

timer:
  auto_start: false
  auto_stop: false
  round_to: 15
  min_duration: 5

hooks:
  prepend_task_code: true
  commit_format: "[%s] %s"
  validate_task: false

automation:
  on_pr_open: in review
  on_pr_merge: done
  complete_on_merge: true
  link_pr_to_task: true
```

---

## Configuration Priority

Settings are applied in order (later overrides earlier):

1. Built-in defaults
2. Global config (`~/.gitscrum/config.yml`)
3. Project config (`.gitscrum.yml`)
4. Environment variables
5. Command-line flags

---

## Environment Variables

| Variable | Description |
|:---------|:------------|
| `GITSCRUM_ACCESS_TOKEN` | OAuth access token (CI/CD) |
| `GITSCRUM_WORKSPACE` | Override workspace |
| `GITSCRUM_PROJECT` | Override project |
| `GITSCRUM_API_URL` | Custom API URL |

---

## Parameters

### config set

| Key | Description |
|:----|:------------|
| `workspace` | Default workspace slug |
| `project` | Default project slug |

### init

| Flag | Description |
|:-----|:------------|
| `-w, --workspace` | Workspace slug |
| `-p, --project` | Project slug |
| `--minimal` | Create minimal config |
| `--full` | Create config with all options |

---

## Tips

- **Repository config**: Commit `.gitscrum.yml` to share settings with your team
- **Global defaults**: Set workspace globally, project per-repository
- **CI/CD override**: Use environment variables in pipelines
- **Minimal config**: Start with workspace/project only, add options as needed
