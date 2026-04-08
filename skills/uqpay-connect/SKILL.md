---
name: uqpay-connect
version: 1.0.0
description: "UQPAY Connect API: create and manage connected sub-accounts for marketplace/platform models. Use when the user needs to onboard sub-merchants, create sub-accounts, upload KYC documents, or operate on behalf of connected accounts."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay account --help"
---

# uqpay-connect (v1)

**CRITICAL — read [`../uqpay-shared/SKILL.md`](../uqpay-shared/SKILL.md) first for auth, config, global flags, and data conventions.**

## Core Concepts

- **Connected Account** — a sub-merchant or partner account under your platform. Has its own balances, beneficiaries, cards.
- **Entity Type** — `COMPANY` (business KYC) or `INDIVIDUAL` (personal KYC). Determines required fields.
- **On-Behalf-Of** — after creating a connected account, use `--on-behalf-of <account_id>` on any banking/issuing/payment command to operate as that account.
- **Sub-Account** — a further nested account under a connected account (for complex hierarchies).
- **Additional Documents** — supplementary KYC documents uploaded after account creation.
- **File Upload** — use `uqpay file upload` to upload documents, then reference the file ID in account creation.

## Resource Relationships

```
Platform (Master Account)
└── Connected Account (COMPANY | INDIVIDUAL)
    ├── Sub-Account
    ├── Additional Documents
    └── (all banking/issuing/payment resources via --on-behalf-of)
```

## Important Notes

- Account creation payloads are the most complex in the API — use the [onboarding reference](references/uqpay-connect-onboarding.md)
- `@+filepath` (data URI format) required for document fields
- No special prefix needed for most fields — phone numbers, postal codes, registration numbers are strings by default
- **Auto number coercion:** `tos_agreement` is auto-converted for both `account create` and `create-sub`. Additionally, `create-sub` auto-converts `inherit`, `internationally`, `ownership_percentage`
- After account creation, status is `PROCESSING` — verification happens asynchronously
- Once verified, use `--on-behalf-of` to operate as the connected account across all domains
- **Run `uqpay account <action> -h`** for complete parameter lists. The `-h` output is the source of truth.

## Command Reference

| Action | Command | Type |
|--------|---------|------|
| list | `uqpay account list [--page-num N --page-size N]` | GET |
| get | `uqpay account get <account_id>` | GET |
| create | `uqpay account create -d ...` | POST — **[see onboarding guide](references/uqpay-connect-onboarding.md) for required fields** |
| create-sub | `uqpay account create-sub -d ...` | POST — **[see onboarding guide](references/uqpay-connect-onboarding.md) for required fields** |
| additional-documents | `uqpay account additional-documents --country SG --business-code BANKING` | GET |

## Cross-Domain Usage

After account creation, operate on behalf of the connected account:

```bash
# Banking operations for sub-account
uqpay --on-behalf-of <account_id> banking balance list

# Issuing operations for sub-account
uqpay --on-behalf-of <account_id> issuing card list

# Payment operations for sub-account
uqpay --on-behalf-of <account_id> payment balance list
```

## Workflows

- [Account Onboarding](references/uqpay-connect-onboarding.md) — full onboarding flow for COMPANY and INDIVIDUAL accounts
