package apierr

import "fmt"

// APIError represents a non-2xx response from the UQPAY API.
type APIError struct {
	ErrorType  string `json:"error"`
	APIType    string `json:"type,omitempty"`
	APICode    string `json:"api_code,omitempty"`
	Message    string `json:"message"`
	StatusCode int    `json:"code"`
}

func (e *APIError) Error() string {
	return e.Message
}

// NetworkError represents a connection failure or timeout.
type NetworkError struct {
	Message string
}

func (e *NetworkError) Error() string {
	return e.Message
}

// ReconcileRequiredError means a mutating request may have reached the API, but
// the CLI could not determine its outcome. Callers must reconcile remote state
// before deciding whether to submit another logical request.
type ReconcileRequiredError struct {
	Message        string
	Method         string
	Path           string
	IdempotencyKey string
}

func (e *ReconcileRequiredError) Error() string {
	return e.Message
}

// ConfigError represents a missing or invalid local configuration.
type ConfigError struct {
	Message string
}

func (e *ConfigError) Error() string {
	return e.Message
}

// ErrorTypeFromStatus returns a short error type string for a given HTTP status code.
func ErrorTypeFromStatus(status int) string {
	switch {
	case status == 401:
		return "authentication_failed"
	case status == 403:
		return "forbidden"
	case status == 404:
		return "not_found"
	case status == 409:
		return "conflict"
	case status == 429:
		return "rate_limited"
	case status >= 500:
		return "server_error"
	default:
		return "invalid_request"
	}
}

// UserMessage returns a human-readable message for well-known status codes,
// falling back to the API-provided message.
func UserMessage(status int, apiMessage string) string {
	switch status {
	case 401:
		return fmt.Sprintf("authentication failed — check your client-id and api-key (%s)", apiMessage)
	case 403:
		return "forbidden — your account lacks permission for this operation"
	default:
		return apiMessage
	}
}
