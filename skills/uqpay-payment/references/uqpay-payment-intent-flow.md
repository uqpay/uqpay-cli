# Payment Intent Flow

> **Prerequisite:** read [`../../uqpay-shared/SKILL.md`](../../uqpay-shared/SKILL.md) first.

Complete payment lifecycle using the CLI.

## Step 1: Create Payment Intent

```bash
uqpay payment intent create \
  -d amount=100 \
  -d currency=USD \
  -d "description=Order #12345" \
  -d merchant_order_id=ORDER-12345 \
  -d return_url=https://yoursite.com/payment/return
```

Returns `payment_intent_id` (e.g. `PI2041756265729757184`) and `client_secret`.

**Key fields:**
- `amount` — payment amount (required)
- `currency` — ISO 4217 (USD, SGD, etc.) (required)
- `merchant_order_id` — unique in your system, acts as idempotency key (required)
- `return_url` — where customer returns after 3DS or redirect (required)
- `description` — shown to customer, max 32 chars

For complete parameter details, run `uqpay payment intent create -h`.

## Step 2: Customer Completes Payment

After creation, the intent status is `REQUIRES_PAYMENT_METHOD`. The customer completes payment on your frontend using the `client_secret`.

For testing via CLI (confirm with payment method):
```bash
uqpay payment intent update <payment_intent_id> \
  -d "payment_method.type=card" \
  -d payment_method.card.number=4000020000000000 \
  -d payment_method.card.exp_month=12 \
  -d payment_method.card.exp_year=2028 \
  -d payment_method.card.cvc=123
```

## Step 3: Check Payment Status

```bash
uqpay payment intent get <payment_intent_id>
```

Check `intent_status`:
- `REQUIRES_CUSTOMER_ACTION` — customer needs to complete 3DS
- `REQUIRES_CAPTURE` — authorized, needs manual capture
- `PENDING` — processing
- `SUCCEEDED` — payment complete
- `FAILED` — payment failed
- `CANCELLED` — payment cancelled

## Step 4: Capture (Manual Capture Mode Only)

Only needed if your intent was created with manual capture:
```bash
# Full capture
uqpay payment intent capture <payment_intent_id>

# Partial capture
uqpay payment intent capture <payment_intent_id> \
  -d amount_to_capture=50
```

## Step 5: Cancel (If Needed)

```bash
uqpay payment intent cancel <payment_intent_id> \
  -d cancellation_reason=requested_by_customer
```

Valid reasons: `duplicate`, `fraudulent`, `requested_by_customer`, `abandoned`

## Step 6: Refund (If Needed)

```bash
# Full refund
uqpay payment refund create \
  -d payment_intent_id=<payment_intent_id> \
  -d amount=100

# Partial refund
uqpay payment refund create \
  -d payment_intent_id=<payment_intent_id> \
  -d amount=30
```

Multiple partial refunds are allowed up to the original amount.

## Debugging Failed Payments

```bash
# List attempts for an intent
uqpay payment attempt list --page-size 50

# Get attempt details
uqpay payment attempt get <attempt_id>
```

Check `attempt_status` and `failure_code` for diagnosis.

## Payment Payout Flow

After payments settle, withdraw funds to your bank account:

```bash
# List payment balances
uqpay payment balance list

# Create payout
uqpay payment payout create \
  -d payout_currency=USD \
  -d payout_amount=500 \
  -d statement_descriptor=WEEKLY-PAYOUT
```
