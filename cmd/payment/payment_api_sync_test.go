package payment

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
)

type paymentSyncTransport func(*http.Request) (*http.Response, error)

func (f paymentSyncTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }
func TestPaymentAPISync(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	prior := http.DefaultTransport
	defer func() { http.DefaultTransport = prior }()
	var bodies []map[string]interface{}
	http.DefaultTransport = paymentSyncTransport(func(r *http.Request) (*http.Response, error) {
		response := `{"capture_id":"cap-1","intent_status":"PARTIALLY_CAPTURED","authorization_expires_at":null,"capture_records":[{"channel_capture_status":0}]}`
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			response = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			expectedPaths := []string{"/api/v2/payment_intents/create", "/api/v2/payment_intents/pi-1/confirm", "/api/v2/payment_intents/pi-1/capture"}
			if len(bodies) >= len(expectedPaths) || r.Method != http.MethodPost || r.URL.Path != expectedPaths[len(bodies)] {
				t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
				return &http.Response{StatusCode: 400, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"error":"unexpected route"}`)), Request: r}, nil
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			bodies = append(bodies, body)
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response)), Request: r}, nil
	})
	original := os.Stdout
	defer func() { os.Stdout = original }()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	os.Stdout = w
	create := newPaymentIntentCreateCmd()
	create.SetArgs([]string{"-d", "authorization_type=pre_authorization", "-d", "payment_method.type=card_present", "-d", "payment_method.card_present.auto_capture=false"})
	if err := create.Execute(); err != nil {
		t.Fatal(err)
	}
	confirm := newPaymentIntentConfirmCmd()
	confirm.SetArgs([]string{"pi-1", "-d", "authorization_type=pre_authorization"})
	if err := confirm.Execute(); err != nil {
		t.Fatal(err)
	}
	capture := newPaymentIntentCaptureCmd()
	capture.SetArgs([]string{"pi-1", "-d", "amount_to_capture=6.00", "-d", "merchant_capture_order_id=order-cap-1", "-d", "final_capture=true"})
	if err := capture.Execute(); err != nil {
		t.Fatal(err)
	}
	w.Close()
	output, _ := io.ReadAll(r)
	if bodies[0]["authorization_type"] != "pre_authorization" || bodies[1]["authorization_type"] != "pre_authorization" {
		t.Fatal(bodies)
	}
	if bodies[0]["payment_method"].(map[string]interface{})["card_present"].(map[string]interface{})["auto_capture"] != false {
		t.Fatal(bodies[0])
	}
	if bodies[2]["amount_to_capture"] != "6.00" || bodies[2]["final_capture"] != true || bodies[2]["merchant_capture_order_id"] != "order-cap-1" {
		t.Fatal(bodies[2])
	}
	if !strings.Contains(string(output), "cap-1") || !strings.Contains(string(output), "PARTIALLY_CAPTURED") {
		t.Fatal(string(output))
	}
}
