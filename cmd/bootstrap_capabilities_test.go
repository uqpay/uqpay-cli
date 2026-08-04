package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type capturedRequest struct {
	method string
	path   string
	query  string
	header http.Header
	body   string
}

func TestBootstrapCommandsMatchPublishedContract(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "client_123")
	t.Setenv("UQPAY_API_KEY", "api_key_123")
	t.Setenv("UQPAY_ENV", "sandbox")

	originalTransport := http.DefaultTransport
	t.Cleanup(func() { http.DefaultTransport = originalTransport })

	var captured []capturedRequest
	http.DefaultTransport = roundTripFunc(func(request *http.Request) (*http.Response, error) {
		responseBody := `{}`
		if strings.HasSuffix(request.URL.Path, "/v1/connect/token") {
			responseBody = `{"auth_token":"token_123","expired_at":4102444800}`
		} else {
			body := ""
			if request.Body != nil {
				data, err := io.ReadAll(request.Body)
				if err != nil {
					t.Fatalf("read request body: %v", err)
				}
				body = string(data)
			}
			captured = append(captured, capturedRequest{
				method: request.Method,
				path:   strings.TrimPrefix(request.URL.Path, "/api"),
				query:  request.URL.RawQuery,
				header: request.Header.Clone(),
				body:   body,
			})
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(responseBody)),
			Request:    request,
		}, nil
	})

	originalStdout := os.Stdout
	quiet, err := os.CreateTemp(t.TempDir(), "uqpay-cli-output-*")
	if err != nil {
		t.Fatalf("create output file: %v", err)
	}
	os.Stdout = quiet
	t.Cleanup(func() {
		os.Stdout = originalStdout
		_ = quiet.Close()
	})

	commands := [][]string{
		{"--output", "json", "rfi", "list", "--status", "ACTION_REQUIRED"},
		{"--output", "json", "rfi", "get", "rfi_123"},
		{"--output", "json", "rfi", "answer", "-d", "rfi_id=rfi_123"},
		{"--output", "json", "issuing", "card", "elevate-limit", "card_123", "-d", "limit_amount=1000"},
		{"--output", "json", "issuing", "card", "enroll-network-protection", "card_123", "--action-code", "41"},
		{"--output", "json", "issuing", "card", "remove-network-protection", "card_123"},
		{"--output", "json", "issuing", "card", "manage-pin", "-d", "card_id=card_123", "-d", "type=SET", "-d", "pin=1234"},
		{"--output", "json", "issuing", "card", "list-arts", "--card-product-id", "product_123"},
		{"--output", "json", "issuing", "card", "set-default-art", "art_123"},
		{"--output", "json", "issuing", "merchant-brand", "--display-name", "Grab"},
		{"--output", "json", "issuing", "transaction", "claim-unsolicited-refund", "-d", "related_transaction_id=txn_123"},
		{"--output", "json", "payment", "terminal", "register", "-d", "firm_code=01", "-d", "firm_sn=SN123", "-d", "terminal_model=PAX A920"},
		{"--output", "json", "payment", "terminal", "get-pin-key", "-d", "terminal_id=terminal_123", "-d", "prv_key=key"},
	}
	for _, args := range commands {
		root := NewRootCmd()
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatalf("uqpay %s returned an error: %v", strings.Join(args, " "), err)
		}
	}

	wantMethods := []string{
		"GET", "GET", "POST", "POST", "POST", "DELETE", "POST",
		"GET", "POST", "GET", "POST", "POST", "POST",
	}
	wantPaths := []string{
		"/v1/rfis",
		"/v1/rfis/rfi_123",
		"/v1/rfis/answer",
		"/v1/issuing/cards/card_123/elevate_limit",
		"/v1/issuing/cards/card_123/risk",
		"/v1/issuing/cards/card_123/risk",
		"/v1/issuing/cards/manage/pin",
		"/v1/issuing/cards/arts",
		"/v1/issuing/cards/arts/default",
		"/v1/issuing/merchant_brands",
		"/v1/issuing/transactions/unsolicited_refund/release",
		"/v2/terminal/register",
		"/v2/terminal/getPinKey",
	}
	if len(captured) != len(wantPaths) {
		t.Fatalf("captured %d API requests, want %d", len(captured), len(wantPaths))
	}
	for i := range captured {
		if captured[i].method != wantMethods[i] || captured[i].path != wantPaths[i] {
			t.Errorf("request %d = %s %s, want %s %s",
				i, captured[i].method, captured[i].path, wantMethods[i], wantPaths[i])
		}
	}

	if !strings.Contains(captured[5].body, `"risk_control":"network_protection"`) {
		t.Errorf("DELETE body = %q, want risk_control", captured[5].body)
	}
	if !strings.Contains(captured[10].body, `"related_transaction_id":"txn_123"`) {
		t.Errorf("claim body = %q, want related_transaction_id", captured[10].body)
	}
	if captured[11].header.Get("x-client-id") != "client_123" ||
		captured[12].header.Get("x-client-id") != "client_123" {
		t.Error("terminal requests do not include the configured x-client-id")
	}
	if !strings.Contains(captured[0].query, "status=ACTION_REQUIRED") {
		t.Errorf("RFI list query = %q, want status filter", captured[0].query)
	}
}
