# Account Onboarding

> **Prerequisite:** read [`../../uqpay-shared/SKILL.md`](../../uqpay-shared/SKILL.md) first.

Complete onboarding flow for creating connected accounts.

## COMPANY Account (via `account create`)

```bash
uqpay account create \
  -d entity_type=COMPANY \
  -d name="Acme Corp" \
  -d country=SG \
  -d contact_details.email=admin@acme.com \
  -d contact_details.phone=+6591234567 \
  -d business_details.legal_entity_name_english="Acme Corp Pte Ltd" \
  -d business_details.incorporation_date=2020-01-01 \
  -d business_details.registration_number=T99CC9999Z \
  -d business_details.business_structure=LIMITED_COMPANY \
  -d "business_details.product_description=Software services" \
  -d business_details.merchant_category_code=7372 \
  -d business_details.estimated_worker_count=BS001 \
  -d business_details.monthly_estimated_revenue.amount=TM001 \
  -d business_details.monthly_estimated_revenue.currency=SGD \
  -d "business_details.account_purpose[]=USE_API" \
  -d "registration_address.line1=1 Raffles Place" \
  -d registration_address.city=Singapore \
  -d registration_address.state=SG \
  -d registration_address.postal_code=048616 \
  -d "business_address[0].line1=1 Raffles Place" \
  -d business_address[0].city=Singapore \
  -d business_address[0].country=SG \
  -d business_address[0].state=SG \
  -d business_address[0].postal_code=048616 \
  -d representatives[0].roles=DIRECTOR \
  -d representatives[0].first_name=John \
  -d representatives[0].last_name=Doe \
  -d representatives[0].nationality=SG \
  -d representatives[0].date_of_birth=1990-01-15 \
  -d representatives[0].identification.type=PASSPORT \
  -d representatives[0].identification.id_number=E1234567 \
  -d "representatives[0].identification.front=@+./id_front.png" \
  -d "representatives[0].address.line1=10 Anson Road" \
  -d representatives[0].address.city=Singapore \
  -d representatives[0].address.country=SG \
  -d representatives[0].address.postal_code=079903
```

## INDIVIDUAL Sub-Account (via `account create-sub`)

```bash
uqpay account create-sub \
  -d entity_type=INDIVIDUAL \
  -d nickname="John Doe" \
  -d individual_info.first_name_english=John \
  -d individual_info.last_name_english=Doe \
  -d individual_info.nationality=GB \
  -d individual_info.phone_number=+447911123456 \
  -d individual_info.email_address=john@example.com \
  -d individual_info.date_of_birth=1990-01-15 \
  -d individual_info.country_or_territory=GB \
  -d "individual_info.street_address=123 Baker Street" \
  -d individual_info.city=London \
  -d individual_info.state=England \
  -d individual_info.postal_code=W1U6RS \
  -d individual_info.employment_status=Employed \
  -d "individual_info.industry=Information Technology/IT" \
  -d "individual_info.job_title=Business and administration professionals" \
  -d "individual_info.company_name=Acme Corp" \
  -d identity_verification.identification_type=PASSPORT \
  -d identity_verification.identification_value=P12345678 \
  -d "identity_verification.identity_docs[0]=@+./passport.png" \
  -d "identity_verification.face_docs[0]=@+./selfie.png" \
  -d "expected_activity.account_purpose[0]=PURCHASE" \
  -d "expected_activity.banking_countries[0]=GB" \
  -d "expected_activity.banking_currencies[0]=GBP" \
  -d expected_activity.internationally=1 \
  -d expected_activity.turnover_monthly=TM002 \
  -d expected_activity.turnover_monthly_currency=GBP \
  -d "proof_documents.proof_of_address[0]=@+./utility_bill.png" \
  -d tos_acceptance.ip=192.168.1.1 \
  -d tos_acceptance.date=2026-04-08T00:00:00Z \
  -d tos_acceptance.user_agent=uqpay-cli \
  -d tos_acceptance.tos_agreement=1
```

## COMPANY Sub-Account (via `account create-sub`)

```bash
uqpay account create-sub \
  -d entity_type=COMPANY \
  -d nickname="SDK Test Sub" \
  -d inherit=-1 \
  -d company_info.legal_business_name="Test Company Ltd" \
  -d company_info.legal_business_name_english="Test Company Ltd" \
  -d company_info.country_of_incorporation=SG \
  -d company_info.company_type=LIMITED_COMPANY \
  -d company_info.phone_number=+6591234567 \
  -d company_info.email_address=company@example.com \
  -d company_info.company_registration_number=T99CS9999Z \
  -d company_info.incorparate_date=2020-01-01 \
  -d "company_info.certification_of_incorporation[0]=@+./cert.png" \
  -d "company_address.street_address=1 Raffles Place" \
  -d company_address.city=Singapore \
  -d company_address.state=SG \
  -d company_address.postal_code=048616 \
  -d ownership_details.representatives[0].legal_first_name_english=Jane \
  -d ownership_details.representatives[0].legal_last_name_english=Smith \
  -d ownership_details.representatives[0].email_address=jane@example.com \
  -d ownership_details.representatives[0].is_applicant=1 \
  -d ownership_details.representatives[0].job_title=DIRECTOR \
  -d ownership_details.representatives[0].nationality=SG \
  -d ownership_details.representatives[0].phone_number=+6591234567 \
  -d ownership_details.representatives[0].date_of_birth=1985-03-20 \
  -d ownership_details.representatives[0].country_or_territory=SG \
  -d "ownership_details.representatives[0].street_address=1 Raffles Place" \
  -d ownership_details.representatives[0].city=Singapore \
  -d ownership_details.representatives[0].state=SG \
  -d ownership_details.representatives[0].postal_code=048616 \
  -d ownership_details.representatives[0].identification_type=PASSPORT \
  -d ownership_details.representatives[0].identification_value=E1234567 \
  -d "ownership_details.representatives[0].identity_docs[0]=@+./id.png" \
  -d "ownership_details.shareholder_docs[0]=@+./shareholder.png" \
  -d business_details.country_or_territory=SG \
  -d "business_details.street_address=1 Raffles Place" \
  -d business_details.city=Singapore \
  -d business_details.state=SG \
  -d business_details.postal_code=048616 \
  -d business_details.industry=7372 \
  -d business_details.turnover_monthly=TM001 \
  -d business_details.number_of_employee=BS001 \
  -d tos_acceptance.ip=192.168.1.1 \
  -d tos_acceptance.date=2026-04-08T00:00:00Z \
  -d tos_acceptance.user_agent=uqpay-cli \
  -d tos_acceptance.tos_agreement=1
```

## After Account Creation

Account status will be `PROCESSING`. Check status:

```bash
uqpay account get <account_id>
```

Once `status` is `ACTIVE`, use `--on-behalf-of` for all domain operations:

```bash
uqpay --on-behalf-of <account_id> banking balance list
uqpay --on-behalf-of <account_id> issuing card list
```

## Check Required Additional Documents

Query what additional documents are needed for a country/business type:

```bash
uqpay account additional-documents --country SG --business-code BANKING
```

## Upload Files

Use `uqpay file upload` to upload supporting documents:

```bash
uqpay file upload ./additional_doc.pdf
```
