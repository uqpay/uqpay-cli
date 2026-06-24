# Changelog

All notable changes to `@uqpay/cli` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
