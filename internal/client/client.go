package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/config"
)

var baseURLs = map[string]string{
	"sandbox":    "https://api-sandbox.uqpaytech.com/api",
	"production": "https://api.uqpay.com/api",
}

var fileBaseURLs = map[string]string{
	"sandbox":    "https://files.uqpaytech.com/api",
	"production": "https://files.uqpay.com/api",
}

var iframeBaseURLs = map[string]string{
	"sandbox":    "https://embedded-sandbox.uqpaytech.com",
	"production": "https://embedded.uqpay.com",
}

// Client is the UQPAY HTTP client.
type Client struct {
	cfg                *config.Config
	baseURL            string
	iframeBase         string
	tokens             *tokenCache
	http               *http.Client
	retryDelayOverride time.Duration
	debug              bool
	debugOut           io.Writer
}

// New creates a Client using real API base URLs from config.
func New(cfg *config.Config) *Client {
	base := baseURLs[cfg.Env]
	if base == "" {
		base = baseURLs["sandbox"]
	}
	iframe := iframeBaseURLs[cfg.Env]
	if iframe == "" {
		iframe = iframeBaseURLs["sandbox"]
	}
	tokenPath := config.TokenCachePath()
	httpClient := &http.Client{Timeout: 30 * time.Second}
	tc := newTokenCache(tokenPath)
	tc.httpClient = httpClient
	return &Client{
		cfg:        cfg,
		baseURL:    base,
		iframeBase: iframe,
		tokens:     tc,
		http:       httpClient,
		debug:      cfg.Debug,
		debugOut:   os.Stderr,
	}
}

// NewFileClient creates a Client pointed at the file-service base URL.
func NewFileClient(cfg *config.Config) *Client {
	base := fileBaseURLs[cfg.Env]
	if base == "" {
		base = fileBaseURLs["sandbox"]
	}
	httpClient := &http.Client{Timeout: 60 * time.Second}
	tc := newTokenCache(config.TokenCachePath())
	tc.httpClient = httpClient
	return &Client{
		cfg:      cfg,
		baseURL:  base,
		tokens:   tc,
		http:     httpClient,
		debug:    cfg.Debug,
		debugOut: os.Stderr,
	}
}

// NewWithBaseURL creates a Client with a custom base URL (for testing).
func NewWithBaseURL(cfg *config.Config, baseURL, cacheDir string) *Client {
	httpClient := &http.Client{Timeout: 30 * time.Second}
	tc := newTokenCache(cacheDir + "/token.json")
	tc.httpClient = httpClient
	return &Client{
		cfg:        cfg,
		baseURL:    baseURL,
		iframeBase: baseURL,
		tokens:     tc,
		http:       httpClient,
		debug:      cfg.Debug,
		debugOut:   os.Stderr,
	}
}

// SetRetryDelay overrides the retry delay (for testing).
func (c *Client) SetRetryDelay(d time.Duration) {
	c.retryDelayOverride = d
}

// SetDebugOut overrides the writer used for debug output (default: os.Stderr).
func (c *Client) SetDebugOut(w io.Writer) {
	c.debugOut = w
}

// IframeBase returns the iframe base URL for the current environment.
func (c *Client) IframeBase() string {
	return c.iframeBase
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, query map[string]string) ([]byte, error) {
	return c.doAttempt(ctx, http.MethodGet, path, query, nil, nil, false, 0, false)
}

// GetH performs a GET request with extra headers.
func (c *Client) GetH(ctx context.Context, path string, query map[string]string, headers map[string]string) ([]byte, error) {
	return c.doAttempt(ctx, http.MethodGet, path, query, nil, headers, false, 0, false)
}

// GetHI performs a GET request with extra headers and an idempotency key.
// Use for APIs that require x-idempotency-key on GET requests.
func (c *Client) GetHI(ctx context.Context, path string, query map[string]string, headers map[string]string) ([]byte, error) {
	return c.doAttempt(ctx, http.MethodGet, path, query, nil, headers, true, 0, false)
}

// Post performs a POST request with JSON body.
func (c *Client) Post(ctx context.Context, path string, body map[string]any) ([]byte, error) {
	return c.doAttempt(ctx, http.MethodPost, path, nil, body, nil, true, 0, false)
}

// PostH performs a POST request with JSON body and extra headers.
func (c *Client) PostH(ctx context.Context, path string, body map[string]any, headers map[string]string) ([]byte, error) {
	return c.doAttempt(ctx, http.MethodPost, path, nil, body, headers, true, 0, false)
}

// PostMultipartH uploads a file via multipart/form-data POST.
// The file is sent as the "file" field. Extra headers (e.g. x-on-behalf-of) can be passed.
func (c *Client) PostMultipartH(ctx context.Context, path string, filePath string, query map[string]string, headers map[string]string) ([]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	fileName := filepath.Base(filePath)
	mimeType := mime.TypeByExtension(filepath.Ext(filePath))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, fileName))
	h.Set("Content-Type", mimeType)
	part, err := w.CreatePart(h)
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(part, f); err != nil {
		return nil, err
	}
	w.Close()

	u := c.baseURL + path
	if len(query) > 0 {
		params := url.Values{}
		for k, v := range query {
			if v != "" {
				params.Set(k, v)
			}
		}
		if encoded := params.Encode(); encoded != "" {
			u += "?" + encoded
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("x-idempotency-key", uuid.New().String())
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	token, err := c.tokens.Get(ctx, c.cfg.Env, c.cfg.ClientID, c.cfg.APIKey, c.baseURL)
	if err == nil && token != "" {
		req.Header.Set("x-auth-token", "Bearer "+token)
	}

	if c.debug {
		fmt.Fprintf(c.debugOut, "\n[DEBUG] → POST %s (multipart/form-data: %s)\n", u, filepath.Base(filePath))
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, &apierr.NetworkError{Message: fmt.Sprintf("connection failed: %s", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &apierr.NetworkError{Message: fmt.Sprintf("failed to read response: %s", err)}
	}

	if c.debug {
		printDebugResponse(c.debugOut, resp.StatusCode, respBody)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}
	return nil, parseAPIError(resp.StatusCode, respBody)
}

func (c *Client) doAttempt(
	ctx context.Context,
	method, path string,
	query map[string]string,
	body map[string]any,
	extraHeaders map[string]string,
	withIdempotency bool,
	attempt int,
	tokenRefreshed bool,
) ([]byte, error) {
	// Build URL
	u := c.baseURL + path
	if len(query) > 0 {
		params := url.Values{}
		for k, v := range query {
			if v != "" {
				params.Set(k, v)
			}
		}
		if encoded := params.Encode(); encoded != "" {
			u += "?" + encoded
		}
	}

	// Build body
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if withIdempotency {
		req.Header.Set("x-idempotency-key", uuid.New().String())
	}
	for k, v := range extraHeaders {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	// Attach auth token (skipped when clientID/apiKey are empty)
	token, err := c.tokens.Get(ctx, c.cfg.Env, c.cfg.ClientID, c.cfg.APIKey, c.baseURL)
	if err == nil && token != "" {
		req.Header.Set("x-auth-token", "Bearer "+token)
	}

	if c.debug {
		printDebugRequest(c.debugOut, method, u, req.Header, body)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		if attempt < maxRetries-1 {
			time.Sleep(retryDelay(c.retryDelayOverride, attempt, 0))
			return c.doAttempt(ctx, method, path, query, body, extraHeaders, withIdempotency, attempt+1, tokenRefreshed)
		}
		return nil, &apierr.NetworkError{Message: fmt.Sprintf("connection failed: %s", err)}
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, &apierr.NetworkError{Message: fmt.Sprintf("failed to read response: %s", err)}
	}

	if c.debug {
		printDebugResponse(c.debugOut, resp.StatusCode, respBody)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	apiErr := parseAPIError(resp.StatusCode, respBody)

	// Token expiry: invalidate and retry once.
	// Check isTokenExpired on the raw body message to avoid fragility with UserMessage wrapping.
	if resp.StatusCode == 401 && !tokenRefreshed {
		var rawBody struct {
			Message string `json:"message"`
			Error   string `json:"error"`
		}
		json.Unmarshal(respBody, &rawBody)
		rawMsg := rawBody.Message
		if rawMsg == "" {
			rawMsg = rawBody.Error
		}
		if isTokenExpired(rawMsg) {
			c.tokens.Invalidate(c.cfg.Env)
			return c.doAttempt(ctx, method, path, query, body, extraHeaders, withIdempotency, attempt, true)
		}
	}

	// Retry on 429 / 5xx
	if shouldRetry(resp.StatusCode, attempt) {
		delay := retryDelay(c.retryDelayOverride, attempt, parseRetryAfter(resp.Header.Get("Retry-After")))
		time.Sleep(delay)
		return c.doAttempt(ctx, method, path, query, body, extraHeaders, withIdempotency, attempt+1, tokenRefreshed)
	}

	return nil, apiErr
}

func printDebugRequest(w io.Writer, method, u string, headers http.Header, body map[string]any) {
	fmt.Fprintf(w, "\n[DEBUG] → %s %s\n", method, u)
	keys := make([]string, 0, len(headers))
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.EqualFold(k, "X-Auth-Token") {
			fmt.Fprintf(w, "[DEBUG]   %s: [hidden]\n", k)
			continue
		}
		fmt.Fprintf(w, "[DEBUG]   %s: %s\n", k, strings.Join(headers[k], ", "))
	}
	if body != nil {
		b, _ := json.MarshalIndent(body, "", "  ")
		fmt.Fprintf(w, "[DEBUG]\n[DEBUG]   %s\n", strings.ReplaceAll(string(b), "\n", "\n[DEBUG]   "))
	}
}

func printDebugResponse(w io.Writer, statusCode int, respBody []byte) {
	fmt.Fprintf(w, "[DEBUG] ← %d %s\n", statusCode, http.StatusText(statusCode))
	if len(respBody) > 0 {
		var pretty bytes.Buffer
		if json.Indent(&pretty, respBody, "[DEBUG]   ", "  ") == nil {
			fmt.Fprintf(w, "[DEBUG]   %s\n\n", pretty.String())
		} else {
			fmt.Fprintf(w, "[DEBUG]   %s\n\n", string(respBody))
		}
	}
}
