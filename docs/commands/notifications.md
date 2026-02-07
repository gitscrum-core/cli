# Notifications

View and manage notifications.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum notifications` | List notifications |
| `gitscrum notifications read` | Mark all as read |

---

## Real-World Scenarios

### Check Notifications

```bash
$ gitscrum notifications
NOTIFICATIONS (5 unread)

● 2h ago   Task assigned: GS-1234 "Implement OAuth"
● 3h ago   @bob mentioned you in GS-1198
● 5h ago   Sprint 12 started
○ 1d ago   GS-1156 marked as blocker
○ 2d ago   Comment on GS-1089

● = unread, ○ = read
```

### Clear Notifications

```bash
$ gitscrum notifications read
Marked 5 notifications as read
```

### Filter by Type

```bash
$ gitscrum notifications --type mentions
MENTIONS

● 3h ago   @bob mentioned you in GS-1198:
           "Hey @alice, can you review this?"

○ 1d ago   @charlie mentioned you in #dev:
           "@alice what do you think about the new API?"
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `--type` | Filter: mentions, assigned, comments, all |
| `-n, --limit` | Number of notifications (default: 10) |
| `--unread` | Show only unread |
| `--json` | Output as JSON |

---

## Tips

- **Morning check**: Start your day by reviewing notifications
- **Stay focused**: Check periodically instead of constantly
- **Filter noise**: Use `--type mentions` for important items
