# PIN, RFI, deposit simulation and settlement contracts

Contract reference: [OpenAPI revision 8267056](https://github.com/uqpay/uqpay-docs/tree/8267056f7fefc1183b069e0387e5ebaecbd2ad27/docs).

## PIN management

`issuing card set-pin` calls `/v1/issuing/cards/pin`. Despite historical reset naming in the SDKs, omitting `type` means initial `SET`. Explicitly pass `RESET` to reset without the current PIN, or `UPDATE` with `old_pin` to verify the current PIN before changing it. Both PIN values must contain six digits. `old_pin` is prohibited for `SET` and `RESET`. Validation and supported operations remain server-authoritative.

A response with `request_status=SUCCESS` and `order_status=PROCESSING` means the request was accepted. Use `card_order_id` with the existing card order retrieval operation to obtain `SUCCESS` or `FAILED`; a failed PIN order may include `failure_code`. PIN orders do not contain amount or card currency.

The legacy `/v1/issuing/cards/manage/pin` operation retains its existing behavior. Its `RESET` means changing with the old PIN; when migrating to the new endpoint use `UPDATE`, not `RESET`. Do not send legacy four-digit PINs to the new endpoint.

## RFI answers

Send the full `rfi_id`, including its prefix. A `TEXT` answer needs non-empty `text`; an `ATTACHMENT` answer contains uploaded file IDs in `attachments`.

## Simulated deposits

Supply `account_id` explicitly, together with `amount`, `currency` and `sender_swift_code`. The recipient must be active and verified. The SDK does not infer the recipient from credentials or request headers. This operation is Sandbox-only.

## Transaction detail

`settlement_status` is available on transaction detail, and may be absent from list items. Values are `UNKNOWN`, `UNSETTLED`, `SETTLED` and `NOT_APPLICABLE`. `SETTLED` means clearing has been recorded, including partial clearing; it does not confirm full settlement. `UNKNOWN` does not mean unsettled.
