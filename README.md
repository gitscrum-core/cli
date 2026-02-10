<p align="center">
  <img width="508" height="129" alt="image" src="https://github.com/user-attachments/assets/56204676-18a3-4e30-8b19-7048f56bc2b7" />
</p>

<h1 align="center">GitScrum CLI: Status follows Code ⚡</h1>

<p align="center">
  Official GitScrum CLI: Git-native task management and time tracking for the terminal..<br/>
  Git-native task management for developers who live in the terminal. Zero context switching. Automation-first. Built to keep you in the flow.
</p>

<p align="center">
  <a href="https://github.com/gitscrum-core/cli/releases"><img src="https://img.shields.io/badge/status-BETA-orange?style=flat-square" alt="Beta"></a>
  <a href="https://github.com/gitscrum-core/cli/releases"><img src="https://img.shields.io/github/v/release/gitscrum-core/cli?style=flat-square&color=000" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/gitscrum-core/cli"><img src="https://goreportcard.com/badge/github.com/gitscrum-core/cli" alt="Go Report"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-000?style=flat-square" alt="MIT License"></a>
  <a href="https://github.com/gitscrum-core/cli/actions/workflows/ci.yml"><img src="https://github.com/gitscrum-core/cli/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
</p>

## Overview

GitScrum CLI gives you full access to your [GitScrum](https://gitscrum.com) workspace from the terminal. Tasks, time tracking, sprints, analytics — everything without leaving your editor.

<p align="center">
<img src="https://github.com/user-attachments/assets/c189d98a-78a2-4f1b-b6fd-d2f6a56487b9" />
</p>

> **⚠️ BETA**: This CLI is in active development. While fully functional, you may encounter bugs or breaking changes. [Report issues](https://github.com/gitscrum-core/cli/issues) to help us improve!

```bash
# Your morning in 30 seconds
$ gitscrum tasks
CODE      TITLE                              STATUS         EFFORT
GS-1234   Implement OAuth flow               In Progress    5 pts
GS-1198   Fix pagination bug                 Todo           2 pts
GS-1201   Add caching layer                  Blocked        8 pts

$ gitscrum timer start GS-1234
Timer started for GS-1234 at 09:15

# ... 3 hours of coding ...

$ gitscrum timer stop
Logged 3h 12m to GS-1234
  Total today: 5h 45m

$ gitscrum sprints burndown
Sprint 12 — Day 7 of 10
52 │●
40 │  ●──●
30 │       ●──●
20 │            ●  ← You are here
10 │
 0 └─────────────────
    M  T  W  T  F

Remaining: 22 pts | Velocity: 5.2/day | On track
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
<summary><strong>macOS / Linux (Homebrew)</strong></summary>

```bash
brew tap gitscrum-core/homebrew-tap
brew install gitscrum
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
| `gitscrum standup` | Daily standup and blockers | [standup](docs/commands/standup.md) |
| `gitscrum analytics` | Sprint and project analytics | [analytics](docs/commands/analytics.md) |

### Time Tracking

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum timer start [TASK]` | Start timer for a task | [timer](docs/commands/timer.md) |
| `gitscrum timer stop` | Stop active timer | [timer](docs/commands/timer.md) |
| `gitscrum timer log TASK DURATION` | Log time manually (e.g., `2h30m`) | [timer](docs/commands/timer.md) |
| `gitscrum timer report` | View time reports | [timer](docs/commands/timer.md) |

### Tasks

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum tasks` | List assigned tasks | [tasks](docs/commands/tasks.md) |
| `gitscrum tasks view CODE` | View task details | [tasks](docs/commands/tasks.md) |
| `gitscrum tasks create TITLE` | Create new task | [tasks](docs/commands/tasks.md) |
| `gitscrum tasks current` | Show task for current git branch | [tasks](docs/commands/tasks.md) |
| `gitscrum tasks branch CODE` | Create git branch from task | [tasks](docs/commands/tasks.md) |

### Sprints

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum sprints` | List all sprints | [sprints](docs/commands/sprints.md) |
| `gitscrum sprints current` | Current sprint with KPIs | [sprints](docs/commands/sprints.md) |
| `gitscrum sprints burndown` | ASCII burndown chart | [sprints](docs/commands/sprints.md) |

### Team & Collaboration

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum chat` | Team discussions | [chat](docs/commands/chat.md) |
| `gitscrum wiki` | Project documentation | [wiki](docs/commands/wiki.md) |
| `gitscrum notifications` | View notifications | [notifications](docs/commands/notifications.md) |

### Clients & Billing

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum clients` | Manage clients | [clients](docs/commands/clients.md) |
| `gitscrum invoices` | Manage invoices | [invoices](docs/commands/invoices.md) |
| `gitscrum proposals` | Manage proposals | [proposals](docs/commands/proposals.md) |

### Configuration

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum auth login` | Initiate OAuth flow | [auth](docs/commands/auth.md) |
| `gitscrum auth logout` | Clear stored credentials | [auth](docs/commands/auth.md) |
| `gitscrum auth status` | Check authentication status | [auth](docs/commands/auth.md) |
| `gitscrum config` | Manage CLI configuration | [config](docs/commands/config.md) |
| `gitscrum init` | Initialize project configuration | [init](docs/commands/init.md) |
| `gitscrum workspaces` | List and switch workspaces | [workspaces](docs/commands/workspaces.md) |
| `gitscrum hooks` | Install/uninstall git hooks | [hooks](docs/commands/hooks.md) |

### CRM (PRO)

| Command | Description | Docs |
|:--------|:------------|:-----|
| `gitscrum crm` | CRM analytics and reports | [crm](docs/commands/crm.md) |

Full reference: [docs/commands/](docs/commands/README.md)

## Real-World Workflows

### Branch → Code → PR → Done

```bash
# Pick your task and create a branch
$ gitscrum tasks branch GS-1234
Switched to branch 'feature/GS-1234-implement-oauth'

# Timer auto-starts when you checkout a task branch
$ gitscrum timer
Active: GS-1234 | Running: 45m

# After PR is opened
$ gitscrum tasks update GS-1234 --status "in review"
GS-1234 moved to In Review

# After merge
$ gitscrum tasks update GS-1234 --status done
GS-1234 marked as Done
```

### Daily Standup

```bash
$ gitscrum standup
DAILY STANDUP — Feb 7, 2026
────────────────────────────────────────
  Completed yesterday:    4 tasks
  In progress:            2 tasks
  Blocked:                1 task
  
$ gitscrum standup blockers
BLOCKERS:
  [!] GS-1156 — Waiting on API spec (bob)
```

### Team Time Report

```bash
$ gitscrum timer report --week --team
TEAM TIME — This Week

MEMBER      HOURS     TOP PROJECT
alice       38h 15m   backend-api
bob         32h 00m   web-app  
charlie     28h 45m   mobile-app

Total: 99h 00m | Avg: 33h/person
```

### Git Integration

```bash
# Branch names auto-detect task codes
$ git checkout feature/GS-1234-oauth
$ gitscrum tasks current
GS-1234: Implement OAuth flow | In Progress | 5 pts

# Create branch from task
$ gitscrum tasks branch GS-1198
Created and switched to: fix/GS-1198-pagination-bug
```

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