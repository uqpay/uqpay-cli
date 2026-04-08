---
name: uqpay-banking
version: 1.0.0
description: "UQPAY Banking API: virtual accounts, balances, beneficiaries, payouts, conversions, exchange rates, transfers, deposits. Covers cross-border payments, FX conversion workflows, and multi-currency treasury operations."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay banking --help"
---

# uqpay-banking (v1)

**CRITICAL: Read [uqpay-shared](../uqpay-shared/SKILL.md) first.** It covers configuration, authentication, global flags, data passing conventions (auto number coercion, `@filepath`, dot notation), output formatting, pagination, and error handling. This skill assumes you already know that material.

## Core Concepts

- **Virtual Account** — a named account that receives inbound payments; each has a unique account number and routing details
- **Balance** — currency-level balance within the banking wallet; tracks available and pending amounts per currency
- **Beneficiary** — a registered recipient (person or company) with verified bank details; must be created before sending payouts
- **Payout** — an outbound payment to a beneficiary; supports same-currency and cross-currency transfers
- **Conversion** — a foreign exchange transaction that converts one currency to another at a quoted rate
- **Exchange Rate** — indicative FX rates between currency pairs; use `conversion quote` for a binding rate
- **Transfer** — an internal transfer between UQPAY accounts (e.g., main account to sub-account)
- **Deposit** — a record of an inbound payment received into a virtual account or directly into the banking wallet

## Resource Relationships

```
Banking Wallet
├── Balance (per currency)
│   └── Transactions (ledger entries)
├── Virtual Account (receives deposits)
│   └── Deposit (inbound payment)
├── Beneficiary (registered recipient)
│   └── Payout (outbound payment to beneficiary)
├── Conversion (FX between balances)
│   └── Exchange Rate / Quote
└── Transfer (internal movement between accounts)
```

## Important Notes

- **No special prefix needed for numeric strings** — amounts, account numbers, and postal codes are sent as strings by default: `-d amount=500`, `-d bank_details.account_number=12345678999`, `-d address.postal_code=10001`
- **Beneficiary `entity_type`** — `INDIVIDUAL` requires `first_name` + `last_name`; `COMPANY` requires `company_name`. Other required fields differ by type. See [beneficiary create reference](references/uqpay-banking-beneficiary-create.md).
- **Beneficiary `payment_method`** — `SWIFT` requires `swift_code` in bank_details; `LOCAL` requires routing codes (`routing_code_type1` + `routing_code_value1`). The clearing_system field must match.
- **`--on-behalf-of`** — pass a sub-account ID to execute banking operations on behalf of a connected account. See [uqpay-connect](../uqpay-connect/SKILL.md).
- **Conversion quotes expire in 75 seconds** — create the conversion promptly after quoting.
- **Payout dates** — `payout_date` must be a valid business day in the payout corridor.
- **Run `uqpay banking <resource> <action> -h`** for complete parameter lists. The `-h` output is the source of truth.

## Command Reference

| Command | Method | Description |
|---------|--------|-------------|
| `uqpay banking virtual-account list` | GET (flags) | List virtual accounts |
| `uqpay banking virtual-account create` | POST (-d) | Create a virtual account |
| `uqpay banking balance list` | GET (flags) | List balances across all currencies |
| `uqpay banking balance get <currency>` | GET | Get balance for a specific currency |
| `uqpay banking balance transactions` | GET (flags) | List balance transactions (ledger entries) |
| `uqpay banking transfer list` | GET (flags) | List transfers |
| `uqpay banking transfer create` | POST (-d) | Create an internal transfer |
| `uqpay banking transfer get <id>` | GET | Get transfer details |
| `uqpay banking deposit list` | GET (flags) | List deposits |
| `uqpay banking deposit get <id>` | GET | Get deposit details |
| `uqpay banking conversion list` | GET (flags) | List conversions |
| `uqpay banking conversion get <id>` | GET | Get conversion details |
| `uqpay banking conversion quote` | POST (-d) | Get a binding FX quote (valid 75s) |
| `uqpay banking conversion create` | POST (-d) | Execute a conversion from a quote |
| `uqpay banking exchange-rate list` | GET (flags) | List indicative exchange rates |
| `uqpay banking beneficiary list` | GET (flags) | List beneficiaries |
| `uqpay banking beneficiary get <id>` | GET | Get beneficiary details |
| `uqpay banking beneficiary create` | POST (-d) | Create a beneficiary — **[see reference](references/uqpay-banking-beneficiary-create.md) for required fields** |
| `uqpay banking beneficiary update <id>` | POST (-d) | Update a beneficiary |
| `uqpay banking beneficiary delete <id>` | POST | Delete a beneficiary |
| `uqpay banking beneficiary check` | POST (-d) | Check if a beneficiary bank account is valid |
| `uqpay banking beneficiary payment-methods` | GET (flags) | List available payment methods for a corridor |
| `uqpay banking conversion dates` | GET (flags) | List available conversion dates |
| `uqpay banking payout list` | GET (flags) | List payouts |
| `uqpay banking payout get <id>` | GET | Get payout details |
| `uqpay banking payout create` | POST (-d) | Create a payout — **[see payout flow](references/uqpay-banking-payout-flow.md) for full workflow** |

## Workflows

- **Payout flow** (create beneficiary, quote FX, convert, pay out): [references/uqpay-banking-payout-flow.md](references/uqpay-banking-payout-flow.md)
- **Beneficiary creation** (entity types, payment methods, required fields): [references/uqpay-banking-beneficiary-create.md](references/uqpay-banking-beneficiary-create.md)
