# Changelog

All notable changes to `@uqpay/cli` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Add `uqpay operation-manifest` as a versioned machine-readable inventory of
  all 100 canonical API operations, including stable and OpenAPI operation IDs,
  HTTP contracts, risk classification, parameter/body metadata,
  connected-account and idempotency capabilities, and Developer Docs links.
  Its JSON Schema is available through `uqpay operation-manifest --schema`.

### Fixed

- Reuse one `x-idempotency-key` for every 401 token refresh and HTTP 429 retry
  belonging to the same logical request. Previously, commands relying on an
  automatically generated key could generate a new key per retry.
- Stop automatically replaying mutating requests after ambiguous connection or
  response-read failures, or HTTP 5xx responses. The CLI now returns structured
  `reconcile_required` output with the method, path, and original idempotency key
  so callers can inspect remote state before retrying.

## [2.0.0]

This major release replaces the previous Virtual Account Create output and
workflow with the Virtual Account application lifecycle contract. Existing CLI
automation must migrate before adopting this version.

### Added

- `virtual-account application list` and `virtual-account application retrieve`
  commands, with pagination, status/country/currency filters, and connected-account
  support.

### Changed

- Virtual Account Create help and examples now describe required `country`, one
  `currency`, optional `LOCAL`/`SWIFT`/omitted method, nickname, HTTP 200
  application output, and asynchronous lifecycle semantics.
- Virtual Account Create continues to require and forward `x-idempotency-key`
  through the existing `--idempotency-key` option; this is retained from the
  previous Create contract, not introduced by the application contract.
- Structured API errors retain the service's strict `type`, string `code`, and
  `message`, including HTTP 400 application not-found/cross-account responses.

### Documentation

- Documented required Gateway `account_id` and `direct_id` fields on successful
  Virtual Account Create/Retrieve data and List summaries. The CLI preserves
  both fields without validation or transformation; `direct_id` remains an
  ordinary string and is `"0"` for main accounts.
- Clarified that the CLI has no webhook parser.

### Breaking

- Virtual Account Create automation must pass `--country`, send one currency,
  and consume HTTP 200 application JSON instead of the previous HTTP 202
  `message` and `request_id` output.
- The CLI does not parse webhook deliveries. Webhook-driven automation outside
  the CLI must correlate by `application_id` and apply only higher
  `public_version` values.

### Migration

- Install with `npm install --global @uqpay/cli@2.0.0` and follow the
  [Virtual Account migration guide](https://developers.uqpay.com/global-account/v1.6/guide/migrate-to-virtual-account-applications).

## [1.2.0]

This bootstrap alignment release establishes the shared stable `1.2` capability
baseline used by all five UQPAY customer SDKs. It covers all 98 callable operations
in the current business API contract; Ramp remains outside the SDK product scope.

### Added

- Connect RFI list, retrieve, and answer commands.
- Issuing card limit, risk, PIN, ART, merchant-brand, and unsolicited-refund
  release commands.
- Payment terminal registration and PIN-key commands.
- DELETE requests with JSON bodies and consistent `x-client-id` transport headers.

### Changed

- Node.js 22 or newer is now required by the npm installer (previously Node.js 16).
  Downloaded UQPAY binaries do not require Node.js at runtime.
- The CLI now follows the stable `1.x` public command and flag compatibility policy.

### Migration

- Upgrade Node.js before installing through npm:
  `npm install --global @uqpay/cli@1.2.0`.

## [0.3.1]

### Fixed

- **`account create-sub` (INDIVIDUAL):** The command help did not document the
  `individual_info` fields the Account Center API now requires, so users following
  the help text built payloads the API rejects with HTTP 400. Updated the help
  text and runnable example to cover the required fields:
  - `individual_info.gender` (`MALE | FEMALE`) and `individual_info.annual_income`
    (string, USD) — required effective 2026-07-02.
  - `individual_info.employment_status`, `individual_info.industry`,
    `individual_info.job_title`, `individual_info.company_name` — required
    effective 2026-03-19 (now annotated with their effective date).
  - `individual_info.state` is now documented as unconditionally required
    (previously listed as GB/US-only), matching the spec's `IndividualInfo.required`
    list.
  - `individual_info.apartment_suite_or_floor` remains optional.

### Notes

- The CLI passes `-d key=value` pairs through verbatim via dot-notation, so no
  request struct or client validation changed — the fix is to the documented
  contract (help text + example) plus a regression test
  (`cmd/connect/account_test.go`) that guards it.
- Verified live against sandbox: a payload without `gender`/`annual_income` is
  rejected with `IndividualInfo.Gender`/`IndividualInfo.AnnualIncome` required
  errors; with both fields present those errors are gone.
