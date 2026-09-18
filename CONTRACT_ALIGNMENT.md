# API contract alignment

Contract reference: [OpenAPI revision 1feb1d2](https://github.com/uqpay/uqpay-docs/tree/1feb1d26d032c53b79ab44a7d48e88c9a91d397d/docs).

## PIN management

`issuing card set-pin` calls `/v1/issuing/cards/pin`. Despite historical reset naming in the SDKs, omitting `type` means initial `SET`. Explicitly pass `RESET` to reset without the current PIN, or `UPDATE` with `old_pin` to verify the current PIN before changing it. Both PIN values must contain six digits. `old_pin` is prohibited for `SET` and `RESET`. Validation and supported operations remain server-authoritative.

A response with `request_status=SUCCESS` and `order_status=PROCESSING` means the request was accepted. Use `card_order_id` with the existing card order retrieval operation to obtain `SUCCESS` or `FAILED`; a failed PIN order may include `failure_code`. PIN orders do not contain amount or card currency.

The legacy `/v1/issuing/cards/manage/pin` operation retains its existing behavior. Its `RESET` means changing with the old PIN; when migrating to the new endpoint use `UPDATE`, not `RESET`. Do not send legacy four-digit PINs to the new endpoint.

### PIN eligibility and migration deadline

The frozen reference contract schedules retirement of `/v1/issuing/cards/manage/pin` for **2026-10-17**. Migrate the legacy `RESET` operation to `UPDATE` on `/v1/issuing/cards/pin`. This date is the published contract schedule; these client tests do not verify Production rollout or actual endpoint retirement.

The card must be active and unexpired. Wait for any PIN operation already processing on that card to finish before submitting another. The new PIN must not repeat a single digit, occur within the cardholder's phone number, or share its last four digits with the card number. For `UPDATE`, it must differ from the current PIN. These checks and card-specific operation support are server-authoritative; a typed request does not establish eligibility.

## RFI answers

Send the full `rfi_id`, including its prefix. A `TEXT` answer needs non-empty `text`; an `ATTACHMENT` answer contains uploaded file IDs in `attachments`.

## Simulated deposits

Supply `account_id` explicitly, together with `amount`, `currency` and `sender_swift_code`. The recipient must be active and verified. The SDK does not infer the recipient from credentials or request headers. This operation is Sandbox-only.

## Transaction detail

`settlement_status` is available on transaction detail, and may be absent from list items. Values are `UNKNOWN`, `UNSETTLED`, `SETTLED` and `NOT_APPLICABLE`. `SETTLED` means clearing has been recorded, including partial clearing; it does not confirm full settlement. `UNKNOWN` does not mean unsettled.

## Beneficiary checks

Provide at least one non-empty `account_number` or `iban`. Both are accepted; `account_number` takes precedence. For LOCAL currencies without a default clearing system (including CNH), supply both `bank_country_code` and `clearing_system`. The default-route currencies are USD, GBP, EUR, SGD, CAD, AUD, HKD, MYR, IDR, PHP and INR. For SWIFT, use `clearing_system=SWIFT`. Validation is server-authoritative.

## Card art updates

Card updates accept `card_art_id` and `name_on_card`. Card art changes apply to virtual and physical cards; both `card_status` and `processing_status` must be `ACTIVE`. An accepted `PROCESSING` response is asynchronous: use `card_order_id` to check the final result.

## JSON output and pagination

JSON output preserves null fields and decimal strings returned by the server. Beneficiary-check and card-update regression fixtures cover request fields and asynchronous output. Issuing card list accepts page sizes from 1 to 100; its default remains 10.

## Boundary and response regression coverage

Offline fixtures exercise all three KYC entry points (create cardholder, update cardholder and create card with inline cardholder fields). They cover all six proof providers, reference lengths 9/10/64/65 and birth dates corresponding to ages 17/18/79/80 on 2026-09-17. These tests prove that the SDK preserves the submitted values; server acceptance, current configuration and environment rollout still require Sandbox verification. Inline cardholder fields include email, first/last name and country code.

Card list pagination preserves page sizes 1, 10 and 100. Banking list page sizes are 1–100. Currency conversions require a fresh quote per operation: `FUNDS_ARRIVED` is intermediate and `TRADE_SETTLED` is the successful terminal state.

Payment intent retrieval supports per-request `x-on-behalf-of` without requiring a caller-supplied idempotency key. Existing automatic GET header behavior remains language-specific; the contract no longer requires the header but does not prohibit it. POST idempotency and retry behavior remain unchanged.

Card creation orders use `CREATE_CARD`. Issuing transfer REST status uses uppercase `PENDING`/`FAILED`/`COMPLETED`; transfer webhook status is a distinct lowercase field. Wallet values remain open strings, including empty and unknown values. The SDK does not infer transaction context from them.

`payment intent get <id> --on-behalf-of <account-id>` sends the proxy header. Command-level tests exercise all three KYC entry points, page sizes 1/10/100 and unchanged JSON output for high-precision amounts and null objects.

## Acquiring GET headers

Offline route-specific checks cover D189–D196: balance list/detail, bank account list/detail, payout list/detail, settlements list and payment intent detail. Each call is exercised with and without `x-on-behalf-of`, while preserving the configured `x-client-id`. Callers need not supply an idempotency key for these GET requests; existing automatic GET headers remain supported. Existing POST idempotency and retry tests remain part of verification.

## Banking balance precision

Offline D122–D129 fixtures cover `available_balance`, `frozen_balance`, `margin_balance` and `prepaid_balance` through both list and detail reads. Each field receives zero with trailing decimal places, a positive amount, a negative amount, large positive/negative amounts, and a long decimal string. Distinct values across fields detect accidental swaps; comparisons retain strings and precision. Long decimal fixtures test client robustness, not server-supported currency precision. CLI verification covers JSON output.

## Acquiring nullable responses

Offline fixtures distinguish missing fields, explicit null, empty strings/objects and populated controls. REST checks cover payment attempts, refunds and payouts. Empty REST event times remain strings. Webhook fixtures exercise intent, attempt, refund, payout and chargeback alert families; empty timestamps or incomplete objects are robustness probes, not claims of server-valid payloads.

The CLI tests actual REST commands and JSON output. Webhook delivery parsing is not a CLI capability and is not applicable to this regression.

## Cardholder KYC event interpretation

For `cardholder.kyc.status_changed`, use `data.cardholder_status` to determine the outcome: `SUCCESS` is approved, `FAILED` is rejected, and `INCOMPLETE` requires additional information. The optional or empty `data.reason` explains rejection or a request for more information; its presence alone does not determine the outcome.
