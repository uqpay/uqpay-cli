package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

func TestIssuingAPISyncHelp(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		words []string
	}{
		{[]string{"issuing", "card", "create", "--help"}, []string{"Business Visa company authorized person", "one-step"}},
		{[]string{"issuing", "card", "assign", "--help"}, []string{"Business Visa may omit", "Personal Visa"}},
		{[]string{"issuing", "cardholder", "create", "--help"}, []string{"Business Mastercard", "Personal Visa", "proof_of_address_issue_date", "identity_verification"}},
		{[]string{"issuing", "cardholder", "update", "--help"}, []string{"supporting-material-only", "null file values preserve", "6 months"}},
	} {
		var out bytes.Buffer
		root := NewRootCmd()
		root.SetOut(&out)
		root.SetArgs(tc.args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		for _, word := range tc.words {
			if !strings.Contains(out.String(), word) {
				t.Fatalf("%v missing %s", tc.args, word)
			}
		}
	}
}

func TestIssuingAPISyncWireAndJSON(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	old := http.DefaultTransport
	defer func() { http.DefaultTransport = old }()
	var got map[string]interface{}
	var expectedPath string
	response := `{"cardholder_id":"holder-1","residential_address":{"line2":null}}`
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := response
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			body = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			if r.Method != http.MethodPost || r.URL.Path != expectedPath {
				t.Fatalf("route: %s %s; expected POST %s", r.Method, r.URL.Path, expectedPath)
			}
			got = nil
			if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
				t.Fatal(err)
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	for _, tc := range []struct {
		args []string
		want string
		path string
	}{
		{[]string{"issuing", "card", "create", "-d", "card_currency=USD", "-d", "card_product_id=product-1"}, `{"card_currency":"USD","card_product_id":"product-1"}`, "/v1/issuing/cards"},
		{[]string{"issuing", "card", "assign", "-d", "card_number=example-pan", "-d", "card_currency=USD", "-d", "card_mode=SINGLE"}, `{"card_number":"example-pan","card_currency":"USD","card_mode":"SINGLE"}`, "/v1/issuing/cards/assign"},
		{[]string{"issuing", "cardholder", "update", "holder-1", "-d", "identity.proof_of_address=cGRm", "-d", "identity.proof_of_address_issue_date=2026-10-01", "-d", "identity.proof_of_address_issuing_country=SG", "-d", "identity.identity_verification=a3lj"}, `{"identity":{"proof_of_address":"cGRm","proof_of_address_issue_date":"2026-10-01","proof_of_address_issuing_country":"SG","identity_verification":"a3lj"}}`, "/v1/issuing/cardholders/holder-1"},
	} {
		expectedPath = "/api" + tc.path
		file, err := os.CreateTemp(t.TempDir(), "output")
		if err != nil {
			t.Fatal(err)
		}
		previous := os.Stdout
		os.Stdout = file
		root := NewRootCmd()
		root.SetArgs(append(tc.args, "--output", "json"))
		err = root.Execute()
		os.Stdout = previous
		if err != nil {
			t.Fatal(err)
		}
		_, _ = file.Seek(0, 0)
		raw, _ := io.ReadAll(file)
		_ = file.Close()
		var actual, want interface{}
		if err := json.Unmarshal(raw, &actual); err != nil {
			t.Fatal(err)
		}
		_ = json.Unmarshal([]byte(response), &want)
		a, _ := json.Marshal(actual)
		b, _ := json.Marshal(want)
		if string(a) != string(b) {
			t.Fatalf("output %s", raw)
		}
		var wantBody map[string]interface{}
		_ = json.Unmarshal([]byte(tc.want), &wantBody)
		a, _ = json.Marshal(got)
		b, _ = json.Marshal(wantBody)
		if string(a) != string(b) {
			t.Fatalf("wire %s", a)
		}
	}
}
