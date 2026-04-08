package client

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/uqpay/uqpay-cli/internal/apierr"
)

const maxRetries = 3 // 1 original + 2 retries

type apiBody struct {
	Type    string `json:"type"`
	Code    any    `json:"code"` // string or number
	Message string `json:"message"`
	Error   string `json:"error"` // 401 auth middleware: {"error": "..."}
}

// parseAPIError converts a non-2xx response into *apierr.APIError.
// Always uses HTTP status as authoritative — never body.code.
func parseAPIError(status int, body []byte) *apierr.APIError {
	var b apiBody
	json.Unmarshal(body, &b) // ignore error; fall through to defaults

	msg := b.Message
	if msg == "" && b.Error != "" {
		msg = b.Error
	}
	if msg == "" {
		msg = fmt.Sprintf("HTTP %d", status)
	}

	return &apierr.APIError{
		ErrorType:  apierr.ErrorTypeFromStatus(status),
		Message:    apierr.UserMessage(status, msg),
		StatusCode: status,
	}
}

// isTokenExpired returns true if a 401 message indicates token expiry.
func isTokenExpired(msg string) bool {
	lower := strings.ToLower(msg)
	patterns := []string{"token has expired", "jwt expired", "login expired", "token expired"}
	for _, p := range patterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}

// shouldRetry returns true if the request should be retried.
func shouldRetry(status, attempt int) bool {
	if attempt >= maxRetries-1 {
		return false
	}
	return status == 429 || status >= 500
}

// retryDelay returns the delay duration for a given attempt.
func retryDelay(override time.Duration, attempt int, retryAfterMs int64) time.Duration {
	if override > 0 {
		return override
	}
	if retryAfterMs > 0 {
		return time.Duration(retryAfterMs) * time.Millisecond
	}
	base := int64(500) << attempt // 500ms * 2^attempt
	if base > 30_000 {
		base = 30_000
	}
	return time.Duration(base) * time.Millisecond
}

// parseRetryAfter parses Retry-After header value to milliseconds.
func parseRetryAfter(header string) int64 {
	if header == "" {
		return 0
	}
	var seconds int64
	fmt.Sscanf(header, "%d", &seconds)
	return seconds * 1000
}
