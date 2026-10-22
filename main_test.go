package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/apierr"
)

func TestReconcileRequiredHasDedicatedExitCode(t *testing.T) {
	err := &apierr.ReconcileRequiredError{Message: "unknown outcome"}
	if got := exitCodeFor(err); got != 5 {
		t.Fatalf("exitCodeFor(ReconcileRequiredError) = %d, want 5", got)
	}
}

func TestCLIUsageErrorsReachStderr(t *testing.T) {
	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"banking", "payout", "validate"}, "required flag(s)"},
		{[]string{"unknown-synthetic-command"}, "unknown command"},
		{[]string{"banking", "payout", "validate", "--unknown-synthetic-flag"}, "unknown flag"},
	} {
		child := exec.Command(os.Args[0], append([]string{"-test.run=TestCLIErrorHelper", "--"}, tc.args...)...)
		child.Env = append(os.Environ(), "UQPAY_CLI_ERROR_HELPER=1")
		var stdout, stderr bytes.Buffer
		child.Stdout = &stdout
		child.Stderr = &stderr
		err := child.Run()
		exit, ok := err.(*exec.ExitError)
		if !ok || exit.ExitCode() != 4 {
			t.Fatalf("%v exit: %v", tc.args, err)
		}
		if stdout.Len() != 0 || !strings.Contains(stderr.String(), tc.want) || strings.Count(stderr.String(), "Error:") != 1 {
			t.Fatalf("%v stdout=%q stderr=%q", tc.args, stdout.String(), stderr.String())
		}
	}
}

func TestCLIErrorHelper(t *testing.T) {
	if os.Getenv("UQPAY_CLI_ERROR_HELPER") != "1" {
		return
	}
	for i, arg := range os.Args {
		if arg == "--" {
			os.Args = append([]string{"uqpay"}, os.Args[i+1:]...)
			break
		}
	}
	main()
	os.Exit(0)
}

func TestHandlerErrorIsNotPrintedTwice(t *testing.T) {
	var stderr bytes.Buffer
	failure := &apierr.APIError{Message: "synthetic failure", StatusCode: 400}
	root := &cobra.Command{Use: "test", SilenceErrors: true, SilenceUsage: true, RunE: func(c *cobra.Command, args []string) error {
		fmt.Fprintln(c.ErrOrStderr(), "structured API error")
		return failure
	}}
	root.SetArgs([]string{})
	root.SetErr(&stderr)
	if execute(root) != failure || stderr.String() != "structured API error\n" {
		t.Fatalf("duplicate/lost error: %q", stderr.String())
	}
}
