# Authentication

Secure OAuth authentication for GitScrum CLI.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum auth login` | Initiate OAuth login |
| `gitscrum auth logout` | Clear stored credentials |
| `gitscrum auth status` | Check authentication status |
| `gitscrum auth whoami` | Show authenticated user |

---

## Real-World Scenarios

### First Time Setup

```bash
$ gitscrum auth login
! GitScrum CLI wants to authenticate with your account.

Press Enter to open the browser...

✓ Logged in successfully!

Tip: Run 'gitscrum config set workspace <slug>' to set your default workspace
```

The CLI uses [OAuth 2.0 Device Authorization Grant](https://datatracker.ietf.org/doc/html/rfc8628):

1. CLI generates a device code
2. You authorize in your browser
3. CLI securely stores the access token locally

Credentials are stored in `~/.gitscrum/token.json` with restricted permissions.

### Checking Auth Status

```bash
$ gitscrum auth status
✓ Authenticated

Workspace: acme-corp
Project:   backend-api
Token:     Valid (expires in 29 days)
```

### Who Am I?

```bash
$ gitscrum auth whoami
Logged in as Alice Smith (alice@example.com)
```

### Logging Out

```bash
$ gitscrum auth logout
✓ Logged out successfully
```

### CI/CD Authentication

Interactive login isn't possible in CI/CD. Use environment variables:

```bash
# Export your access token
export GITSCRUM_ACCESS_TOKEN="your-oauth-access-token"

# Now all commands authenticate automatically
gitscrum tasks list
```

To get your access token:
1. Run `gitscrum auth login` locally
2. Copy `access_token` from `~/.gitscrum/token.json`
3. Add as secret in your CI/CD platform

```bash
# Extract token
cat ~/.gitscrum/token.json | jq -r '.access_token'
```

---

## Security

| Aspect | Protection |
|:-------|:-----------|
| **Authentication** | OAuth 2.0 Device Grant — no password transmission |
| **Token Storage** | Local filesystem with `0600` permissions |
| **Token Scope** | Minimum required permissions |
| **CI/CD** | Environment variable injection |

### Token Location

| Platform | Path |
|:---------|:-----|
| Linux/macOS | `~/.gitscrum/token.json` |
| Windows | `%USERPROFILE%\.gitscrum\token.json` |

### Never Commit Tokens

Add to `.gitignore`:

```gitignore
.gitscrum/
~/.gitscrum/
```

---

## Troubleshooting

### Token Expired

```bash
$ gitscrum tasks
Error: token expired. Run 'gitscrum auth login' to re-authenticate

$ gitscrum auth login
✓ Logged in successfully!
```

### Authentication Failed in CI

Check that `GITSCRUM_ACCESS_TOKEN` is correctly set:

```bash
# Verify token is available
echo $GITSCRUM_ACCESS_TOKEN | head -c 20
# Should show: eyJhbGciOiJSUzI1...
```

---

## Tips

- **Stay logged in**: Tokens are long-lived; you rarely need to re-authenticate
- **Multiple accounts**: The CLI supports one account at a time
- **Security first**: Never share your access token or commit it to repositories
