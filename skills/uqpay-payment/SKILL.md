---
name: uqpay-payment
version: 1.0.0
description: "UQPAY Payment API: payment intents, payment attempts, refunds, settlements, balances, payouts, and bank accounts. Use when the user needs to accept payments, manage payment lifecycle (create/confirm/capture/cancel), process refunds, view settlements, or manage payment bank accounts."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay payment --help"
---

# uqpay-payment (v1)

**CRITICAL — read [`../uqpay-shared/SKILL.md`](../uqpay-shared/SKILL.md) first for auth, config, global flags, and data conventions.**

## Core Concepts

- **Payment Intent** — a payment session representing a customer's intent to pay. Contains amount, currency, return URL. Expires after 30 minutes. ID prefix: `PI`.
- **Payment Attempt** — each customer try within an intent. May redirect for 3DS. ID prefix: `PA`.
- **Capture** — claiming authorized funds. Automatic by default; manual capture available.
- **Refund** — returning funds to the customer (full or partial).
- **Settlement** — periodic fund sweeps from captured payments to your balance.
- **Payment Balance** — funds available from settled payments.
- **Payment Payout** — withdrawing from payment balance to a bank account.
- **Bank Account** — registered bank account for receiving payouts.

## Resource Relationships

```
Payment Intent (PI...)
├── Payment Attempt (PA...)
├── Capture (auto or manual)
├── Refund
└── Settlement

Payment Balance
└── Payment Payout
    └── Bank Account
```

## Payment Intent Lifecycle

```
REQUIRES_PAYMENT_METHOD → REQUIRES_CUSTOMER_ACTION (3DS) → REQUIRES_CAPTURE → PENDING → SUCCEEDED
                                                                                      → FAILED
                                                                                      → CANCELLED
```

## Important Notes

- `merchant_order_id` must be unique per intent (used as idempotency key)
- Amounts are sent as strings by default (e.g. `-d amount=100`)
- Intent expires 30 minutes after creation
- `confirm` triggers actual charge and may redirect for 3DS
- `capture` only needed for manual capture mode
- `cancel` releases held funds (use `-d cancellation_reason=requested_by_customer`)
- `--on-behalf-of` supported for connected account payments

## Command Reference

| Resource | Action | Command | Type |
|----------|--------|---------|------|
| intent | list | `uqpay payment intent list [--page-num N --page-size N]` | GET |
| intent | get | `uqpay payment intent get <id>` | GET |
| intent | create | `uqpay payment intent create -d ...` | POST |
| intent | update | `uqpay payment intent update <id> -d ...` | POST |
| intent | confirm | `uqpay payment intent confirm <id> -d ...` | POST |
| intent | capture | `uqpay payment intent capture <id> -d ...` | POST |
| intent | cancel | `uqpay payment intent cancel <id> -d ...` | POST |
| attempt | list | `uqpay payment attempt list [--page-num N --page-size N]` | GET |
| attempt | get | `uqpay payment attempt get <id>` | GET |
| refund | list | `uqpay payment refund list [--page-num N --page-size N]` | GET |
| refund | get | `uqpay payment refund get <id>` | GET |
| refund | create | `uqpay payment refund create -d ...` | POST |
| settlement | list | `uqpay payment settlement list [--page-num N --page-size N]` | GET |
| balance | list | `uqpay payment balance list` | GET |
| balance | get | `uqpay payment balance get <currency>` | GET |
| bank-account | list | `uqpay payment bank-account list` | GET |
| bank-account | get | `uqpay payment bank-account get <id>` | GET |
| bank-account | create | `uqpay payment bank-account create -d ...` | POST |
| bank-account | update | `uqpay payment bank-account update <id> -d ...` | POST |
| payout | list | `uqpay payment payout list` | GET |
| payout | get | `uqpay payment payout get <id>` | GET |
| payout | create | `uqpay payment payout create -d ...` | POST |

## Workflows

- [Payment Intent Flow](references/uqpay-payment-intent-flow.md) — full payment lifecycle from creation to refund
