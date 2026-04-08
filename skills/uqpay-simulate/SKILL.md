---
name: uqpay-simulate
version: 1.0.0
description: "UQPAY Simulator: sandbox-only commands to simulate deposits, card authorizations, and reversals for testing. Use when the user needs to test payment flows without real money in the sandbox environment."
metadata:
  requires:
    bins: ["uqpay"]
  cliHelp: "uqpay simulate --help"
---

# uqpay-simulate (v1)

**CRITICAL — read [`../uqpay-shared/SKILL.md`](../uqpay-shared/SKILL.md) first for auth, config, global flags, and data conventions.**

**Sandbox only.** These commands only work with `--env sandbox`.

## Core Concepts

- **Simulate Deposit** — creates an inbound deposit to your banking balance (triggers deposit webhook)
- **Simulate Authorization** — triggers a card authorization on an active issuing card (triggers transaction webhook)
- **Simulate Reversal** — reverses a previously simulated authorization

## Command Reference

| Action | Command | Description |
|--------|---------|-------------|
| deposit | `uqpay simulate deposit -d ...` | Simulate inbound deposit |
| authorization | `uqpay simulate authorization -d ...` | Simulate card authorization |
| reversal | `uqpay simulate reversal <transaction_id>` | Reverse a simulated authorization |

## Simulate Deposit

Add funds to your banking balance for testing:

```bash
uqpay simulate deposit \
  -d amount=100 \
  -d currency=SGD \
  -d sender_swift_code=WELGBE22
```

Verify with:
```bash
uqpay banking balance list
uqpay banking deposit list
```

## Simulate Card Authorization

Test card transactions on an active issuing card:

```bash
uqpay simulate authorization \
  -d card_id=<card_id> \
  -d transaction_amount=10 \
  -d transaction_currency=USD \
  -d "merchant_name=Test Store" \
  -d merchant_category_code=5734
```

Verify with:
```bash
uqpay issuing transaction list
uqpay issuing card get <card_id>     # check available_balance
```

## Simulate Reversal

Reverse a previously simulated authorization:

```bash
# Get the transaction_id from the authorization
uqpay issuing transaction list -o json | jq '.data[0].transaction_id'

# Reverse it
uqpay simulate reversal <transaction_id>
```

## Testing Recipes

### Full Card Flow Test

```bash
# 1. Fund the issuing balance
uqpay simulate deposit -d amount=1000 -d currency=SGD -d sender_swift_code=WELGBE22

# 2. Create and recharge a card (see uqpay-issuing skill)
uqpay issuing card recharge <card_id> -d amount=100 -d currency=SGD

# 3. Simulate a purchase
uqpay simulate authorization \
  -d card_id=<card_id> \
  -d transaction_amount=25 \
  -d transaction_currency=SGD \
  -d "merchant_name=Coffee Shop" \
  -d merchant_category_code=5812

# 4. Check transaction appeared
uqpay issuing transaction list

# 5. Reverse if needed
uqpay simulate reversal <transaction_id>
```

### Deposit Notification Test

```bash
# 1. Simulate deposit
uqpay simulate deposit -d amount=500 -d currency=SGD -d sender_swift_code=WELGBE22

# 2. Verify balance updated
uqpay banking balance get SGD

# 3. Check deposit record
uqpay banking deposit list
```
