package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type tokenEntry struct {
	AuthToken string `json:"auth_token"`
	ExpiredAt int64  `json:"expired_at"` // Unix timestamp seconds
}

type tokenCache struct {
	mu         sync.Mutex
	cachePath  string
	entries    map[string]*tokenEntry // keyed by env
	httpClient *http.Client
}

func newTokenCache(cachePath string) *tokenCache {
	return &tokenCache{
		cachePath:  cachePath,
		entries:    map[string]*tokenEntry{},
		httpClient: http.DefaultClient,
	}
}

// Get returns a valid token, fetching a new one if needed.
// Returns ("", nil) when clientID or apiKey is empty (test/no-auth mode).
func (tc *tokenCache) Get(ctx context.Context, env, clientID, apiKey, baseURL string) (string, error) {
	if clientID == "" || apiKey == "" {
		return "", nil
	}
	tc.mu.Lock()
	defer tc.mu.Unlock()

	entry := tc.load(env)
	if entry != nil && !tc.isExpiringSoon(entry) {
		return entry.AuthToken, nil
	}
	return tc.fetch(ctx, env, clientID, apiKey, baseURL)
}

// Invalidate removes the cached token for an env.
func (tc *tokenCache) Invalidate(env string) {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	delete(tc.entries, env)
	tc.persist()
}

func (tc *tokenCache) load(env string) *tokenEntry {
	if e, ok := tc.entries[env]; ok {
		return e
	}
	data, err := os.ReadFile(tc.cachePath)
	if err != nil {
		return nil
	}
	var all map[string]*tokenEntry
	if err := json.Unmarshal(data, &all); err != nil {
		return nil
	}
	for k, v := range all {
		tc.entries[k] = v
	}
	return tc.entries[env]
}

func (tc *tokenCache) isExpiringSoon(e *tokenEntry) bool {
	if e.ExpiredAt == 0 {
		return true
	}
	expiry := time.Unix(e.ExpiredAt, 0)
	return time.Until(expiry) < 5*time.Minute
}

func (tc *tokenCache) fetch(ctx context.Context, env, clientID, apiKey, baseURL string) (string, error) {
	url := baseURL + "/v1/connect/token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-client-id", clientID)
	req.Header.Set("x-api-key", apiKey)

	resp, err := tc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to obtain auth token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("auth token request failed with status %d", resp.StatusCode)
	}

	var result struct {
		AuthToken string `json:"auth_token"`
		ExpiredAt int64  `json:"expired_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	entry := &tokenEntry{AuthToken: result.AuthToken, ExpiredAt: result.ExpiredAt}
	tc.entries[env] = entry
	tc.persist()
	return entry.AuthToken, nil
}

func (tc *tokenCache) persist() {
	if tc.cachePath == "" {
		return
	}
	os.MkdirAll(filepath.Dir(tc.cachePath), 0700)
	data, _ := json.Marshal(tc.entries)
	os.WriteFile(tc.cachePath, data, 0600)
}
