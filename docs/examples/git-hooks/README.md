# Git Hooks for GitScrum CLI
#
# This directory contains git hooks that integrate with GitScrum.
# 
# Installation:
#   chmod +x *.sh
#   cp *.sh .git/hooks/
#   # Remove .sh extension after copying
#
# Or use a git hooks manager like Husky:
#   npx husky add .husky/commit-msg './docs/examples/git-hooks/commit-msg.sh'

## Available Hooks

| Hook | Purpose |
|------|---------|
| `commit-msg.sh` | Prepend task code to commit messages |
| `prepare-commit-msg.sh` | Auto-insert task code template |
| `post-checkout.sh` | Start timer on task branch checkout |
| `pre-push.sh` | Validate task exists before push |
| `post-merge.sh` | Complete task after merge to main |

## Quick Setup

```bash
# Copy all hooks
for hook in commit-msg prepare-commit-msg post-checkout pre-push post-merge; do
  cp "docs/examples/git-hooks/${hook}.sh" ".git/hooks/${hook}"
  chmod +x ".git/hooks/${hook}"
done
```
