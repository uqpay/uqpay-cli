# Card Creation Reference

Detailed parameter reference for `uqpay issuing card create`.

## Required Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `card_product_id` | string | Product ID from `product list` |
| `cardholder_id` | string | Cardholder ID from `cardholder create` or `cardholder list` |
| `card_currency` | string | Card currency (must match product's supported currency) |

## Optional Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| `card_name` | string | Custom name/label for the card |
| `card_limit` | number | Required for SHARE mode products; max number of cards per cardholder |
| `metadata` | object | Arbitrary key-value pairs for your reference |

## KYC Supplementation

When a product requires more KYC data than the cardholder record provides, supply the missing fields via `cardholder_required_fields`. This supplements the cardholder's KYC without modifying the cardholder record itself.

### Available Fields

```bash
# Personal info
-d cardholder_required_fields.gender=MALE
-d cardholder_required_fields.nationality=SG

# Identity document
-d cardholder_required_fields.identity.type=PASSPORT
-d cardholder_required_fields.identity.number=E1234567X
-d cardholder_required_fields.identity.front_file=@./id_front.jpg
-d cardholder_required_fields.identity.back_file=@./id_back.jpg

# Residential address
-d "cardholder_required_fields.residential_address.line1=10 Anson Road"
-d cardholder_required_fields.residential_address.line2=
-d cardholder_required_fields.residential_address.city=Singapore
-d cardholder_required_fields.residential_address.state=
-d cardholder_required_fields.residential_address.country=SG
-d cardholder_required_fields.residential_address.postal_code=079903
```

Identity document types: `PASSPORT`, `ID_CARD`, `DRIVER_LICENSE`

Note: `@./id_front.jpg` uses the `@filepath` convention to base64-encode the file contents automatically. See [uqpay-shared](../../uqpay-shared/SKILL.md) for details.

## Spending Controls

Set per-card transaction limits using the `spending_controls` array.

```bash
# Single limit: max $500 per transaction
-d spending_controls[0].amount=500
-d spending_controls[0].interval=PER_TRANSACTION

# Multiple limits
-d spending_controls[0].amount=500
-d spending_controls[0].interval=PER_TRANSACTION
-d spending_controls[1].amount=2000
-d spending_controls[1].interval=DAILY
-d spending_controls[2].amount=10000
-d spending_controls[2].interval=MONTHLY
```

Available intervals: `PER_TRANSACTION`, `DAILY`, `WEEKLY`, `MONTHLY`, `YEARLY`, `ALL_TIME`

## Risk Controls

Configure 3DS and MCC (Merchant Category Code) restrictions.

```bash
# Allow or deny 3DS transactions
-d risk_controls.allow_3ds_transactions=Y

# MCC whitelist (only allow these merchant categories)
-d risk_controls.allowed_mcc[0]=5411
-d risk_controls.allowed_mcc[1]=5812
-d risk_controls.allowed_mcc[2]=5541

# MCC blacklist (block these merchant categories)
-d risk_controls.blocked_mcc[0]=7995
-d risk_controls.blocked_mcc[1]=7801
```

`allow_3ds_transactions`: `Y` (allow) or `N` (deny)

`allowed_mcc` and `blocked_mcc` are **mutually exclusive** — use one or the other, not both. If neither is set, all MCCs are allowed.

## Examples

### Simple Virtual Card

Minimal creation with just the required fields.

```bash
uqpay issuing card create \
  -d card_product_id=prod_abc123 \
  -d cardholder_id=ch_xyz789 \
  -d card_currency=SGD
```

### Card with KYC Supplementation

When the product requires identity verification beyond what the cardholder has on file.

```bash
uqpay issuing card create \
  -d card_product_id=prod_abc123 \
  -d cardholder_id=ch_xyz789 \
  -d card_currency=USD \
  -d cardholder_required_fields.gender=FEMALE \
  -d cardholder_required_fields.nationality=SG \
  -d cardholder_required_fields.identity.type=PASSPORT \
  -d cardholder_required_fields.identity.number=E1234567X \
  -d cardholder_required_fields.identity.front_file=@./passport_front.jpg \
  -d cardholder_required_fields.identity.back_file=@./passport_back.jpg \
  -d "cardholder_required_fields.residential_address.line1=10 Anson Road" \
  -d cardholder_required_fields.residential_address.city=Singapore \
  -d cardholder_required_fields.residential_address.country=SG \
  -d cardholder_required_fields.residential_address.postal_code=079903
```

### Card with Spending Controls

Create a card with transaction limits and MCC restrictions.

```bash
uqpay issuing card create \
  -d card_product_id=prod_abc123 \
  -d cardholder_id=ch_xyz789 \
  -d card_currency=SGD \
  -d card_name="Expense Card" \
  -d spending_controls[0].amount=500 \
  -d spending_controls[0].interval=PER_TRANSACTION \
  -d spending_controls[1].amount=5000 \
  -d spending_controls[1].interval=MONTHLY \
  -d risk_controls.allow_3ds_transactions=Y \
  -d risk_controls.allowed_mcc[0]=5411 \
  -d risk_controls.allowed_mcc[1]=5812 \
  -d risk_controls.allowed_mcc[2]=5541
```
