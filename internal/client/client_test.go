package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/config"
)

// newTestClient creates a Client pointing at a test server, with token auth bypassed.
func newTestClient(t *testing.T, apiServer *httptest.Server) *client.Client {
	t.Helper()
	cfg := &config.Config{
		ClientID: "",  // empty = no token fetch attempted
		APIKey:   "",
		Env:      "sandbox",
	}
	return client.NewWithBaseURL(cfg, apiServer.URL, t.TempDir())
}

func TestGetSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/v1/issuing/cards" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, err := c.Get(context.Background(), "/v1/issuing/cards", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty response")
	}
}

func TestPostSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["card_currency"] != "USD" {
			t.Errorf("unexpected body: %v", body)
		}
		if r.Header.Get("x-idempotency-key") == "" {
			t.Error("missing x-idempotency-key header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "card_123"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, err := c.Post(context.Background(), "/v1/issuing/cards", map[string]any{"card_currency": "USD"})
	if err != nil {
		t.Fatal(err)
	}
	var resp map[string]any
	json.Unmarshal(data, &resp)
	if resp["id"] != "card_123" {
		t.Errorf("unexpected response: %v", resp)
	}
}

func TestGetQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("status") != "ACTIVE" {
			t.Errorf("expected status=ACTIVE, got %s", r.URL.Query().Get("status"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": []any{}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Get(context.Background(), "/v1/issuing/cards", map[string]string{"status": "ACTIVE"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestAPIError404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]any{
			"type":    "not_found",
			"code":    "card_not_found",
			"message": "Card does not exist",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Get(context.Background(), "/v1/issuing/cards/bad_id", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *apierr.APIError, got %T", err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
	}
	if apiErr.Message != "Card does not exist" {
		t.Errorf("Message = %q", apiErr.Message)
	}
}

func TestAPIErrorLegacyFormat(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		// Legacy format: body.code=200 but HTTP 400 — must trust HTTP status
		json.NewEncoder(w).Encode(map[string]any{
			"code":    200,
			"message": "Invalid parameter",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Get(context.Background(), "/v1/test", nil)
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *apierr.APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400 (must trust HTTP status, not body.code)", apiErr.StatusCode)
	}
}

func TestRetryOn429(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(429)
			json.NewEncoder(w).Encode(map[string]any{"message": "rate limited"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"id": "ok"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(1 * time.Millisecond) // speed up test
	data, err := c.Get(context.Background(), "/v1/test", nil)
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
	_ = data
}

func TestRetryExhausted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]any{"message": "internal error"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(1 * time.Millisecond)
	_, err := c.Get(context.Background(), "/v1/test", nil)
	if err == nil {
		t.Fatal("expected error after exhausted retries")
	}
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected *apierr.APIError, got %T", err)
	}
	if apiErr.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestDebugOutput(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	cfg := &config.Config{ClientID: "", APIKey: "", Env: "sandbox", Debug: true}
	c := client.NewWithBaseURL(cfg, srv.URL, t.TempDir())

	var dbg bytes.Buffer
	c.SetDebugOut(&dbg)

	if _, err := c.Get(context.Background(), "/test", nil); err != nil {
		t.Fatal(err)
	}

	got := dbg.String()
	if !strings.Contains(got, "GET") {
		t.Errorf("debug output missing method: %q", got)
	}
	if !strings.Contains(got, "/test") {
		t.Errorf("debug output missing path: %q", got)
	}
	if !strings.Contains(got, "200") {
		t.Errorf("debug output missing response status: %q", got)
	}
}

func TestDebugOff(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	cfg := &config.Config{ClientID: "", APIKey: "", Env: "sandbox", Debug: false}
	c := client.NewWithBaseURL(cfg, srv.URL, t.TempDir())

	var dbg bytes.Buffer
	c.SetDebugOut(&dbg)

	if _, err := c.Get(context.Background(), "/test", nil); err != nil {
		t.Fatal(err)
	}

	if dbg.Len() > 0 {
		t.Errorf("expected no debug output when debug=false, got: %q", dbg.String())
	}
}
