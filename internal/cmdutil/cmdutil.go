package cmdutil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/config"
)

// Flag values bound by root command's PersistentFlags.
var (
	FlagEnv      string
	FlagClientID string
	FlagAPIKey   string
	FlagOutput   string
	FlagDebug    bool
)

func LoadConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, &apierr.ConfigError{Message: fmt.Sprintf("failed to load config: %s", err)}
	}
	cfg.ApplyEnvVars()
	if FlagEnv != "" {
		cfg.Env = FlagEnv
	}
	if FlagClientID != "" {
		cfg.ClientID = FlagClientID
	}
	if FlagAPIKey != "" {
		cfg.APIKey = FlagAPIKey
	}
	if FlagOutput != "" {
		cfg.Output = FlagOutput
	}
	cfg.Debug = FlagDebug
	return cfg, nil
}

func WriteError(err error, outputFmt string) {
	WriteErrorTo(os.Stderr, err, outputFmt)
}

// WriteErrorTo writes a stable human- or machine-readable representation of an error.
func WriteErrorTo(w io.Writer, err error, outputFmt string) {
	var apiErr *apierr.APIError
	var netErr *apierr.NetworkError
	var reconcileErr *apierr.ReconcileRequiredError
	var cfgErr *apierr.ConfigError
	switch {
	case errors.As(err, &apiErr):
		if outputFmt == "json" {
			if apiErr.APIType != "" || apiErr.APICode != "" {
				fmt.Fprintf(w, "{\"type\":%q,\"code\":%q,\"message\":%q}\n", apiErr.APIType, apiErr.APICode, apiErr.Message)
			} else {
				fmt.Fprintf(w, "{\"error\":%q,\"message\":%q,\"code\":%d}\n", apiErr.ErrorType, apiErr.Message, apiErr.StatusCode)
			}
		} else {
			fmt.Fprintf(w, "Error: %s\n", apiErr.Message)
		}
	case errors.As(err, &reconcileErr):
		if outputFmt == "json" {
			payload := struct {
				Error          string `json:"error"`
				Message        string `json:"message"`
				Code           int    `json:"code"`
				Method         string `json:"method"`
				Path           string `json:"path"`
				IdempotencyKey string `json:"idempotency_key"`
			}{
				Error:          "reconcile_required",
				Message:        reconcileErr.Message,
				Method:         reconcileErr.Method,
				Path:           reconcileErr.Path,
				IdempotencyKey: reconcileErr.IdempotencyKey,
			}
			_ = json.NewEncoder(w).Encode(payload)
		} else {
			fmt.Fprintf(w, "Error: %s (method=%s path=%s idempotency_key=%s)\n",
				reconcileErr.Message, reconcileErr.Method, reconcileErr.Path, reconcileErr.IdempotencyKey)
		}
	case errors.As(err, &netErr):
		if outputFmt == "json" {
			fmt.Fprintf(w, "{\"error\":\"network_error\",\"message\":%q,\"code\":0}\n", netErr.Message)
		} else {
			fmt.Fprintf(w, "Error: %s\n", netErr.Message)
		}
	case errors.As(err, &cfgErr):
		if outputFmt == "json" {
			fmt.Fprintf(w, "{\"error\":\"config_error\",\"message\":%q,\"code\":0}\n", cfgErr.Message)
		} else {
			fmt.Fprintf(w, "Error: %s\n", cfgErr.Message)
		}
	default:
		if outputFmt == "json" {
			fmt.Fprintf(w, "{\"error\":\"unknown\",\"message\":%q,\"code\":0}\n", err.Error())
		} else {
			fmt.Fprintf(w, "Error: %s\n", err.Error())
		}
	}
}

func ParseJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}
	return nil
}

func MarshalJSON(v any) ([]byte, error) {
	return json.Marshal(v)
}
