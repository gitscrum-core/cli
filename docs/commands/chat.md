# Chat

Team chat and discussions from your terminal.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum chat` | List recent messages |
| `gitscrum chat send MESSAGE` | Send a message |
| `gitscrum chat channels` | List channels |

---

## Real-World Scenarios

### Checking Messages

```bash
$ gitscrum chat
#general — Recent Messages

bob (2h ago):
  Deployed v2.3.0 to staging. Please test!

alice (1h ago):
  @bob Tested, looks good. Ready for prod.

charlie (30m ago):
  Anyone available for a quick code review?
```

### Send a Quick Message

```bash
$ gitscrum chat send "PR #42 is ready for review"
Message sent to #general
```

### Switch Channels

```bash
$ gitscrum chat --channel dev
#dev — Recent Messages

alice (4h ago):
  FYI: Changed the API response format for /users endpoint

bob (3h ago):
  @alice Sounds good, updated the client accordingly
```

---

## Parameters

| Flag | Description |
|:-----|:------------|
| `--channel` | Channel name (default: general) |
| `-n, --limit` | Number of messages (default: 10) |
| `--json` | Output as JSON |

---

## Tips

- **Stay in flow**: Check and send messages without leaving terminal
- **CI notifications**: Send build status to chat from pipelines
- **Quick updates**: Share progress without context switching
