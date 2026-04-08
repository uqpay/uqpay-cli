# Payout Flow

A complete payout involves up to 4 steps. Same-currency payouts skip steps 2-3.

## Step 1: Create Beneficiary

Register the recipient with their bank details. This only needs to be done once per recipient — reuse the `beneficiary_id` for future payouts.

```bash
uqpay banking beneficiary create \
  -d entity_type=INDIVIDUAL \
  -d first_name=John -d last_name=Doe \
  -d nickname="John Doe" \
  -d payment_method=SWIFT \
  -d address.country=US -d "address.city=New York" -d address.state=NY \
  -d "address.street_address=123 Main St" -d address.postal_code=10001 \
  -d "bank_details.bank_name=Bank of America" \
  -d "bank_details.bank_address=123 Main St, New York" \
  -d bank_details.account_holder="John Doe" \
  -d bank_details.account_currency_code=USD \
  -d bank_details.bank_country_code=US \
  -d bank_details.clearing_system=SWIFT \
  -d bank_details.swift_code=WELGBE22 \
  -d bank_details.account_number=12345678999
```

Save the returned `id` — this is the `beneficiary_id` needed in step 4.

## Step 2: Get Exchange Rate Quote (Cross-Currency Only)

If the payout currency differs from your balance currency, get a binding FX quote first.

```bash
uqpay banking conversion quote \
  -d sell_currency=SGD \
  -d sell_amount=1000 \
  -d buy_currency=USD \
  -d conversion_date=2026-04-08
```

Save the returned `quote_id`. **The quote is valid for 75 seconds** — proceed to step 3 immediately.

## Step 3: Create Conversion (Cross-Currency Only)

Execute the conversion using the quote from step 2.

```bash
uqpay banking conversion create \
  -d quote_id=<quote_id from step 2> \
  -d sell_currency=SGD \
  -d sell_amount=1000 \
  -d buy_currency=USD \
  -d conversion_date=2026-04-08
```

This moves funds from your SGD balance to your USD balance at the quoted rate.

## Step 4: Create Payout

Send the payment to the beneficiary.

```bash
uqpay banking payout create \
  -d currency=USD \
  -d amount=500 \
  -d purpose_code=GOODS_PURCHASED \
  -d "payout_reference=Invoice #001" \
  -d fee_paid_by=OURS \
  -d payout_date=2026-04-10 \
  -d beneficiary_id=<beneficiary_id from step 1>
```

## Notes

- **Same-currency payouts** skip steps 2 and 3. If you already hold sufficient balance in the payout currency, go directly from step 1 to step 4.
- **Quote validity** — conversion quotes expire after 75 seconds. If the quote expires, request a new one before creating the conversion.
- **Payout date** — `payout_date` must be a valid business day in the relevant payment corridor. Weekends and bank holidays will be rejected.
- **Purpose codes** — the required `purpose_code` depends on the payment corridor. Run `uqpay banking payout create -h` to see accepted values.
- **Fee options** — `fee_paid_by` can be `OURS` (sender pays all fees) or `SHARED` (fees split between sender and recipient).
- **Tracking** — after creating a payout, use `uqpay banking payout get <id>` to monitor its status.
