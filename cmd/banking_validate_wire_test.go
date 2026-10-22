package cmd

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestBankingValidateWireBothEntrypoints(t *testing.T) {
	// Isolate CLI config/token cache, as in the existing command contract tests.
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	previous := http.DefaultTransport
	defer func() { http.DefaultTransport = previous }()
	originalStdout := os.Stdout
	defer func() { os.Stdout = originalStdout }()
	response := `{"currency":"USD","amount":"1.00","fee_amount":"5","total_amount":"6","channel_code":"MASTERCARD","sufficient_funds":false,"beneficiary_valid":true}`
	calls := 0
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := response
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			body = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			calls++
			if r.Method != "GET" || r.URL.Path != "/api/v1/payouts/validation" {
				t.Errorf("request: %s %s", r.Method, r.URL.Path)
			}
			want := map[string]string{"beneficiary_id": "saved + id", "currency": "USD", "amount": "1.00", "fee_paid_by": "OURS"}
			for key, value := range want {
				if r.URL.Query().Get(key) != value {
					t.Errorf("query %s = %s", key, r.URL.Query().Get(key))
				}
			}
			if r.Header.Get("x-on-behalf-of") != "sub-account" {
				t.Error("missing subaccount header")
			}
			if r.Body != nil && r.ContentLength > 0 {
				t.Error("GET has body")
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	for _, prefix := range [][]string{{"payout"}, {"banking", "payout"}} {
		args := append(append([]string{}, prefix...), "validate", "--beneficiary-id", "saved + id", "--currency", "USD", "--amount", "1.00", "--fee-paid-by", "OURS", "--on-behalf-of", "sub-account", "--output", "json")
		captured, err := os.CreateTemp(t.TempDir(), "stdout")
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = captured
		root := NewRootCmd()
		root.SetArgs(args)
		err = root.Execute()
		os.Stdout = originalStdout
		if err != nil {
			_ = captured.Close()
			t.Fatal(err)
		}
		if _, err := captured.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		var actual, expected any
		if err := json.NewDecoder(captured).Decode(&actual); err != nil {
			t.Fatal(err)
		}
		_ = captured.Close()
		_ = json.Unmarshal([]byte(response), &expected)
		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("JSON output lost fields: %+v", actual)
		}
	}
	if calls != 2 {
		t.Errorf("expected two business requests, got %d", calls)
	}
}
