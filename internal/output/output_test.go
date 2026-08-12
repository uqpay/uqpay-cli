package output_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/uqpay/uqpay-cli/internal/output"
)

func TestJSONPassthrough(t *testing.T) {
	data := []byte(`{"id":"card_123","status":"ACTIVE"}`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "json"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, `"card_123"`) {
		t.Errorf("JSON output missing id: %s", got)
	}
	// Must be plain text — no ANSI escape codes
	if strings.Contains(got, "\x1b[") {
		t.Error("JSON output must not contain ANSI escape codes")
	}
}

func TestYAMLOutput(t *testing.T) {
	data := []byte(`{"id":"card_123","status":"ACTIVE"}`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "yaml"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "card_123") {
		t.Errorf("YAML output missing id: %s", got)
	}
	if !strings.Contains(got, "id:") {
		t.Errorf("YAML output missing key: %s", got)
	}
}

func TestTableList(t *testing.T) {
	data := []byte(`{"data":[{"id":"card_1","status":"ACTIVE"},{"id":"card_2","status":"FROZEN"}]}`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "card_1") || !strings.Contains(got, "card_2") {
		t.Errorf("table missing rows: %s", got)
	}
}

func TestTableSingleRecord(t *testing.T) {
	data := []byte(`{"id":"card_123","status":"ACTIVE","currency":"USD"}`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "card_123") {
		t.Errorf("table missing value: %s", got)
	}
}

func TestDirectArray(t *testing.T) {
	data := []byte(`[{"id":"card_1"},{"id":"card_2"}]`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "card_1") {
		t.Errorf("table missing row from array: %s", got)
	}
}

func TestEmptyList(t *testing.T) {
	data := []byte(`{"data":[]}`)
	var buf bytes.Buffer
	if err := output.Print(&buf, data, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "No results") {
		t.Errorf("expected 'No results' for empty list, got: %s", got)
	}
}

func TestPaginationHint(t *testing.T) {
	// 10 rows → should show pagination hint
	rows := make([]any, 10)
	for i := range rows {
		rows[i] = map[string]any{"id": fmt.Sprintf("item_%d", i)}
	}
	b, _ := json.Marshal(map[string]any{"data": rows})

	var buf bytes.Buffer
	if err := output.Print(&buf, b, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if !strings.Contains(got, "page-num") {
		t.Errorf("expected pagination hint with --page-num, got:\n%s", got)
	}
}

func TestNoPaginationHintForSmallResult(t *testing.T) {
	// 3 rows → no hint
	rows := []any{
		map[string]any{"id": "a"},
		map[string]any{"id": "b"},
		map[string]any{"id": "c"},
	}
	b, _ := json.Marshal(map[string]any{"data": rows})

	var buf bytes.Buffer
	if err := output.Print(&buf, b, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "page-num") {
		t.Errorf("expected no pagination hint for 3 results, got:\n%s", got)
	}
}

func TestNoPaginationHintForDirectArray(t *testing.T) {
	// Direct array (not {"data": [...]}) → no hint even with 10 items
	rows := make([]any, 10)
	for i := range rows {
		rows[i] = map[string]any{"id": fmt.Sprintf("item_%d", i)}
	}
	b, _ := json.Marshal(rows)

	var buf bytes.Buffer
	if err := output.Print(&buf, b, "table"); err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if strings.Contains(got, "page-num") {
		t.Errorf("expected no pagination hint for direct array, got:\n%s", got)
	}
}

func printAndDecodeJSON(t *testing.T, fixture string) map[string]any {
	t.Helper()
	var stdout bytes.Buffer
	if err := output.Print(&stdout, []byte(fixture), "json"); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &got); err != nil {
		t.Fatalf("stdout is not structured JSON: %v\n%s", err, stdout.String())
	}
	return got
}

func TestVirtualAccountApplicationCreateOutputPreservesSkippedResult(t *testing.T) {
	got := printAndDecodeJSON(t, `{
		"data": {
			"application_id": "app-create",
			"public_version": 1,
			"country": "SG",
			"currency": "USD",
			"status": "SUBMITTED",
			"results": [
				{
					"payment_method": "LOCAL",
					"status": "SKIPPED",
					"virtual_accounts": [],
					"error": {
						"code": "VA_METHOD_NOT_SUPPORTED",
						"message": "LOCAL is not supported for the requested country and currency"
					}
				},
				{
					"payment_method": "SWIFT",
					"status": "SUBMITTED",
					"virtual_accounts": [],
					"error": null
				}
			]
		}
	}`)
	data := got["data"].(map[string]any)
	results := data["results"].([]any)
	skipped := results[0].(map[string]any)
	if skipped["status"] != "SKIPPED" {
		t.Fatalf("SKIPPED status lost: %#v", skipped)
	}
	errBody := skipped["error"].(map[string]any)
	if errBody["code"] != "VA_METHOD_NOT_SUPPORTED" {
		t.Fatalf("SKIPPED error lost: %#v", errBody)
	}
	if submitted := results[1].(map[string]any); submitted["error"] != nil {
		t.Fatalf("null result error changed: %#v", submitted)
	}
}

func TestVirtualAccountApplicationRetrieveOutputPreservesFailureAndBankDetails(t *testing.T) {
	got := printAndDecodeJSON(t, `{
		"data": {
			"application_id": "app-detail",
			"public_version": 4,
			"country": "BH",
			"currency": "GBP",
			"status": "PARTIALLY_COMPLETED",
			"results": [
				{
					"payment_method": "LOCAL",
					"status": "FAILED",
					"virtual_accounts": [],
					"error": {
						"code": "VA_PROVISIONING_FAILED",
						"message": "Virtual account provisioning failed"
					}
				},
				{
					"payment_method": "SWIFT",
					"status": "CLOSED",
					"virtual_accounts": [{
						"account_bank_id": "bank-1",
						"account_holder": "Example Merchant Ltd.",
						"account_number": "001234",
						"country_code": "BH",
						"currency": "GBP",
						"bank_name": "Example Bank",
						"bank_address": "Manama, Bahrain",
						"clearing_system": {"type": "bic_swift", "value": "BANKBHBM"},
						"status": "CLOSED",
						"close_reason": ""
					}],
					"error": null
				}
			]
		}
	}`)
	results := got["data"].(map[string]any)["results"].([]any)
	failedError := results[0].(map[string]any)["error"].(map[string]any)
	if failedError["code"] != "VA_PROVISIONING_FAILED" ||
		failedError["message"] != "Virtual account provisioning failed" {
		t.Fatalf("async failure fields lost: %#v", failedError)
	}
	bank := results[1].(map[string]any)["virtual_accounts"].([]any)[0].(map[string]any)
	closeReason, exists := bank["close_reason"]
	if !exists || closeReason != "" {
		t.Fatalf("always-present empty close_reason lost: %#v", bank)
	}
	clearing := bank["clearing_system"].(map[string]any)
	if clearing["type"] != "bic_swift" || clearing["value"] != "BANKBHBM" {
		t.Fatalf("clearing_system lost: %#v", clearing)
	}
}

func TestVirtualAccountApplicationListOutputPreservesEnvelopeAndSummary(t *testing.T) {
	got := printAndDecodeJSON(t, `{
		"total_pages": 3,
		"total_items": 101,
		"data": [{
			"application_id": "app-summary",
			"public_version": 2,
			"country": "SG",
			"currency": "USD",
			"status": "COMPLETED",
			"created_at": "2026-08-12T10:00:00Z"
		}]
	}`)
	if got["total_pages"] != float64(3) || got["total_items"] != float64(101) {
		t.Fatalf("pagination envelope lost: %#v", got)
	}
	summary := got["data"].([]any)[0].(map[string]any)
	for key, want := range map[string]any{
		"application_id": "app-summary",
		"public_version": float64(2),
		"country":        "SG",
		"currency":       "USD",
		"status":         "COMPLETED",
		"created_at":     "2026-08-12T10:00:00Z",
	} {
		if summary[key] != want {
			t.Errorf("summary %s = %#v, want %#v", key, summary[key], want)
		}
	}
}
