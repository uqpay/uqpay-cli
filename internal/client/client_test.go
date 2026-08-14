package client_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
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
		ClientID: "", // empty = no token fetch attempted
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

func TestVirtualAccountCreatePreservesExplicitIdempotencyKeyAcrossReplay(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/virtual/accounts" || r.Method != http.MethodPost {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		keys = append(keys, r.Header.Get("x-idempotency-key"))
		if got := r.Header.Get("x-on-behalf-of"); got != "acct-sub" {
			t.Errorf("x-on-behalf-of = %q", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"data": map[string]any{
			"application_id": "app-1", "public_version": 1, "country": "SG",
			"currency": "USD", "status": "SUBMITTED", "results": []any{},
		}})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	headers := map[string]string{"x-idempotency-key": "va-retry-001", "x-on-behalf-of": "acct-sub"}
	body := map[string]any{"country": "SG", "currency": "USD"}
	for i := 0; i < 2; i++ {
		if _, err := c.PostH(context.Background(), "/v1/virtual/accounts", body, headers); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 2 || keys[0] != "va-retry-001" || keys[1] != keys[0] {
		t.Fatalf("idempotent replay keys = %#v", keys)
	}
}

func TestConfiguredClientIDIsSentOnBusinessRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1/connect/token" {
			json.NewEncoder(w).Encode(map[string]any{
				"auth_token": "token_123",
				"expired_at": time.Now().Add(time.Hour).Unix(),
			})
			return
		}
		if got := r.Header.Get("x-client-id"); got != "client_123" {
			t.Errorf("x-client-id = %q, want client_123", got)
		}
		json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()

	cfg := &config.Config{ClientID: "client_123", APIKey: "api_key_123", Env: "sandbox"}
	c := client.NewWithBaseURL(cfg, srv.URL, t.TempDir())
	if _, err := c.Get(context.Background(), "/v2/payment/balances", nil); err != nil {
		t.Fatal(err)
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
	if apiErr.APIType != "not_found" || apiErr.APICode != "card_not_found" {
		t.Errorf("strict API fields lost: type=%q code=%q", apiErr.APIType, apiErr.APICode)
	}
}

func TestVirtualAccountApplicationNotFoundPreservesStrict400Body(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"type": "not_found", "code": "virtual_account_application_not_found",
			"message": "Virtual account application not found",
		})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Get(context.Background(), "/v1/virtual/applications/app-missing", nil)
	var apiErr *apierr.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 || apiErr.APIType != "not_found" ||
		apiErr.APICode != "virtual_account_application_not_found" ||
		apiErr.Message != "Virtual account application not found" {
		t.Fatalf("unexpected strict error: %#v", apiErr)
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

func TestWriteRetriesReuseAutomaticIdempotencyKey(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("x-idempotency-key"))
		if len(keys) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]any{"message": "rate limited"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "ok"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(time.Millisecond)
	if _, err := c.Post(context.Background(), "/v1/write", map[string]any{"amount": "1.00"}); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 3 || keys[0] == "" || keys[1] != keys[0] || keys[2] != keys[0] {
		t.Fatalf("write retry keys = %#v", keys)
	}
}

func TestWriteRetryPreservesCallerIdempotencyKey(t *testing.T) {
	var keys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		keys = append(keys, r.Header.Get("x-idempotency-key"))
		if len(keys) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]any{"message": "rate limited"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "ok"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(time.Millisecond)
	headers := map[string]string{"X-Idempotency-Key": "caller-stable-key"}
	if _, err := c.PostH(context.Background(), "/v1/write", map[string]any{"amount": "1.00"}, headers); err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || keys[0] != "caller-stable-key" || keys[1] != keys[0] {
		t.Fatalf("caller retry keys = %#v", keys)
	}
}

func TestTokenRefreshReusesWriteIdempotencyKey(t *testing.T) {
	var tokenRequests int
	var writeKeys []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/connect/token" {
			tokenRequests++
			json.NewEncoder(w).Encode(map[string]any{
				"auth_token": fmt.Sprintf("token-%d", tokenRequests),
				"expired_at": time.Now().Add(time.Hour).Unix(),
			})
			return
		}
		writeKeys = append(writeKeys, r.Header.Get("x-idempotency-key"))
		if len(writeKeys) == 1 {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]any{"error": "token has expired"})
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"id": "ok"})
	}))
	defer srv.Close()

	cfg := &config.Config{ClientID: "client", APIKey: "api-key", Env: "sandbox"}
	c := client.NewWithBaseURL(cfg, srv.URL, t.TempDir())
	if _, err := c.Post(context.Background(), "/v1/write", map[string]any{"amount": "1.00"}); err != nil {
		t.Fatal(err)
	}
	if tokenRequests != 2 {
		t.Fatalf("token requests = %d, want 2", tokenRequests)
	}
	if len(writeKeys) != 2 || writeKeys[0] == "" || writeKeys[1] != writeKeys[0] {
		t.Fatalf("token refresh write keys = %#v", writeKeys)
	}
}

func TestAmbiguousWriteNetworkFailureRequiresReconciliationWithoutRetry(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("test server does not support hijacking")
		}
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(time.Millisecond)
	_, err := c.Post(context.Background(), "/v1/write", map[string]any{"amount": "1.00"})
	var reconcileErr *apierr.ReconcileRequiredError
	if !errors.As(err, &reconcileErr) {
		t.Fatalf("expected ReconcileRequiredError, got %T: %v", err, err)
	}
	if got := attempts.Load(); got != 1 {
		t.Fatalf("ambiguous write attempts = %d, want 1", got)
	}
	if reconcileErr.Method != http.MethodPost || reconcileErr.Path != "/v1/write" || reconcileErr.IdempotencyKey == "" {
		t.Fatalf("unexpected reconcile context: %#v", reconcileErr)
	}
}

func TestAmbiguousWriteServerErrorRequiresReconciliationWithoutRetry(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{"message": "internal error"})
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	c.SetRetryDelay(time.Millisecond)
	_, err := c.Post(context.Background(), "/v1/write", map[string]any{"amount": "1.00"})
	var reconcileErr *apierr.ReconcileRequiredError
	if !errors.As(err, &reconcileErr) {
		t.Fatalf("expected ReconcileRequiredError, got %T: %v", err, err)
	}
	if attempts != 1 {
		t.Fatalf("ambiguous 5xx write attempts = %d, want 1", attempts)
	}
}

func TestAmbiguousWriteResponseReadFailureRequiresReconciliationWithoutRetry(t *testing.T) {
	var attempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.Header().Set("Content-Length", "100")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"partial"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, err := c.Delete(context.Background(), "/v1/write", nil)
	var reconcileErr *apierr.ReconcileRequiredError
	if !errors.As(err, &reconcileErr) {
		t.Fatalf("expected ReconcileRequiredError, got %T: %v", err, err)
	}
	if attempts != 1 {
		t.Fatalf("ambiguous response-read attempts = %d, want 1", attempts)
	}
}

func TestAmbiguousMultipartFailurePreservesCallerKeyForReconciliation(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		if got := r.Header.Get("x-idempotency-key"); got != "upload-stable-key" {
			t.Errorf("x-idempotency-key = %q", got)
		}
		hijacker := w.(http.Hijacker)
		conn, _, err := hijacker.Hijack()
		if err != nil {
			t.Fatal(err)
		}
		conn.Close()
	}))
	defer srv.Close()

	filePath := filepath.Join(t.TempDir(), "document.txt")
	if err := os.WriteFile(filePath, []byte("test"), 0o600); err != nil {
		t.Fatal(err)
	}
	c := newTestClient(t, srv)
	_, err := c.PostMultipartH(
		context.Background(),
		"/v1/files/upload",
		filePath,
		nil,
		map[string]string{"x-idempotency-key": "upload-stable-key"},
	)
	var reconcileErr *apierr.ReconcileRequiredError
	if !errors.As(err, &reconcileErr) {
		t.Fatalf("expected ReconcileRequiredError, got %T: %v", err, err)
	}
	if attempts.Load() != 1 || reconcileErr.IdempotencyKey != "upload-stable-key" {
		t.Fatalf("unexpected multipart reconcile context: attempts=%d error=%#v", attempts.Load(), reconcileErr)
	}
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
