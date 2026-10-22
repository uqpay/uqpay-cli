package cmd

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func TestPublicAssignMetadata(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	old := http.DefaultTransport
	defer func() { http.DefaultTransport = old }()
	var body map[string]interface{}
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		response := `{"card_id":"card-1"}`
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			response = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			if r.Method != "POST" || r.URL.Path != "/api/v1/issuing/cards/assign" {
				t.Fatalf("route %s %s", r.Method, r.URL.Path)
			}
			body = nil
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response)), Request: r}, nil
	})
	for _, sku := range []string{"", "SKU123456789"} {
		args := []string{"issuing", "card", "assign", "-d", "cardholder_id=holder-1", "-d", "card_number=example-pan", "-d", "card_currency=USD", "-d", "card_mode=SINGLE", "--output", "json"}
		expected := map[string]interface{}{"cardholder_id": "holder-1", "card_number": "example-pan", "card_currency": "USD", "card_mode": "SINGLE"}
		if sku != "" {
			args = append(args, "-d", "metadata.card_sku_code="+sku)
			expected["metadata"] = map[string]interface{}{"card_sku_code": sku}
		}
		root := NewRootCmd()
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(body, expected) {
			t.Fatalf("wire %#v", body)
		}
	}
	var help bytes.Buffer
	root := NewRootCmd()
	root.SetOut(&help)
	root.SetArgs([]string{"issuing", "card", "assign", "--help"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	for _, word := range []string{"metadata.card_sku_code", "12 characters", "Personal Visa", "Omit metadata"} {
		if !strings.Contains(help.String(), word) {
			t.Fatalf("help missing %s", word)
		}
	}
}
