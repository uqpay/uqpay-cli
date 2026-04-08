---
name: uqpay-issuing
version: 1.0.0
description: "UQPAY Issuing API: card program management, cardholder KYC, virtual/physical card issuance, spending controls, balance operations, and transaction monitoring."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay issuing --help"
---

# uqpay-issuing (v1)

**CRITICAL: Read [uqpay-shared](../uqpay-shared/SKILL.md) first.** It covers configuration, authentication, global flags, `-d` data passing, `@filepath` encoding, dot notation, array indexing, pagination, and output formats. This skill assumes you know all of that.

## Core Concepts

- **Products** — Card templates defined by your card program. Each product specifies the BIN, currency, card form (virtual/physical), card scheme (VISA/MASTERCARD), and mode type (SINGLE/SHARE). You cannot create products via CLI; they are pre-configured. Use `product list` to discover available products.

- **Cardholders** — KYC identity records. A cardholder represents a real person with name, date of birth, contact info, address, and optionally identity documents. Cardholders must be created before cards can be issued to them.

- **Cards** — Virtual or physical cards linked to a cardholder and a product. Cards hold a balance, can have spending/risk controls, and go through a lifecycle of status transitions. Card creation is asynchronous.

- **Card Orders** — When a card is created, a card order tracks the async provisioning status (PROCESSING, SUCCESS, FAILED). Poll the order to know when the card is ready.

- **Spending & Risk Controls** — Per-card rules: transaction amount limits (per-transaction, daily, weekly, monthly, yearly, all-time), 3DS settings, and MCC whitelists/blacklists.

- **Secure Card Data** — Sensitive card details (PAN, CVV, expiry) accessible via `get-secure` (requires special permission) or `iframe-url` (embeddable secure iframe URL as fallback).

## Resource Relationships

```
Product (card template)
└── Card
    ├── Cardholder (KYC identity)
    ├── Card Order (async status)
    └── Transaction

Balance
└── Transaction (recharge/withdraw/auth)

Transfer (between issuing accounts)
Report (settlement/ledger)
```

## Important Notes

- **Card creation is async.** `card create` returns a `card_id` and `card_order_id` with `order_status: PROCESSING`. Poll with `card get-order --card-id <id>` until status becomes `SUCCESS` or `FAILED`.

- **Card status transitions:**
  - `PENDING` → `ACTIVE` → `FROZEN` (reversible, can go back to `ACTIVE`) / `CANCELLED` (terminal, irreversible)
  - Physical cards may start in `PENDING` and require activation

- **`get-secure` may need special permission.** If your account lacks access, use `iframe-url` instead to get a secure iframe URL for displaying card details.

- **No prefix needed** — phone numbers, PAN, PIN, postal codes are strings by default. Amount fields (`card_limit`, `spending_controls[n].amount`, recharge/withdraw `amount`) are auto-converted to numbers by the CLI.

- **Use `@filepath` for identity documents** in cardholder create/update. Example: `-d identity.front_file=@./passport.jpg` encodes the file as pure base64.

- **Card modes:**
  - `SINGLE` — one card per cardholder, standard use case
  - `SHARE` — multiple cards can share a cardholder, requires `card_limit` on creation

## Command Reference

| Resource | Action | Description |
|----------|--------|-------------|
| `balance` | `list` | List all issuing balances by currency |
| `balance` | `get <currency>` | Get balance for a specific currency |
| `balance` | `transactions` | List balance transactions (recharge, withdraw, auth) |
| `cardholder` | `list` | List cardholders with optional filters |
| `cardholder` | `get <id>` | Get cardholder details by ID |
| `cardholder` | `create` | Create a new cardholder with KYC info |
| `cardholder` | `update <id>` | Update cardholder information |
| `card` | `list` | List cards with optional filters |
| `card` | `get <id>` | Get card details by ID |
| `card` | `create` | Create a new card (async) |
| `card` | `update <id>` | Update card metadata/controls |
| `card` | `update-status <id>` | Change card status (ACTIVE, FROZEN, CANCELLED) |
| `card` | `get-secure <id>` | Get sensitive card data (PAN, CVV, expiry) |
| `card` | `iframe-url <id>` | Get secure iframe URL for card details |
| `card` | `get-order` | Check card order status by card ID |
| `card` | `recharge <id>` | Load funds onto a card |
| `card` | `withdraw <id>` | Withdraw funds from a card |
| `card` | `activate` | Activate a physical card |
| `card` | `assign` | Assign a card to a cardholder |
| `card` | `set-pin` | Set or reset card PIN |
| `transaction` | `list` | List card transactions with filters |
| `transaction` | `get <id>` | Get transaction details by ID |
| `product` | `list` | List available card products |
| `transfer` | `create` | Create a transfer between issuing accounts |
| `transfer` | `get <id>` | Get transfer details by ID |
| `report` | `create` | Request a settlement/ledger report |
| `report` | `download <id>` | Download a generated report |

Use `uqpay issuing <resource> <action> -h` for full parameter details on any command.

## Workflows

See the [references/](references/) directory for step-by-step guides:

- [Card Lifecycle](references/uqpay-issuing-card-lifecycle.md) — end-to-end flow from product selection through card creation, funding, status management, and cancellation
- [Card Creation Details](references/uqpay-issuing-card-create.md) — detailed parameter reference for card creation including KYC supplementation, spending controls, and risk controls
