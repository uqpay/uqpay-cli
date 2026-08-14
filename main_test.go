package main

import (
	"testing"

	"github.com/uqpay/uqpay-cli/internal/apierr"
)

func TestReconcileRequiredHasDedicatedExitCode(t *testing.T) {
	err := &apierr.ReconcileRequiredError{Message: "unknown outcome"}
	if got := exitCodeFor(err); got != 5 {
		t.Fatalf("exitCodeFor(ReconcileRequiredError) = %d, want 5", got)
	}
}
