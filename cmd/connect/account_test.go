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
// IndividualInfo.required list), not GB/US-only.

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

func TestCreateSubHelpDocumentsRepresentativeDOBOnlyAsCompanyOptional(t *testing.T) {
	const companyStart = "Parameters (COMPANY entity):"
	companyIndex := strings.Index(accountCreateSubHelp, companyStart)
	if companyIndex < 0 {
		t.Fatalf("help text missing %q section", companyStart)
	}
	company := accountCreateSubHelp[companyIndex+len(companyStart):]

	const optionalStart = "Optional (when inherit=-1):"
	optionalIndex := strings.Index(company, optionalStart)
	if optionalIndex < 0 {
		t.Fatalf("COMPANY help missing %q section", optionalStart)
	}
	required := company[:optionalIndex]
	optional := company[optionalIndex:]
	const representativeDOB = "ownership_details.representatives[0].date_of_birth"
	if strings.Contains(required, representativeDOB) {
		t.Fatalf("%q must not be documented as required for COMPANY", representativeDOB)
	}
	if !strings.Contains(optional, representativeDOB) || !strings.Contains(optional, "YYYY-MM-DD") {
		t.Fatalf("%q must be documented as an optional YYYY-MM-DD field for COMPANY", representativeDOB)
	}

	individual := individualRequiredSection(t)
	if !strings.Contains(individual, "individual_info.date_of_birth") {
		t.Fatal("individual_info.date_of_birth must remain in the INDIVIDUAL Required block")
	}
}
