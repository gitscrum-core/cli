# Hooks

Git hooks integration for GitScrum.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum hooks install` | Install git hooks |
| `gitscrum hooks uninstall` | Remove git hooks |
| `gitscrum hooks status` | Show installed hooks |

---

## Real-World Scenarios

### Installing Hooks

```bash
$ gitscrum hooks install
Installing git hooks...

✓ prepare-commit-msg  Prepends task code to commit messages
✓ pre-push           Validates task status before push

Hooks installed to .git/hooks/
```

### How It Works

After installing hooks, your commits automatically include task codes:

```bash
$ git checkout feature/GS-1234-oauth-flow
$ git commit -m "Add login endpoint"

# Commit message becomes:
# [GS-1234] Add login endpoint
```

### Check Hook Status

```bash
$ gitscrum hooks status
Git Hooks Status

HOOK               STATUS      DESCRIPTION
prepare-commit-msg Installed   Prepends task code
pre-push           Installed   Validates task status
commit-msg         Not installed
```

### Remove Hooks

```bash
$ gitscrum hooks uninstall
Removing git hooks...

✓ prepare-commit-msg removed
✓ pre-push removed

Hooks uninstalled.
```

---

## Available Hooks

| Hook | Description |
|:-----|:------------|
| `prepare-commit-msg` | Prepends task code from branch name |
| `pre-push` | Validates task exists and is not blocked |
| `commit-msg` | Validates commit message format |

---

## Configuration

Configure hooks in `.gitscrum.yml`:

```yaml
hooks:
  # Prepend task code to commit messages
  prepend_task_code: true
  
  # Commit message format
  # First %s = task code, second %s = original message
  commit_format: "[%s] %s"
  
  # Validate task exists before commit
  validate_task: false
  
  # Block commits if task is blocked
  block_on_blocker: false
```

---

## Tips

- **Team consistency**: Install hooks via `gitscrum init` for all developers
- **Skip when needed**: Use `git commit --no-verify` to bypass hooks
- **Check before push**: Pre-push hook prevents pushing blocked tasks
