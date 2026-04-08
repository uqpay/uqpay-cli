package dotparam_test

import (
	"reflect"
	"testing"

	"github.com/uqpay/uqpay-cli/internal/dotparam"
)

func TestFlatField(t *testing.T) {
	got, err := dotparam.Parse([]string{"currency=USD"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"currency": "USD"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDotNotation(t *testing.T) {
	// Numbers now stay as strings by default
	got, err := dotparam.Parse([]string{"bank_details.account_number=12345678"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"bank_details": map[string]any{
			"account_number": "12345678",
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestIndexedArray(t *testing.T) {
	// Numbers stay as strings; use num: prefix for actual numbers
	got, err := dotparam.Parse([]string{
		"spending_controls[0].amount=1000",
		"spending_controls[0].interval=PER_TRANSACTION",
		"spending_controls[1].amount=5000",
		"spending_controls[1].interval=MONTHLY",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"spending_controls": []any{
			map[string]any{"amount": "1000", "interval": "PER_TRANSACTION"},
			map[string]any{"amount": "5000", "interval": "MONTHLY"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestSimpleArrayAppend(t *testing.T) {
	got, err := dotparam.Parse([]string{
		"payment_method_types[]=card",
		"payment_method_types[]=googlepay",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"payment_method_types": []any{"card", "googlepay"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestDeepNesting(t *testing.T) {
	got, err := dotparam.Parse([]string{
		"cardholder_required_fields.residential_address.city=London",
		"cardholder_required_fields.residential_address.country=GB",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"cardholder_required_fields": map[string]any{
			"residential_address": map[string]any{
				"city":    "London",
				"country": "GB",
			},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestMissingEquals(t *testing.T) {
	_, err := dotparam.Parse([]string{"noequalssign"})
	if err == nil {
		t.Error("expected error for missing '='")
	}
}

func TestMultipleFields(t *testing.T) {
	got, err := dotparam.Parse([]string{
		"card_currency=USD",
		"cardholder_id=ch_123",
		"card_product_id=prod_456",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{
		"card_currency":   "USD",
		"cardholder_id":   "ch_123",
		"card_product_id": "prod_456",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestNumPrefix(t *testing.T) {
	// num: prefix forces integer
	got, err := dotparam.Parse([]string{"inherit=num:-1"})
	if err != nil {
		t.Fatal(err)
	}
	if got["inherit"] != int64(-1) {
		t.Errorf("expected int64(-1), got %T: %v", got["inherit"], got["inherit"])
	}
}

func TestNumPrefixFloat(t *testing.T) {
	got, err := dotparam.Parse([]string{"rate=num:3.14"})
	if err != nil {
		t.Fatal(err)
	}
	if got["rate"] != float64(3.14) {
		t.Errorf("expected float64(3.14), got %T: %v", got["rate"], got["rate"])
	}
}

func TestNumPrefixInvalid(t *testing.T) {
	// num: with non-numeric value falls back to string
	got, err := dotparam.Parse([]string{"field=num:abc"})
	if err != nil {
		t.Fatal(err)
	}
	if got["field"] != "abc" {
		t.Errorf("expected \"abc\", got %v", got["field"])
	}
}

func TestStrPrefix(t *testing.T) {
	// str: prefix still works for backward compatibility
	got, err := dotparam.Parse([]string{"merchant_category_code=str:5734"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := got["merchant_category_code"].(string)
	if !ok {
		t.Fatalf("expected string, got %T: %v", got["merchant_category_code"], got["merchant_category_code"])
	}
	if v != "5734" {
		t.Errorf("expected \"5734\", got %q", v)
	}
}

func TestNumbersStayString(t *testing.T) {
	// Pure numbers without prefix stay as strings
	got, err := dotparam.Parse([]string{"amount=100", "postal_code=10001"})
	if err != nil {
		t.Fatal(err)
	}
	if got["amount"] != "100" {
		t.Errorf("expected string \"100\", got %T: %v", got["amount"], got["amount"])
	}
	if got["postal_code"] != "10001" {
		t.Errorf("expected string \"10001\", got %T: %v", got["postal_code"], got["postal_code"])
	}
}

func TestBoolCoercion(t *testing.T) {
	got, err := dotparam.Parse([]string{"enabled=true", "disabled=false"})
	if err != nil {
		t.Fatal(err)
	}
	if got["enabled"] != true {
		t.Errorf("expected bool true, got %T: %v", got["enabled"], got["enabled"])
	}
	if got["disabled"] != false {
		t.Errorf("expected bool false, got %T: %v", got["disabled"], got["disabled"])
	}
}

func TestQuotedString(t *testing.T) {
	got, err := dotparam.Parse([]string{`name="hello world"`})
	if err != nil {
		t.Fatal(err)
	}
	if got["name"] != "hello world" {
		t.Errorf("expected \"hello world\", got %v", got["name"])
	}
}

func TestCoerceNumbersFlat(t *testing.T) {
	got, _ := dotparam.Parse([]string{"amount=100", "currency=USD"})
	dotparam.CoerceNumbers(got, "amount")
	if got["amount"] != int64(100) {
		t.Errorf("expected int64(100), got %T: %v", got["amount"], got["amount"])
	}
	if got["currency"] != "USD" {
		t.Errorf("expected string \"USD\", got %T: %v", got["currency"], got["currency"])
	}
}

func TestCoerceNumbersFloat(t *testing.T) {
	got, _ := dotparam.Parse([]string{"amount=10.50"})
	dotparam.CoerceNumbers(got, "amount")
	if got["amount"] != float64(10.50) {
		t.Errorf("expected float64(10.5), got %T: %v", got["amount"], got["amount"])
	}
}

func TestCoerceNumbersNested(t *testing.T) {
	got, _ := dotparam.Parse([]string{"tos_acceptance.tos_agreement=1", "tos_acceptance.ip=1.2.3.4"})
	dotparam.CoerceNumbers(got, "tos_agreement")
	tos := got["tos_acceptance"].(map[string]any)
	if tos["tos_agreement"] != int64(1) {
		t.Errorf("expected int64(1), got %T: %v", tos["tos_agreement"], tos["tos_agreement"])
	}
	if tos["ip"] != "1.2.3.4" {
		t.Errorf("expected string, got %T: %v", tos["ip"], tos["ip"])
	}
}

func TestCoerceNumbersArray(t *testing.T) {
	got, _ := dotparam.Parse([]string{
		"spending_controls[0].amount=500",
		"spending_controls[0].interval=PER_TRANSACTION",
	})
	dotparam.CoerceNumbers(got, "amount")
	arr := got["spending_controls"].([]any)
	item := arr[0].(map[string]any)
	if item["amount"] != int64(500) {
		t.Errorf("expected int64(500), got %T: %v", item["amount"], item["amount"])
	}
	if item["interval"] != "PER_TRANSACTION" {
		t.Errorf("expected string, got %T: %v", item["interval"], item["interval"])
	}
}

func TestCoerceNumbersNonNumericSkipped(t *testing.T) {
	got, _ := dotparam.Parse([]string{"amount=abc"})
	dotparam.CoerceNumbers(got, "amount")
	if got["amount"] != "abc" {
		t.Errorf("expected string \"abc\", got %T: %v", got["amount"], got["amount"])
	}
}

func TestCoerceNumbersEmptyFields(t *testing.T) {
	got, _ := dotparam.Parse([]string{"amount=100"})
	dotparam.CoerceNumbers(got) // no fields listed
	if got["amount"] != "100" {
		t.Errorf("expected string \"100\" unchanged, got %T: %v", got["amount"], got["amount"])
	}
}
