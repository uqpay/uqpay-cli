package main

import (
	"errors"
	"os"

	"github.com/uqpay/uqpay-cli/cmd"
	"github.com/uqpay/uqpay-cli/internal/apierr"
)

func main() {
	root := cmd.NewRootCmd()
	if err := root.Execute(); err != nil {
		os.Exit(exitCodeFor(err))
	}
}

func exitCodeFor(err error) int {
	if err == nil {
		return 0
	}
	var apiErr *apierr.APIError
	if errors.As(err, &apiErr) {
		return 1
	}
	var netErr *apierr.NetworkError
	if errors.As(err, &netErr) {
		return 3
	}
	var cfgErr *apierr.ConfigError
	if errors.As(err, &cfgErr) {
		return 2
	}
	return 4
}
