<p align="center">
  <img src="https://site-assets.gitscrum.com/vscode/gitscrum-white.png" alt="GitScrum" width="160"/>
</p>

<h1 align="center">GitScrum CLI</h1>

<p align="center">
  Command-line interface for GitScrum.<br/>
  Manage projects, track time, and ship faster — without leaving your terminal.
</p>

<p align="center">
  <a href="https://github.com/gitscrum-core/cli/releases"><img src="https://img.shields.io/github/v/release/gitscrum-core/cli?style=flat-square&color=000" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/gitscrum-core/cli"><img src="https://img.shields.io/badge/go_report-A+-000?style=flat-square" alt="Go Report"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-000?style=flat-square" alt="MIT License"></a>
  <a href="https://github.com/gitscrum-core/cli/actions"><img src="https://img.shields.io/badge/tests-passing-000?style=flat-square" alt="Tests"></a>
</p>

<br/>

## Overview

GitScrum CLI gives you full operational access to your [GitScrum](https://gitscrum.com) workspace from the command line. Manage tasks, sprints, time tracking, user stories, epics, kanban workflows, team discussions, wiki, notes, client CRM, invoicing, proposals, budget tracking, analytics dashboards, standup reports, and activity feeds.

Everything your team does in the GitScrum, you can now do from your terminal or CI/CD pipeline.

```
$ gitscrum tasks current
GS-1234: Implement user authentication
Status: In Progress | Sprint: Sprint 12 | Effort: 5 pts

$ gitscrum timer start
Timer started for GS-1234 at 09:15

$ gitscrum sprints burndown
Sprint 12 Burndown (Day 5 of 10)
║████████████░░░░░░░░░░░░║ 48% complete
Remaining: 26 pts | Velocity: 5.2 pts/day | On track

$ gitscrum timer stop
Logged 2h 15m to GS-1234
```

**Git-aware. Team-ready. CI/CD compatible.**

---

## Quick start

### Install

```bash
curl -sL https://raw.githubusercontent.com/gitscrum-core/cli/main/install.sh | sh
```

Or install via package managers:

<details>
<summary><strong>macOS (Homebrew)</strong></summary>

```bash
brew install gitscrum-core/tap/gitscrum
```
</details>

<details>
<summary><strong>Windows (Scoop)</strong></summary>

```powershell
scoop bucket add gitscrum https://github.com/gitscrum-core/scoop-bucket
scoop install gitscrum
```
</details>

<details>
<summary><strong>Go Install</strong></summary>

```bash
go install github.com/gitscrum-core/cli/cmd/gitscrum@latest
```
</details>

### Authenticate

```bash
gitscrum auth login
```

The CLI initiates an [OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628) flow. Authorize in your browser — credentials are stored locally and never transmitted to third parties.

### Configure

```bash
gitscrum config set workspace my-company
gitscrum config set project my-project
```

---

## Commands

### Core

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum tasks` | List, view, create, and update tasks | [tasks](docs/commands/tasks.md) |
| `gitscrum timer` | Start, stop, and log time entries | [timer](docs/commands/timer.md) |
| `gitscrum sprints` | Sprint management and analytics | [sprints](docs/commands/sprints.md) |
| `gitscrum projects` | Project listing and details | [projects](docs/commands/projects.md) |

### Time Tracking

| Command | Description |
|:--------|:------------|
| `gitscrum timer start [TASK]` | Start timer for a task |
| `gitscrum timer stop` | Stop active timer |
| `gitscrum timer log TASK DURATION` | Log time manually (e.g., `2h30m`) |
| `gitscrum timer report` | View time reports |

### Tasks

| Command | Description |
|:--------|:------------|
| `gitscrum tasks` | List assigned tasks |
| `gitscrum tasks view CODE` | View task details |
| `gitscrum tasks create TITLE` | Create new task |
| `gitscrum tasks current` | Show task for current git branch |
| `gitscrum tasks branch CODE` | Create git branch from task |

### Sprints

| Command | Description |
|:--------|:------------|
| `gitscrum sprints` | List all sprints |
| `gitscrum sprints current` | Current sprint with KPIs |
| `gitscrum sprints burndown` | ASCII burndown chart |

### Authentication

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum auth login` | Initiate OAuth flow | [auth](docs/commands/auth.md) |
| `gitscrum auth logout` | Clear stored credentials | |
| `gitscrum auth status` | Check authentication status | |
| `gitscrum auth whoami` | Show authenticated user | |

Full reference: [docs/commands/](docs/commands/README.md)

---

## Git Integration

The CLI automatically detects task codes from your git branch:

```bash
$ git checkout -b feature/GS-1234-user-auth
$ gitscrum tasks current
GS-1234: Implement user authentication
```

Create branches from tasks:

```bash
$ gitscrum tasks branch GS-1234
Created and switched to: feature/GS-1234-implement-user-auth
```

---

## Project Configuration

Create a `.gitscrum.yml` in your repository root to share settings across your team:

```bash
gitscrum init -w my-workspace -p my-project
```

```yaml
version: "1"
workspace: my-company
project: my-project

branch:
  default_prefix: feature
  include_title: true
  max_length: 60

hooks:
  prepend_task_code: true
  commit_format: "[%s] %s"

automation:
  on_pr_merge: done
  complete_on_merge: true
```

Full options: [docs/examples/.gitscrum.yml](docs/examples/.gitscrum.yml)

---

## CI/CD Integration

For headless environments, authenticate using an access token:

```bash
export GITSCRUM_ACCESS_TOKEN="your-oauth-access-token"
gitscrum tasks list
```

### Examples

| Platform | Examples |
|:---------|:---------|
| GitHub Actions | [github-actions/](docs/examples/github-actions/) |
| GitLab CI | [gitlab-ci/](docs/examples/gitlab-ci/) |
| Bitbucket Pipelines | [bitbucket-pipelines/](docs/examples/bitbucket-pipelines/) |

Common use cases:
- Sync PR/MR status with task status
- Auto-link commits to tasks
- Generate sprint reports on schedule
- Track deployments

---

## Output Formats

```bash
gitscrum tasks              # Table output (default)
gitscrum tasks --json       # JSON for scripting
gitscrum tasks -q           # Quiet mode (IDs only)
```

---

## Security

The CLI is designed around the **principle of least privilege**.

| Layer | Protection |
|:------|:-----------|
| **Authentication** | OAuth 2.0 Device Grant — credentials never touch remote servers. |
| **Token storage** | Local filesystem with restricted permissions (`~/.gitscrum/`). |
| **CI/CD** | Environment variable authentication for headless environments. |

Found a vulnerability? Report privately to **security@gitscrum.com**.

Full details: [SECURITY.md](SECURITY.md)

---

## Documentation

| | |
|:--|:--|
| **[Commands Reference](docs/COMMANDS.md)** | All commands with parameters and examples |
| **[Configuration Guide](docs/CONFIGURATION.md)** | CLI and project configuration options |
| **[CI/CD Examples](docs/examples/README.md)** | GitHub Actions, GitLab CI, Bitbucket Pipelines |
| **[Security](SECURITY.md)** | Security model and token handling |
| **[Contributing](CONTRIBUTING.md)** | Development setup and contribution guidelines |
| **[Changelog](CHANGELOG.md)** | Version history |

---

## Development

```bash
git clone https://github.com/gitscrum-core/cli.git
cd cli
make build
make test
```

| Requirement | Version |
|:------------|:--------|
| Go | >= 1.21 |

Full guide: [CONTRIBUTING.md](CONTRIBUTING.md)

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

```bash
git checkout -b feature/my-feature
# make changes, add tests
make test
git commit -m "feat: describe your change"
```

---

## License

MIT — see [LICENSE](LICENSE).

---

<p align="center">
  <a href="https://gitscrum.com">Website</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="https://docs.gitscrum.com/en/cli">Docs</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="https://github.com/gitscrum-core/cli/issues">Issues</a>&nbsp;&nbsp;·&nbsp;&nbsp;<a href="CHANGELOG.md">Changelog</a>
</p>

<p align="center">
  Built by <a href="https://gitscrum.com">GitScrum</a>
</p>
