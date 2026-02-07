# Changelog

All notable changes to GitScrum CLI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project configuration via `.gitscrum.yml`
- `GITSCRUM_ACCESS_TOKEN` environment variable for CI/CD
- GitHub Actions, GitLab CI, and Bitbucket Pipelines examples
- Cross-platform install script (`install.sh`)
- Comprehensive automation examples

### Changed
- Improved error messages with actionable suggestions
- Enhanced help text for all commands

### Fixed
- Timer not detecting task from branch name

## [1.0.0] - 2026-XX-XX

### Added
- Initial release
- OAuth Device Flow authentication
- Task management commands (list, view, create, update)
- Time tracking (start, stop, log, report)
- Sprint analytics (current, burndown)
- Project listing
- Git integration (branch detection, task code extraction)
- Multiple output formats (table, JSON, quiet)
- Shell completion (bash, zsh, fish, PowerShell)
- Configuration management

### Security
- Secure token storage in user home directory
- OAuth 2.0 authentication

[Unreleased]: https://github.com/gitscrum-core/cli/compare/v1.0.0...HEAD
[1.0.0]: https://github.com/gitscrum-core/cli/releases/tag/v1.0.0
