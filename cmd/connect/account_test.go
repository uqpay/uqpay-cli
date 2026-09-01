package connect

import (
	"strings"
	"testing"
)

// The CLI passes account create-sub fields through verbatim via dot-notation
// (-d key=value), so its only "contract" for what an INDIVIDUAL payload must
// contain is the help text. These tests guard that contract against the
// Account Center breaking changes to Create SubAccount individual_info:
//   - 2026-03-19: employment_status, industry, job_title, company_name
//   - 2026-07-02: gender, annual_income
// and that state is documented as unconditionally required (per the spec's
// IndividualInfo.required list), not GB/US-only. COMPANY checks below cover
// the v3 non-inherited onboarding contract.

// individualRequiredSection returns the "Required:" block of the INDIVIDUAL
// entity section in the create-sub help, i.e. everything between the
// "Parameters (INDIVIDUAL entity):" header and the next "Required (" /
// "Optional:" / "Parameters (" boundary.
func individualRequiredSection(t *testing.T) string {
	t.Helper()
	const start = "Parameters (INDIVIDUAL entity):"
	i := strings.Index(accountCreateSubHelp, start)
	if i < 0 {
		t.Fatalf("help text missing %q section", start)
	}
	rest := accountCreateSubHelp[i+len(start):]
	end := len(rest)
	for _, marker := range []string{"\n  Required (", "\n  Optional:", "\nParameters ("} {
		if j := strings.Index(rest, marker); j >= 0 && j < end {
			end = j
		}
	}
	return rest[:end]
}

func TestCreateSubHelpDocumentsRequiredIndividualFields(t *testing.T) {
	required := individualRequiredSection(t)

	// Field path that must appear in the unconditional Required block, with the
	// effective date that made it required (for the failure message).
	cases := []struct {
		field string
		since string
	}{
		{"individual_info.employment_status", "2026-03-19"},
		{"individual_info.industry", "2026-03-19"},
		{"individual_info.job_title", "2026-03-19"},
		{"individual_info.company_name", "2026-03-19"},
		{"individual_info.gender", "2026-07-02"},
		{"individual_info.annual_income", "2026-07-02"},
		{"individual_info.state", "spec required list"},
	}
	for _, tc := range cases {
		if !strings.Contains(required, tc.field) {
			t.Errorf("create-sub help INDIVIDUAL Required block is missing %q (required since %s)", tc.field, tc.since)
		}
	}
}

func TestCreateSubHelpGenderEnumDocumented(t *testing.T) {
	required := individualRequiredSection(t)
	if !strings.Contains(required, "individual_info.gender") {
		t.Fatal("gender field not documented")
	}
	// gender is an enum: only MALE or FEMALE.
	for _, v := range []string{"MALE", "FEMALE"} {
		if !strings.Contains(required, v) {
			t.Errorf("gender enum value %q not documented in INDIVIDUAL Required block", v)
		}
	}
}

func TestCreateSubHelpExampleIncludesNewRequiredFields(t *testing.T) {
	// The runnable example must produce a payload the API will accept, so it has
	// to set every newly required field.
	for _, field := range []string{
		"individual_info.gender",
		"individual_info.annual_income",
		"individual_info.state",
	} {
		if !strings.Contains(accountCreateSubHelp, field+"=") {
			t.Errorf("create-sub help example does not set %q", field)
		}
	}
}

func companyNonInheritedRequiredSection(t *testing.T) string {
	t.Helper()
	const companyStart = "Parameters (COMPANY entity):"
	companyIndex := strings.Index(accountCreateSubHelp, companyStart)
	if companyIndex < 0 {
		t.Fatalf("help text missing %q section", companyStart)
	}
	company := accountCreateSubHelp[companyIndex+len(companyStart):]

	const requiredStart = "Required (when inherit=-1):"
	requiredIndex := strings.Index(company, requiredStart)
	if requiredIndex < 0 {
		t.Fatalf("COMPANY help missing %q section", requiredStart)
	}
	rest := company[requiredIndex+len(requiredStart):]
	end := len(rest)
	for _, marker := range []string{"\n  Optional", "\nExamples:"} {
		if i := strings.Index(rest, marker); i >= 0 && i < end {
			end = i
		}
	}
	return rest[:end]
}

func TestCreateSubHelpDocumentsCompanyV3RequiredFields(t *testing.T) {
	required := companyNonInheritedRequiredSection(t)
	expected := []string{
		"company_info.legal_business_name",
		"company_info.legal_business_name_english",
		"company_info.country_of_incorporation",
		"company_info.company_type",
		"company_info.phone_number",
		"company_info.email_address",
		"company_info.company_registration_number",
		"company_info.incorparate_date",
		"company_info.certification_of_incorporation[]",
		"company_address.street_address",
		"company_address.city",
		"company_address.state",
		"company_address.postal_code",
		"ownership_details.representatives[0].legal_first_name_english",
		"ownership_details.representatives[0].legal_last_name_english",
		"ownership_details.representatives[0].email_address",
		"ownership_details.representatives[0].is_applicant",
		"ownership_details.representatives[0].job_title",
		"ownership_details.representatives[0].ownership_percentage",
		"ownership_details.representatives[0].nationality",
		"ownership_details.representatives[0].date_of_birth",
		"ownership_details.representatives[0].country_or_territory",
		"ownership_details.representatives[0].street_address",
		"ownership_details.representatives[0].city",
		"ownership_details.representatives[0].state",
		"ownership_details.representatives[0].postal_code",
		"ownership_details.representatives[0].identification_type",
		"ownership_details.representatives[0].identification_value",
		"ownership_details.representatives[0].identity_docs[]",
		"ownership_details.shareholder_docs[]",
		"business_details.country_or_territory",
		"business_details.street_address",
		"business_details.city",
		"business_details.state",
		"business_details.postal_code",
		"business_details.industry",
		"business_details.turnover_monthly",
		"business_details.number_of_employee",
		"business_details.account_purpose[]",
		"business_details.banking_currencies[]",
		"business_details.banking_countries[]",
		"business_details.articles_of_association[]",
	}

	actual := make(map[string]bool)
	for _, line := range strings.Split(required, "\n") {
		columns := strings.Fields(line)
		if len(columns) > 0 {
			actual[columns[0]] = true
		}
	}
	expectedSet := make(map[string]bool, len(expected))
	for _, field := range expected {
		expectedSet[field] = true
		if !actual[field] {
			t.Errorf("create-sub help COMPANY required block is missing %q", field)
		}
	}
	for field := range actual {
		if !expectedSet[field] {
			t.Errorf("create-sub help COMPANY required block contains unexpected field %q", field)
		}
	}
}

func TestCreateSubHelpDocumentsCompanyV3PurposeEnum(t *testing.T) {
	required := companyNonInheritedRequiredSection(t)
	for _, purpose := range []string{
		"PAYMENT_COLLECTION",
		"PAYOUT_DISBURSEMENT",
		"MULTI_CURRENCY_BANKING",
		"CARD_ISSUING",
		"CRYPTO_RAMP",
		"GLOBAL_TRANSFER",
		"TREASURY_FX",
		"OTHERS",
	} {
		if !strings.Contains(required, purpose) {
			t.Errorf("create-sub help COMPANY purpose enum is missing %q", purpose)
		}
	}
	for _, removed := range []string{"BUSINESS_PAYMENT", "BILL_PAYMENT", "INVESTMENT"} {
		if strings.Contains(required, removed) {
			t.Errorf("create-sub help COMPANY purpose enum still contains removed value %q", removed)
		}
	}
}

func TestParseCreateSubDataPreservesOwnershipPercentageString(t *testing.T) {
	body, err := parseCreateSubData([]string{
		"entity_type=COMPANY",
		"inherit=-1",
		"ownership_details.representatives[0].ownership_percentage=0",
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := body["inherit"]; got != int64(-1) {
		t.Fatalf("inherit = %T(%v), want int64(-1)", got, got)
	}
	representatives := body["ownership_details"].(map[string]any)["representatives"].([]any)
	got := representatives[0].(map[string]any)["ownership_percentage"]
	if got != "0" {
		t.Fatalf("ownership_percentage = %T(%v), want string %q", got, got, "0")
	}
}
