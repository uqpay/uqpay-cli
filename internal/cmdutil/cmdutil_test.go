package cmdutil_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
)

func TestWriteErrorToSerializesReconcileRequired(t *testing.T) {
	err := &apierr.ReconcileRequiredError{
		Message:        "request outcome is unknown; reconcile remote state before retrying",
		Method:         "POST",
		Path:           "/v1/transfers",
		IdempotencyKey: "stable-key",
	}
	var output bytes.Buffer
	cmdutil.WriteErrorTo(&output, err, "json")

	var got map[string]any
	if json.Unmarshal(output.Bytes(), &got) != nil {
		t.Fatalf("invalid JSON output: %q", output.String())
	}
	if got["error"] != "reconcile_required" || got["method"] != "POST" ||
		got["path"] != "/v1/transfers" || got["idempotency_key"] != "stable-key" {
		t.Fatalf("unexpected reconcile output: %#v", got)
	}
}
