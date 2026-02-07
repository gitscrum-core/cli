# Security Policy

## Supported Versions

| Version | Supported |
|:--------|:----------|
| 1.x.x   | Yes |
| < 1.0   | No |

---

## Reporting a Vulnerability

**Do not** create a public GitHub issue for security vulnerabilities.

Email: **security@gitscrum.com**

Include:
- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (optional)

---

## Response Timeline

| Stage | Timeline |
|:------|:---------|
| Acknowledgment | 48 hours |
| Initial assessment | 7 days |
| Resolution | 30-90 days (severity dependent) |

You will be credited in release notes unless you prefer anonymity.

---

## Scope

**In scope:**
- GitScrum CLI application
- Authentication and token handling
- Configuration file security
- API communication

**Out of scope:**
- GitScrum web application (report separately)
- Third-party dependencies (report to maintainers)

---

## Security Model

### Authentication

The CLI uses [OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628):

1. CLI requests device code from GitScrum
2. User authorizes in browser
3. CLI polls for access token
4. Token stored locally with restricted permissions

Credentials are never transmitted through the CLI.

### Token Storage

| Platform | Location | Permissions |
|:---------|:---------|:------------|
| Linux/macOS | `~/.gitscrum/token.json` | `0600` |
| Windows | `%USERPROFILE%\.gitscrum\token.json` | User-only |

### CI/CD Authentication

For headless environments, use `GITSCRUM_ACCESS_TOKEN` environment variable:

```bash
export GITSCRUM_ACCESS_TOKEN="your-token"
```

Store tokens as secrets in your CI/CD platform. Never commit tokens to repositories.

---

## Best Practices

1. **Keep CLI updated** — Check releases for security patches
2. **Protect tokens** — Never commit `~/.gitscrum/` to version control
3. **Use secrets** — Store `GITSCRUM_ACCESS_TOKEN` in CI/CD secrets
4. **Verify downloads** — Download only from official sources

---

## Security Updates

Security fixes are released as patch versions and documented in [CHANGELOG.md](CHANGELOG.md).

Subscribe to releases to receive notifications:
1. Go to [github.com/gitscrum-core/cli](https://github.com/gitscrum-core/cli)
2. Click "Watch" → "Custom" → "Releases"
