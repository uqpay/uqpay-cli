package cmdutil

import (
	"encoding/json"
	"errors"
	"fmt"
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
	var apiErr *apierr.APIError
	var netErr *apierr.NetworkError
	var cfgErr *apierr.ConfigError
	switch {
	case errors.As(err, &apiErr):
		if outputFmt == "json" {
			fmt.Fprintf(os.Stderr, "{\"error\":%q,\"message\":%q,\"code\":%d}\n", apiErr.ErrorType, apiErr.Message, apiErr.StatusCode)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", apiErr.Message)
		}
	case errors.As(err, &netErr):
		if outputFmt == "json" {
			fmt.Fprintf(os.Stderr, "{\"error\":\"network_error\",\"message\":%q,\"code\":0}\n", netErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", netErr.Message)
		}
	case errors.As(err, &cfgErr):
		if outputFmt == "json" {
			fmt.Fprintf(os.Stderr, "{\"error\":\"config_error\",\"message\":%q,\"code\":0}\n", cfgErr.Message)
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", cfgErr.Message)
		}
	default:
		if outputFmt == "json" {
			fmt.Fprintf(os.Stderr, "{\"error\":\"unknown\",\"message\":%q,\"code\":0}\n", err.Error())
		} else {
			fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
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
