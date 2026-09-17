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

func TestAlignedContractHelp(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		words []string
	}{
		{[]string{"issuing", "card", "set-pin", "--help"}, []string{"UPDATE", "old_pin", "PROCESSING"}},
		{[]string{"issuing", "card", "list", "--help"}, []string{"1-100"}},
		{[]string{"issuing", "card", "update", "--help"}, []string{"card_art_id", "physical", "PROCESSING"}},
		{[]string{"beneficiary", "check", "--help"}, []string{"bank_country_code", "At least one"}},
		{[]string{"simulate", "deposit", "--help"}, []string{"account_id"}},
		{[]string{"rfi", "answer", "--help"}, []string{"TEXT", "attachments"}},
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
				t.Errorf("%v help missing %s", tc.args, word)
			}
		}
	}
}

func TestAlignedContractWireAndOutput(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	previous := http.DefaultTransport
	defer func() { http.DefaultTransport = previous }()
	stdout := os.Stdout
	defer func() { os.Stdout = stdout }()
	var captured map[string]interface{}
	var path, response string
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := response
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			body = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			path = r.URL.Path
			captured = nil
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
					t.Fatal(err)
				}
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	cases := []struct {
		args                 []string
		path, want, response string
	}{
		{[]string{"beneficiary", "check", "-d", "entity_type=COMPANY", "-d", "payment_method=LOCAL", "-d", "currency=EUR", "-d", "iban=DE89370400440532013000", "-d", "bank_country_code=DE"}, "/v1/beneficiaries/check", `{"entity_type":"COMPANY","payment_method":"LOCAL","currency":"EUR","iban":"DE89370400440532013000","bank_country_code":"DE"}`, `{"valid":true,"reason":null}`},
		{[]string{"issuing", "card", "update", "card-1", "-d", "card_art_id=art-1", "-d", "name_on_card=Test"}, "/v1/issuing/cards/card-1", `{"card_art_id":"art-1","name_on_card":"Test"}`, `{"card_order_id":"art-order","order_status":"PROCESSING","amount":"12345678901234567890.12345678","metadata":null}`},
		{[]string{"issuing", "card", "set-pin", "-d", "card_id=card-1", "-d", "pin=135790", "-d", "type=UPDATE", "-d", "old_pin=024680"}, "/v1/issuing/cards/pin",
			`{"card_id":"card-1","pin":"135790","type":"UPDATE","old_pin":"024680"}`, `{"request_status":"SUCCESS","card_id":"card-1","card_order_id":"order-1","order_status":"PROCESSING","create_time":"2026-09-17T00:00:00Z"}`},
		{[]string{"rfi", "answer", "-d", "rfi_id=ACTREQ-test", "-d", "answer[0].key=note", "-d", "answer[0].type=TEXT", "-d", "answer[0].text=source of funds", "-d", "answer[1].key=document", "-d", "answer[1].type=ATTACHMENT", "-d", "answer[1].attachments[0]=file-1"}, "/v1/rfis/answer",
			`{"rfi_id":"ACTREQ-test","answer":[{"key":"note","type":"TEXT","text":"source of funds"},{"key":"document","type":"ATTACHMENT","attachments":["file-1"]}]}`, `{"rfi_id":"ACTREQ-test","request":[{"answer":{"attachments":[{"file_name":"proof.pdf","size":42}]}}]}`},
		{[]string{"simulate", "deposit", "-d", "account_id=account-1", "-d", "amount=10", "-d", "currency=SGD", "-d", "sender_swift_code=WELGBE22"}, "/v1/simulation/deposit",
			`{"account_id":"account-1","amount":10,"currency":"SGD","sender_swift_code":"WELGBE22"}`, `{"deposit_id":"deposit-1","amount":"10","deposit_status":"PENDING"}`},
		{[]string{"issuing", "transaction", "get", "tx-1"}, "/v1/issuing/transactions/tx-1", `null`, `{"transaction_id":"tx-1","transaction_amount":"123456789.01","settlement_status":"SETTLED"}`},
	}
	for _, tc := range cases {
		response = tc.response
		f, err := os.CreateTemp(t.TempDir(), "stdout")
		if err != nil {
			t.Fatal(err)
		}
		os.Stdout = f
		root := NewRootCmd()
		root.SetArgs(append([]string{"--output", "json"}, tc.args...))
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if !strings.HasSuffix(path, tc.path) {
			t.Fatalf("path %s want %s", path, tc.path)
		}
		var want map[string]interface{}
		_ = json.Unmarshal([]byte(tc.want), &want)
		got, _ := json.Marshal(captured)
		expected, _ := json.Marshal(want)
		if string(got) != string(expected) {
			t.Fatalf("wire: %s want %s", got, expected)
		}
		_, _ = f.Seek(0, 0)
		var out interface{}
		if err := json.NewDecoder(f).Decode(&out); err != nil {
			t.Fatal(err)
		}
		var expectedResponse interface{}
		_ = json.Unmarshal([]byte(tc.response), &expectedResponse)
		got, _ = json.Marshal(out)
		expected, _ = json.Marshal(expectedResponse)
		if string(got) != string(expected) {
			t.Fatalf("output: %s want %s", got, expected)
		}
		_ = f.Close()
	}
}
