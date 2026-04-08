# Beneficiary Create Reference

Beneficiaries are registered recipients for payouts. The required fields depend on `entity_type` and `payment_method`.

## INDIVIDUAL Entity Type

### Required Fields

| Field | Description |
|-------|-------------|
| `entity_type` | `INDIVIDUAL` |
| `first_name` | Recipient's first name |
| `last_name` | Recipient's last name |
| `nickname` | Display name for the beneficiary |
| `payment_method` | `SWIFT` or `LOCAL` |

### Address (required, dot notation)

| Field | Description |
|-------|-------------|
| `address.country` | ISO 3166-1 alpha-2 country code (e.g., `US`, `SG`, `GB`) |
| `address.city` | City name |
| `address.state` | State or province |
| `address.street_address` | Street address |
| `address.postal_code` | Postal/ZIP code |

### Bank Details (required, dot notation)

| Field | Description |
|-------|-------------|
| `bank_details.bank_name` | Name of the recipient's bank |
| `bank_details.bank_address` | Address of the bank |
| `bank_details.account_holder` | Name on the bank account |
| `bank_details.account_currency_code` | Currency of the bank account (e.g., `USD`, `SGD`) |
| `bank_details.bank_country_code` | ISO 3166-1 alpha-2 country code of the bank |
| `bank_details.clearing_system` | `SWIFT` or local clearing system name |
| `bank_details.swift_code` | SWIFT/BIC code (required when `payment_method=SWIFT`) |
| `bank_details.account_number` | Bank account number |

### Additional Fields for LOCAL Payment Method

| Field | Description |
|-------|-------------|
| `bank_details.routing_code_type1` | Type of routing code (e.g., `aba` for US, `sort_code` for UK) |
| `bank_details.routing_code_value1` | The routing code value |

## COMPANY Entity Type

### Required Fields

| Field | Description |
|-------|-------------|
| `entity_type` | `COMPANY` |
| `company_name` | Registered company name |
| `nickname` | Display name for the beneficiary |
| `payment_method` | `SWIFT` or `LOCAL` |

Address and bank_details fields are the same as INDIVIDUAL above.

**Note:** COMPANY does not use `first_name` or `last_name`. Use `company_name` instead.

## Payment Method Differences

### SWIFT

- International wire transfers via the SWIFT network
- Requires `bank_details.swift_code` (8 or 11 character BIC)
- `bank_details.clearing_system` should be `SWIFT`
- Works for most international corridors

### LOCAL

- Domestic payment rails (ACH, Faster Payments, SEPA, etc.)
- Requires `bank_details.routing_code_type1` and `bank_details.routing_code_value1`
- `bank_details.clearing_system` should match the local system
- Common routing code types by country:

| Country | `routing_code_type1` | Example |
|---------|---------------------|---------|
| US | `aba` | `021000021` |
| UK | `sort_code` | `40-47-84` |
| AU | `bsb_code` | `062-000` |
| CA | `institution_no` / `transit_no` | `004` / `01234` |

## Example: INDIVIDUAL with SWIFT

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
  -d bank_details.swift_code=BOFAUS3N \
  -d bank_details.account_number=12345678999
```

## Example: COMPANY with LOCAL (US ACH)

```bash
uqpay banking beneficiary create \
  -d entity_type=COMPANY \
  -d company_name="Acme Corp" \
  -d nickname="Acme Corp" \
  -d payment_method=LOCAL \
  -d address.country=US -d "address.city=San Francisco" -d address.state=CA \
  -d "address.street_address=456 Market St" -d address.postal_code=94105 \
  -d "bank_details.bank_name=Chase Bank" \
  -d "bank_details.bank_address=456 Market St, San Francisco" \
  -d "bank_details.account_holder=Acme Corp" \
  -d bank_details.account_currency_code=USD \
  -d bank_details.bank_country_code=US \
  -d bank_details.clearing_system=ACH \
  -d bank_details.routing_code_type1=aba \
  -d bank_details.routing_code_value1=021000021 \
  -d bank_details.account_number=9876543210
```
