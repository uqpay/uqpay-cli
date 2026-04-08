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
