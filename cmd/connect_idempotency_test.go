package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCreateSubExplicitAndAutomaticIdempotency(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Setenv("UQPAY_CLIENT_ID", "offline-client")
	t.Setenv("UQPAY_API_KEY", "offline-key")
	t.Setenv("UQPAY_ENV", "sandbox")
	previous := http.DefaultTransport
	defer func() { http.DefaultTransport = previous }()
	original := os.Stdout
	sink, err := os.CreateTemp(t.TempDir(), "stdout")
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = sink
	defer func() { os.Stdout = original; sink.Close() }()
	var keys, bodies []string
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		response := `{}`
		if strings.HasSuffix(r.URL.Path, "/v1/connect/token") {
			response = `{"auth_token":"offline-token","expired_at":4102444800}`
		} else {
			if r.Method != "POST" || r.URL.Path != "/api/v1/accounts/create_accounts" {
				t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			}
			body, e := io.ReadAll(r.Body)
			if e != nil {
				t.Fatal(e)
			}
			keys = append(keys, r.Header.Get("x-idempotency-key"))
			bodies = append(bodies, string(body))
		}
		return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(response)), Request: r}, nil
	})
	const key = "11111111-1111-4111-8111-111111111111"
	for _, extra := range [][]string{{"--idempotency-key", key}, {"--idempotency-key", key}, nil, nil} {
		root := NewRootCmd()
		root.SetArgs(append([]string{"account", "create-sub", "-d", "entity_type=COMPANY", "-d", "nickname=synthetic", "--output", "json"}, extra...))
		if err := root.Execute(); err != nil {
			t.Fatal(err)
		}
	}
	if len(keys) != 4 {
		t.Fatalf("requests: %d", len(keys))
	}
	if keys[0] != key || keys[1] != key || bodies[0] != bodies[1] {
		t.Fatalf("explicit retry changed key or body")
	}
	if keys[2] == keys[3] || keys[2] == key || keys[3] == key {
		t.Fatalf("automatic keys not fresh")
	}
	for _, k := range keys {
		if _, err := uuid.Parse(k); err != nil {
			t.Fatalf("invalid UUID: %q", k)
		}
	}
}

func TestCreateSubRejectsInvalidIdempotencyBeforeHTTP(t *testing.T) {
	previous := http.DefaultTransport
	defer func() { http.DefaultTransport = previous }()
	http.DefaultTransport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		t.Fatalf("unexpected request: %s", r.URL.Path)
		return nil, nil
	})
	for _, key := range []string{"", "not-a-uuid", "AAAAAAAA-AAAA-4AAA-8AAA-AAAAAAAAAAAA", "11111111111141118111111111111111"} {
		root := NewRootCmd()
		root.SetOut(io.Discard)
		root.SetErr(io.Discard)
		root.SetArgs([]string{"account", "create-sub", "--idempotency-key", key})
		if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "lowercase UUID") {
			t.Fatalf("key %q: %v", key, err)
		}
	}
}
