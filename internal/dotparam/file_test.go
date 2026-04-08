package dotparam_test

import (
	"strings"
	"testing"

	"github.com/uqpay/uqpay-cli/internal/dotparam"
)

func TestFileEncoding(t *testing.T) {
	// Default @filepath: pure base64 (no data URI prefix)
	result, err := dotparam.Parse([]string{"front_file=@/tmp/id_front.png"})
	if err != nil {
		t.Fatal(err)
	}
	v, ok := result["front_file"].(string)
	if !ok {
		t.Fatal("expected string")
	}
	if strings.HasPrefix(v, "data:") {
		t.Errorf("expected pure base64, got data URI: %s...", v[:50])
	}
	t.Logf("pure base64: %s...", v[:40])

	// @+filepath: data URI format
	result2, err := dotparam.Parse([]string{"identity_docs[]=@+/tmp/id_front.png"})
	if err != nil {
		t.Fatal(err)
	}
	v2, ok := result2["identity_docs"].([]any)[0].(string)
	if !ok {
		t.Fatal("expected string in array")
	}
	if !strings.HasPrefix(v2, "data:image/png;base64,") {
		t.Errorf("expected data URI, got: %s...", v2[:50])
	}
	t.Logf("data URI: %s...", v2[:50])
}
