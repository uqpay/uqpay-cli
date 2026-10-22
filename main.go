package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/cmd"
	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/update"
)

func main() {
	update.CheckForUpdate()
	root := cmd.NewRootCmd()
	if err := execute(root); err != nil {
		os.Exit(exitCodeFor(err))
	}
}

// Cobra rejects unknown commands, arguments and required flags before RunE.
// Command handlers already render their errors; only report pre-handler failures here.
func execute(root *cobra.Command) error {
	started := false
	var track func(*cobra.Command)
	track = func(command *cobra.Command) {
		if run := command.RunE; run != nil {
			command.RunE = func(c *cobra.Command, args []string) error { started = true; return run(c, args) }
		}
		if run := command.Run; run != nil {
			command.Run = func(c *cobra.Command, args []string) { started = true; run(c, args) }
		}
		for _, child := range command.Commands() {
			track(child)
		}
	}
	track(root)
	err := root.Execute()
	if err != nil && !started {
		fmt.Fprintln(root.ErrOrStderr(), "Error:", err)
	}
	return err
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
	var reconcileErr *apierr.ReconcileRequiredError
	if errors.As(err, &reconcileErr) {
		return 5
	}
	var cfgErr *apierr.ConfigError
	if errors.As(err, &cfgErr) {
		return 2
	}
	return 4
}
