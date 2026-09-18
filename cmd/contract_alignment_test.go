package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestAlignedContractHelp(t *testing.T) {
	for _, tc := range []struct {
		args  []string
		words []string
	}{
		{[]string{"issuing", "card", "set-pin", "--help"}, []string{"UPDATE", "old_pin", "PROCESSING"}},
		{[]string{"conversion", "create", "--help"}, []string{"fresh quote", "FUNDS_ARRIVED", "TRADE_SETTLED"}},
		{[]string{"payment", "intent", "get", "--help"}, []string{"--on-behalf-of"}},
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
	var path, response, pageSize, pageNumber, status, method string
	var headers http.Header
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body := response
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			body = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			path = r.URL.Path
			method = r.Method
			pageNumber = r.URL.Query().Get("page_number")
			status = r.URL.Query().Get("status")
			pageSize = r.URL.Query().Get("page_size")
			headers = r.Header.Clone()
			captured = nil
			if r.Body != nil {
				if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
					t.Fatal(err)
				}
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	type contractCase struct {
		args                 []string
		path, want, response string
	}
	cases := []contractCase{
		{[]string{"payment", "intent", "get", "pi-1", "--on-behalf-of", "sub-account"}, "/v2/payment_intents/pi-1", `null`, `{"metadata":null,"latest_payment_attempt":null,"next_action":null,"complete_time":"","amount":"12345678901234567890.12345678"}`},
		{[]string{"account", "get", "account-1"}, "/v1/accounts/account-1", `null`, `{"entity_type":"COMPANY","business_details":{"legal_entity_name":"Example"}}`},
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

	// Frozen account summaries/details and issuing money: full CLI JSON output.
	moneyRaw, err := os.ReadFile("testdata/account-money.json")
	if err != nil {
		t.Fatal(err)
	}
	var moneyCases []struct {
		Operation string          `json:"operation"`
		Path      string          `json:"path"`
		Body      json.RawMessage `json:"body"`
	}
	if err := json.Unmarshal(moneyRaw, &moneyCases); err != nil {
		t.Fatal(err)
	}
	moneyCommands := map[string][]string{"accounts.list": {"account", "list"}, "accounts.get": {"account", "get", "account-1"}, "transactions.get": {"issuing", "transaction", "get", "tx-1"}, "transactions.list": {"issuing", "transaction", "list"}, "transfers.get": {"issuing", "transfer", "get", "transfer-1"}}
	for _, fixture := range moneyCases {
		args, ok := moneyCommands[fixture.Operation]
		if !ok {
			t.Fatal(fixture.Operation)
		}
		cases = append(cases, contractCase{args, fixture.Path, `null`, string(fixture.Body)})
	}
	// D044/D094: detail-only status; missing detail is legacy robustness.
	for _, status := range []string{"UNKNOWN", "UNSETTLED", "SETTLED", "NOT_APPLICABLE", ""} {
		body := `{"transaction_id":"tx-1"}`
		if status != "" {
			body = `{"transaction_id":"tx-1","settlement_status":"` + status + `"}`
		}
		cases = append(cases, contractCase{[]string{"issuing", "transaction", "get", "tx-1"}, "/v1/issuing/transactions/tx-1", `null`, body})
	}
	cases = append(cases, contractCase{[]string{"issuing", "transaction", "list"}, "/v1/issuing/transactions", `null`, `{"data":[{"transaction_id":"tx-1"}],"total_pages":1,"total_items":1}`})
	// RFI list/detail and PIN order fixtures are shared across the five clients.
	fixtureBytes, err := os.ReadFile("testdata/rfi-orders.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		RFIs   []json.RawMessage `json:"rfis"`
		Orders []json.RawMessage `json:"orders"`
	}
	if err := json.Unmarshal(fixtureBytes, &fixtures); err != nil {
		t.Fatal(err)
	}
	for _, raw := range fixtures.RFIs {
		var rfi struct {
			ID string `json:"rfi_id"`
		}
		if err := json.Unmarshal(raw, &rfi); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, contractCase{[]string{"rfi", "get", rfi.ID}, "/v1/rfis/" + rfi.ID, `null`, string(raw)}, contractCase{[]string{"rfi", "list", "--page-size", "10", "--page-num", "2", "--status", "ACTION_REQUIRED"}, "/v1/rfis", `null`, `{"data":[` + string(raw) + `],"total_pages":3,"total_items":21}`})
	}
	for _, raw := range fixtures.Orders {
		var order struct {
			ID string `json:"card_order_id"`
		}
		if err := json.Unmarshal(raw, &order); err != nil {
			t.Fatal(err)
		}
		cases = append(cases, contractCase{[]string{"issuing", "card", "get-order", order.ID}, "/v1/issuing/cards/" + order.ID + "/order", `null`, string(raw)})
	}
	// AQ-RESPONSE: actual CLI JSON output for independent response shapes.
	for _, tc := range []struct {
		args     []string
		path     string
		fixtures []string
	}{
		{[]string{"payment", "attempt", "get", "pa-1"}, "/v2/payment/payment_attempts/pa-1", []string{`{}`, `{"complete_time":"","advice_code":"","authentication_data":{"cvv_result":""}}`, `{"complete_time":"2026-09-17T00:00:00Z","advice_code":"01","authentication_data":{"cvv_result":"M"}}`}},
		{[]string{"payment", "refund", "get", "re-1"}, "/v2/payment/refunds/re-1", []string{`{}`, `{"metadata":null}`, `{"metadata":{}}`, `{"metadata":{"ref":"0001"}}`}},
		{[]string{"payment", "payout", "get", "po-1"}, "/v2/payment/payout/po-1", []string{`{}`, `{"completed_time":""}`, `{"completed_time":"2026-09-17T00:00:00Z"}`}},
	} {
		for _, fixture := range tc.fixtures {
			cases = append(cases, contractCase{tc.args, tc.path, `null`, fixture})
		}
	}

	// D122-D129: distinct string values on every balance field, list and detail.
	fields := []string{"available_balance", "frozen_balance", "margin_balance", "prepaid_balance"}
	amounts := []string{"0.00", "1.23", "-0.01", "12345678901234567890.12", "-12345678901234567890.12", "0.12345678901234567890"}
	for i := range amounts {
		balance := map[string]interface{}{"currency": "USD"}
		for j, field := range fields {
			balance[field] = amounts[(i+j)%len(amounts)]
		}
		detail, _ := json.Marshal(balance)
		list, _ := json.Marshal(map[string]interface{}{"data": []interface{}{balance}, "total_pages": 1, "total_items": 1})
		cases = append(cases, contractCase{[]string{"banking", "balance", "get", "USD"}, "/v1/balances/USD", `null`, string(detail)}, contractCase{[]string{"banking", "balance", "list"}, "/v1/balances", `null`, string(list)})
	}

	// D189-D196: all changed GET routes, with and without delegation.
	for _, route := range []struct {
		args []string
		path string
	}{
		{[]string{"payment", "balance", "list"}, "/v2/payment/balances"},
		{[]string{"payment", "balance", "get", "USD"}, "/v2/payment/balances/USD"},
		{[]string{"payment", "bank-account", "list"}, "/v2/payment/bankaccount"},
		{[]string{"payment", "bank-account", "get", "ba-1"}, "/v2/payment/bankaccount/ba-1"},
		{[]string{"payment", "payout", "list"}, "/v2/payment/payout"},
		{[]string{"payment", "payout", "get", "po-1"}, "/v2/payment/payout/po-1"},
		{[]string{"payment", "settlement", "list"}, "/v2/payment/settlements"},
		{[]string{"payment", "intent", "get", "pi-1"}, "/v2/payment_intents/pi-1"},
	} {
		for _, account := range []string{"sub-account", ""} {
			args := append([]string{}, route.args...)
			if account != "" {
				args = append(args, "--on-behalf-of", account)
			}
			cases = append(cases, contractCase{args, route.path, `null`, `{}`})
		}
	}

	for _, size := range []int{1, 10, 100} {
		cases = append(cases, contractCase{[]string{"issuing", "card", "list", "--page-size", strconv.Itoa(size)}, "/v1/issuing/cards", `null`, `{"data":[{"card_id":"card-1","card_limit":"12345678901234567890.12345678","metadata":"{\"ref\":\"0001\"}","risk_controls":null}]}`})
	}
	for _, provider := range []string{"SUMSUB", "MYINFO", "JUMIO", "DIDIT", "SHUFTI", "REGTANK"} {
		for _, length := range []int{9, 10, 64, 65} {
			for _, dob := range []string{"2009-09-17", "2008-09-17", "1947-09-17", "1946-09-17"} {
				fields := map[string]interface{}{"email": "test@example.test", "first_name": "Test", "last_name": "User", "country_code": "SG", "date_of_birth": dob, "kyc_verification": map[string]interface{}{"method": "THIRD_PARTY", "kyc_proof": map[string]interface{}{"provider": provider, "reference_id": strings.Repeat("r", length)}}}
				for _, op := range []string{"create", "update", "card"} {
					args := []string{"issuing", "cardholder", op}
					path := "/v1/issuing/cardholders"
					prefix := ""
					var body interface{} = fields
					if op == "update" {
						args = append(args, "holder-1")
						path += "/holder-1"
					}
					if op == "card" {
						args = []string{"issuing", "card", "create"}
						path = "/v1/issuing/cards"
						prefix = "cardholder_required_fields."
						body = map[string]interface{}{"cardholder_required_fields": fields}
					}
					for k, v := range map[string]string{"email": "test@example.test", "first_name": "Test", "last_name": "User", "country_code": "SG", "date_of_birth": dob, "kyc_verification.method": "THIRD_PARTY", "kyc_verification.kyc_proof.provider": provider, "kyc_verification.kyc_proof.reference_id": strings.Repeat("r", length)} {
						args = append(args, "-d", prefix+k+"="+v)
					}
					raw, _ := json.Marshal(body)
					cases = append(cases, contractCase{args, path, string(raw), `{}`})
				}
			}
		}
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
		if len(tc.args) > 2 && tc.args[0] == "payment" && (tc.args[2] == "get" || tc.args[2] == "list") {
			expectedAccount := ""
			for i, arg := range tc.args {
				if arg == "--on-behalf-of" {
					expectedAccount = tc.args[i+1]
				}
			}
			if headers.Get("x-on-behalf-of") != expectedAccount || headers.Get("x-client-id") != "offline-client" {
				t.Fatalf("%v: headers %v", tc.args, headers)
			}
		}

		if tc.args[0] == "rfi" && tc.args[1] == "list" {
			if pageSize != "10" || pageNumber != "2" || status != "ACTION_REQUIRED" {
				t.Fatalf("RFI query: %s %s %s", pageSize, pageNumber, status)
			}
		}
		if tc.want == "null" && method != "GET" {
			t.Fatalf("expected GET, got %s", method)
		}
		if pageSize != "" {
			for i, arg := range tc.args {
				if arg == "--page-size" && pageSize != tc.args[i+1] {
					t.Fatal("page size changed")
				}
			}
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
