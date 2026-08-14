---
name: uqpay-shared
version: 1.0.0
description: "UQPAY CLI foundation: configuration, authentication, global flags, data passing conventions (auto number coercion, @filepath, dot notation), output formatting, pagination, error handling. Read this first before using any domain skill."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay --help"
---

# uqpay-shared (v1)

This is the foundation skill. All domain skills depend on it.

## Configuration

Config file: `~/.uqpay/config.yaml`

```bash
# Set config values
uqpay config set client-id <your-client-id>
uqpay config set api-key <your-api-key>
uqpay config set env sandbox          # sandbox | production
uqpay config set output table         # table | json | yaml

# View current config
uqpay config get
```

Environment variables override config file:

| Env Var | Config Key |
|---------|-----------|
| `UQPAY_CLIENT_ID` | client-id |
| `UQPAY_API_KEY` | api-key |
| `UQPAY_ENV` | env |
| `UQPAY_OUTPUT` | output |

CLI flags override everything: `--env`, `--client-id`, `--api-key`, `-o`/`--output`.

## Authentication

UQPAY uses OAuth2 client credentials. The CLI handles token lifecycle automatically — no manual token management needed.

- **Sandbox**: `api-sandbox.uqpaytech.com`
- **Production**: `api.uqpay.com`

Quick start:
```bash
uqpay config set client-id YOUR_CLIENT_ID
uqpay config set api-key YOUR_API_KEY
uqpay config set env sandbox
uqpay banking balance list              # test it
```

Or inline (no config file needed):
```bash
uqpay --env sandbox --client-id ID --api-key KEY banking balance list
```

## Global Flags

| Flag | Description |
|------|------------|
| `--env` | Environment: `sandbox` or `production` |
| `--client-id` | Override client ID |
| `--api-key` | Override API key |
| `-o`, `--output` | Output format: `table` (default), `json`, `yaml` |
| `--debug` | Print full HTTP request and response details to stderr |
| `--on-behalf-of` | Sub-account ID for connected account operations (see [uqpay-connect](../uqpay-connect/SKILL.md)) |

## Command Structure

```
uqpay <domain> <resource> <action> [flags]
```

**Domains:**
- `banking` — balances, transfers, payouts, beneficiaries, conversions, exchange rates, deposits, virtual accounts
- `issuing` — cards, cardholders, products, transactions, balances, transfers, reports
- `payment` — intents, attempts, refunds, settlements, balances, payouts, bank accounts

**Other commands:**
- `account` — connected account management (Connect API)
- `config` — CLI configuration
- `file` — file upload and download links
- `simulate` — sandbox-only testing (deposit, authorization, reversal)

**Shortcuts** (top-level aliases for common resources):
- `beneficiary` → `banking beneficiary`
- `conversion` → `banking conversion`
- `exchange-rate` → `banking exchange-rate`
- `payout` → `banking payout`
- `card` → `issuing card`
- `cardholder` → `issuing cardholder`

## Data Passing Conventions

### Read vs Write Split

- **GET operations** (list, get): use `--flags` for query parameters
- **POST/PUT operations** (create, update): use `-d key=value` for request body

### Dot Notation for Nested Objects

```bash
-d bank_details.account_number=123456
-d bank_details.swift_code=DBSSSGSG
-d address.city=Singapore
-d address.country=SG
```

Produces: `{"bank_details": {"account_number": "123456", "swift_code": "DBSSSGSG"}, "address": {"city": "Singapore", "country": "SG"}}`

### Array Indexing

```bash
-d spending_controls[0].amount=500
-d spending_controls[0].interval=PER_TRANSACTION
-d risk_controls.allowed_mcc[0]=5411
-d risk_controls.allowed_mcc[1]=5812
```

Append syntax (no index): `-d payment_method_types[]=card`

### Type Coercion Rules

The CLI auto-coerces values:

| Input | Result | Rule |
|-------|--------|------|
| `USD` | `"USD"` (string) | All values default to string |
| `100` | `"100"` (string) | All values default to string |
| `amount=100` | `100` (number) | Known number fields auto-converted |
| `true`/`false` | `true`/`false` (bool) | Booleans are recognized |

### Automatic Number Coercion

The CLI automatically converts known number-type fields to JSON numbers. Known fields include: `amount`, `transaction_amount`, `card_limit`, `no_pin_payment_amount`, `payout_amount`, `amount_to_capture`, `inherit`, `internationally`, `tos_agreement`, `ownership_percentage`. Just type the value normally:

```bash
-d amount=100                        # 100 (number, auto-converted)
-d inherit=-1                        # -1 (number, auto-converted)
-d tos_agreement=1                   # 1 (number, auto-converted)
-d postal_code=10001                 # "10001" (string, not a known number field)
-d merchant_category_code=5734       # "5734" (string)
```

The exact set of auto-converted fields depends on each command's implementation. Run `uqpay <command> -h` to see which fields are typed as `number`.

Manual overrides `num:` and `str:` still work but are rarely needed — use `num:` to force an unknown field to a number, or `str:` to force a known number field to a string.

### File Encoding with `@`

For fields that expect base64-encoded file content:

```bash
# Pure base64 (for issuing identity documents)
-d identity.front_file=@./passport.jpg

# Data URI format (for connect account documents)
-d "identity_verification.identity_docs[0]=@+./id_front.png"
```

- `@filepath` → pure base64 string
- `@+filepath` → `data:<mime>;base64,<b64>` (data URI with auto-detected MIME type)

## Pagination

List commands support pagination:

```bash
uqpay banking beneficiary list --page-num 1 --page-size 20
```

Default: page 1, 10 results per page. When 10+ results are returned, a hint is shown:
```
(10 results — use --page-num 2 for the next page, --page-size to change limit)
```

## Output Formats

```bash
uqpay banking balance list              # table (default, human-readable)
uqpay banking balance list -o json      # JSON (machine-readable, pipe to jq)
uqpay banking balance list -o yaml      # YAML
```

Use `-o json` when piping to other tools or when you need exact field values.

## Debugging

```bash
uqpay --debug banking balance list
```

Prints to stderr:
- `[DEBUG] → METHOD URL` — request method and URL
- `[DEBUG]   Header: value` — request headers (auth token hidden)
- `[DEBUG]   {json body}` — request body (for POST/PUT)
- `[DEBUG] ← STATUS` — response status code
- `[DEBUG]   {json body}` — response body

## Error Handling

| Exit Code | Meaning |
|-----------|---------|
| 0 | Success |
| 1 | API error (4xx/5xx from server) |
| 2 | Config error (missing credentials, invalid env) |
| 3 | Network error (unreachable, timeout) |
| 4 | Other error |

Common troubleshooting:
- **401 Unauthorized** → check client-id and api-key
- **403 Forbidden** → account lacks permission for this operation
- **"expected text, got number"** → this should not happen with current defaults
- **"field X is required"** → check `-h` for required parameters

## Safety Rules

- API key is masked in `config get` output (shows only last 4 chars)
- Use `--debug` to inspect requests — auth tokens are hidden in debug output
- Always verify destructive operations (delete, cancel) before confirming
- Sandbox (`--env sandbox`) is safe for testing; production changes are real
