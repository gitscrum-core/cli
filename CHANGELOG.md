# Changelog

All notable changes to GitScrum CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.0-beta] - 2026-02-09

First public beta release! 🎉

### Added
- **Authentication**: OAuth 2.0 Device Flow for secure login
- **Task Management**: List, view, create, update, and complete tasks
- **Time Tracking**: Start/stop timers, log time manually, view reports
- **Sprint Analytics**: Current sprint KPIs, burndown charts (ASCII)
- **Git Integration**: Branch detection, task code extraction from branch names
- **Daily Standups**: View completed tasks, blockers, team status
- **Project Management**: List projects, view details, switch contexts
- **Team Collaboration**: Chat, wiki, notifications
- **Client & Billing**: Manage clients, invoices, proposals
- **CRM**: Analytics and reports (PRO feature)
- **Configuration**: Project config via `.gitscrum.yml`
- **CI/CD Support**: `GITSCRUM_ACCESS_TOKEN` for headless environments
- **Multiple Output Formats**: Table, JSON, quiet mode
- **Shell Completion**: Bash, Zsh, Fish, PowerShell
- **Cross-platform**: Windows, macOS, Linux

### Examples Included
- GitHub Actions workflows
- GitLab CI pipelines
- Bitbucket Pipelines
- Git hooks (commit-msg, prepare-commit-msg)
- Automation scripts

### Security
- OAuth 2.0 Device Authorization Grant (RFC 8628)
- Secure token storage with restricted permissions
- No secrets or credentials in codebase

[Unreleased]: https://github.com/gitscrum-core/cli/compare/v1.0.0-beta...HEAD
[1.0.0-beta]: https://github.com/gitscrum-core/cli/releases/tag/v1.0.0-beta
