# Card Lifecycle

End-to-end flow for issuing a card, from product selection through cancellation.

## Step 1: List Products

Find a suitable card product for your use case.

```bash
uqpay issuing product list
```

Look for a product matching your needs:
- `card_currency` — the currency the card operates in (SGD, USD, EUR, etc.)
- `card_form` — `VIR` (virtual) or `PHY` (physical)
- `card_scheme` — `VISA` or `MASTERCARD`
- `mode_type` — `SINGLE` (one cardholder per card) or `SHARE` (multiple cards per cardholder)

Note the `card_product_id` for the next steps.

## Step 2: Create Cardholder

Create the KYC identity that the card will be linked to.

```bash
uqpay issuing cardholder create \
  -d first_name=Jane -d last_name=Smith \
  -d date_of_birth=1990-01-15 \
  -d email=jane@example.com \
  -d phone_number=91234567 \
  -d country_code=SG \
  -d gender=FEMALE -d nationality=SG \
  -d "delivery_address.city=Singapore" \
  -d delivery_address.country=SG \
  -d "delivery_address.line1=10 Anson Road" \
  -d delivery_address.postal_code=079903
```

Note the returned `cardholder_id`.

## Step 3: Create Card

Create a card linked to the product and cardholder.

```bash
uqpay issuing card create \
  -d card_product_id=<product_id> \
  -d cardholder_id=<cardholder_id> \
  -d card_currency=SGD
```

The response includes:
- `card_id` — the card identifier
- `card_order_id` — the async order tracking ID
- `order_status` — will be `PROCESSING`

## Step 4: Poll Card Order

Card creation is async. If `order_status` is `PROCESSING`, poll until it resolves.

```bash
uqpay issuing card get-order <card_order_id>
```

The `card_order_id` is returned in the card create response.

Possible statuses:
- `PROCESSING` — still being provisioned, poll again
- `SUCCESS` — card is ready
- `FAILED` — creation failed, check error details

## Step 5: Recharge Card

Load funds onto the card before it can be used for transactions.

```bash
uqpay issuing card recharge <card_id> -d amount=100 -d currency=SGD
```

## Step 6: Manage Lifecycle

### Freeze a Card (Reversible)

Temporarily suspend a card due to suspicious activity or cardholder request.

```bash
uqpay issuing card update-status <card_id> \
  -d card_status=FROZEN \
  -d reason="Suspicious activity"
```

### Unfreeze a Card

Reactivate a frozen card.

```bash
uqpay issuing card update-status <card_id> \
  -d card_status=ACTIVE \
  -d reason="Cleared"
```

### Withdraw Funds

Pull funds back from the card to the issuing balance.

```bash
uqpay issuing card withdraw <card_id> -d amount=50 -d currency=SGD
```

### Cancel a Card (Irreversible)

Permanently cancel a card. This cannot be undone.

```bash
uqpay issuing card update-status <card_id> \
  -d card_status=CANCELLED \
  -d reason="No longer needed"
```

## Physical Card Additional Steps

Physical cards require extra steps before they can be used.

### Activate a Physical Card

Physical cards arrive in `PENDING` status. Activate using the activation code printed on the card carrier.

```bash
uqpay issuing card activate \
  -d card_id=<card_id> \
  -d activation_code=<code> \
  -d pin=123456
```

### Assign Card to Cardholder

For cards that need to be assigned after creation (e.g., pre-printed inventory).

```bash
uqpay issuing card assign \
  -d cardholder_id=<cardholder_id> \
  -d card_number=5550710000000001 \
  -d card_currency=SGD \
  -d card_mode=SINGLE
```

### Set PIN

Set or reset the card PIN for ATM and POS transactions.

```bash
uqpay issuing card set-pin \
  -d card_id=<card_id> \
  -d pin=123456
```
