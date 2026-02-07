# Invoices

Invoice management and billing.

---

## Commands

| Command | Description |
|:--------|:------------|
| `gitscrum invoices` | List all invoices |
| `gitscrum invoices view <uuid>` | View invoice details |
| `gitscrum invoices create` | Create a new invoice |
| `gitscrum invoices send <uuid>` | Send invoice to client |
| `gitscrum invoices mark-paid <uuid>` | Mark invoice as paid |

---

## Real-World Scenarios

### List Invoices

```bash
$ gitscrum invoices
INVOICES:

  [paid] INV-2026-042 - USD 12500.00
     Client: TechStart Inc
     Due: 2026-01-15 • Status: paid

  [sent] INV-2026-041 - USD 28000.00
     Client: Global Bank
     Due: 2026-02-15 • Status: sent

  [!!] INV-2026-040 - USD 8500.00
     Client: RetailCo
     Due: 2026-02-01 • Status: overdue
```

### View Invoice

```bash
$ gitscrum invoices view a1b2c3d4
Invoice INV-2026-041
──────────────────────────────────────────────────

Client: Global Bank
Amount: USD 28000.00
Status: [sent] sent
Due Date: 2026-02-15
```

### Create Invoice

```bash
$ gitscrum invoices create --client techstart --amount 12500 --due 2026-03-15
Invoice created: INV-2026-043
  Amount: USD 12500.00
  Client: TechStart Inc
```

### Send to Client

```bash
$ gitscrum invoices send a1b2c3d4
Invoice INV-2026-043 sent to client
```

### Mark as Paid

```bash
$ gitscrum invoices mark-paid a1b2c3d4
Invoice INV-2026-043 marked as paid
```

---

## Parameters

### invoices create

| Flag | Description |
|:-----|:------------|
| `--client` | Client slug (required) |
| `--amount` | Invoice amount (required) |
| `--currency` | Currency (default: USD) |
| `--due` | Due date (YYYY-MM-DD) |

---

## Tips

- **End of month**: Generate invoices from project work
- **Follow up**: Check overdue invoices with `gitscrum crm` dashboard
- **Quick mark paid**: Use `mark-paid` when payment received
