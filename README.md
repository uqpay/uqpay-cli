# uqpay-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.21-blue.svg)](https://go.dev/)
[![npm version](https://img.shields.io/npm/v/@uqpay/cli.svg)](https://www.npmjs.com/package/@uqpay/cli)

The official [UQPAY](https://www.uqpay.com/) CLI tool — built for humans and AI Agents. Covers Banking, Card Issuing, and Payment APIs with 100+ commands and 6 AI Agent [Skills](./skills/).

[Install](#installation) · [Quick Start](#quick-start) · [AI Agent Skills](#agent-skills) · [Commands](#command-structure) · [Advanced](#advanced-usage) · [Contributing](#contributing)

## Why uqpay-cli?

- **Agent-Native Design** — 6 structured [Skills](./skills/) out of the box, AI Agents can operate UQPAY APIs with zero extra setup
- **Wide Coverage** — 3 business domains (Banking, Issuing, Payment), 100+ commands, Connect & Simulator included
- **Zero Type Hassle** — All values default to strings, number fields auto-converted per API spec — just type `-d amount=100` and it works
- **Open Source** — MIT license, `npm install` and go
- **Up and Running in 1 Minute** — Set credentials, start calling APIs
- **Cross-Platform** — macOS, Linux, Windows (amd64 & arm64)

## Features

| Domain | Capabilities |
|--------|-------------|
| 🏦 Banking | Balances, transfers, deposits, payouts, beneficiaries, FX conversions, exchange rates, virtual accounts |
| 💳 Issuing | Card products, cardholders, virtual/physical cards, recharge/withdraw, transactions, spending controls, reports |
| 💰 Payment | Payment intents, attempts, refunds, settlements, bank accounts, payouts |
| 🔗 Connect | Connected accounts, sub-account onboarding (COMPANY/INDIVIDUAL), KYC document upload |
| 🧪 Simulator | Sandbox-only: simulate deposits, card authorizations, reversals |
| 📁 File | Upload files, get download links |

## Installation

### From npm (recommended)

```bash
npm install -g @uqpay/cli
```

### From source

Requires Go 1.21+.

```bash
git clone https://github.com/uqpay/uqpay-cli.git
cd uqpay-cli
make install
```

### Enable tab completion (optional)

```bash
uqpay setup-completion
source ~/.zshrc  # or ~/.bashrc
```

## Quick Start

```bash
# 1. Configure credentials (one-time)
uqpay config set client-id YOUR_CLIENT_ID
uqpay config set api-key YOUR_API_KEY
uqpay config set env sandbox

# 2. Start using
uqpay banking balance list
uqpay issuing card list
uqpay payment intent list
```

Or inline without config file:

```bash
uqpay --env sandbox --client-id ID --api-key KEY banking balance list
```

## Quick Start (AI Agent)

```bash
# Step 1 — Install
npm install -g @uqpay/cli

# Step 2 — Configure (provide credentials from UQPAY dashboard)
uqpay config set client-id <CLIENT_ID>
uqpay config set api-key <API_KEY>
uqpay config set env sandbox

# Step 3 — Verify
uqpay banking balance list
```

## Agent Skills

Located in [`skills/`](./skills/), these provide structured guides for AI Agents to operate UQPAY APIs correctly.

| Skill | Description |
|-------|-------------|
| [`uqpay-shared`](./skills/uqpay-shared/SKILL.md) | Configuration, authentication, global flags, data conventions, auto number coercion |
| [`uqpay-banking`](./skills/uqpay-banking/SKILL.md) | Balances, beneficiaries, payouts, conversions, transfers, deposits, virtual accounts |
| [`uqpay-issuing`](./skills/uqpay-issuing/SKILL.md) | Cards, cardholders, products, transactions, transfers, reports |
| [`uqpay-payment`](./skills/uqpay-payment/SKILL.md) | Payment intents, attempts, refunds, settlements, bank accounts, payouts |
| [`uqpay-connect`](./skills/uqpay-connect/SKILL.md) | Connected accounts, sub-account onboarding, `--on-behalf-of` operations |
| [`uqpay-simulate`](./skills/uqpay-simulate/SKILL.md) | Sandbox testing: deposits, authorizations, reversals |

## Command Structure

```
uqpay <domain> <resource> <action> [flags]
```

### Domains

```bash
uqpay banking balance list           # Banking API
uqpay issuing card list              # Issuing API
uqpay payment intent list            # Payment API
```

### Other Commands

```bash
uqpay account list                   # Connected accounts (Connect API)
uqpay config get                     # CLI configuration
uqpay file upload ./doc.pdf          # File upload
uqpay simulate deposit -d ...        # Sandbox simulator
uqpay setup-completion               # Install shell tab-completion
```

### Shortcuts

Top-level aliases for common resources:

```bash
uqpay beneficiary list               # = uqpay banking beneficiary list
uqpay card list                      # = uqpay issuing card list
uqpay payout list                    # = uqpay banking payout list
uqpay conversion list                # = uqpay banking conversion list
uqpay cardholder list                # = uqpay issuing cardholder list
uqpay exchange-rate list             # = uqpay banking exchange-rate list
```

## Data Passing

**Read operations** use flags, **write operations** use `-d key=value`:

```bash
# Read (GET)
uqpay banking beneficiary list --page-num 2 --page-size 20

# Write (POST) — dot notation for nested objects
uqpay banking beneficiary create \
  -d entity_type=INDIVIDUAL \
  -d first_name=John \
  -d last_name=Doe \
  -d bank_details.swift_code=DBSSSGSG \
  -d bank_details.account_number=1234567890
```

All values are strings by default. Number fields (like `amount` in card recharge) are auto-converted — no prefix needed.

## Advanced Usage

### Output Formats

```bash
uqpay banking balance list              # table (default)
uqpay banking balance list -o json      # JSON
uqpay banking balance list -o yaml      # YAML
```

### Debug Mode

```bash
uqpay --debug banking balance list      # Print HTTP request/response details
```

### Sub-Account Operations

```bash
uqpay --on-behalf-of <account-id> banking balance list
```

### File Encoding

```bash
# Pure base64 (issuing identity documents)
-d identity.front_file=@./passport.jpg

# Data URI (connect account documents)
-d "certification[0]=@+./cert.png"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `UQPAY_CLIENT_ID` | Override client ID |
| `UQPAY_API_KEY` | Override API key |
| `UQPAY_ENV` | Override environment (sandbox/production) |
| `UQPAY_OUTPUT` | Override output format |

## Contributing

Contributions welcome! Please submit an [Issue](https://github.com/uqpay/uqpay-cli/issues) or [Pull Request](https://github.com/uqpay/uqpay-cli/pulls).

## License

[MIT](./LICENSE)
