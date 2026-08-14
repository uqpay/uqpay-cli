package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestOperationManifestCoversCanonicalAPICommands(t *testing.T) {
	root := NewRootCmd()
	manifest, err := buildOperationManifest(root)
	if err != nil {
		t.Fatal(err)
	}
	if manifest.Schema != operationManifestSchema {
		t.Fatalf("schema = %q", manifest.Schema)
	}
	if got, want := len(manifest.Operations), 100; got != want {
		t.Fatalf("operation count = %d, want %d", got, want)
	}

	seenIDs := map[string]bool{}
	seenCommands := map[string]bool{}
	for _, operation := range manifest.Operations {
		command := strings.Join(operation.Command, " ")
		if operation.OperationID == "" || operation.APIOperationID == "" || operation.Domain == "" || operation.Resource == "" || operation.Action == "" {
			t.Fatalf("incomplete operation identity for %s: %#v", command, operation)
		}
		if seenIDs[operation.OperationID] {
			t.Fatalf("duplicate operation_id %q", operation.OperationID)
		}
		if seenCommands[command] {
			t.Fatalf("duplicate command %q", command)
		}
		seenIDs[operation.OperationID] = true
		seenCommands[command] = true
		if !strings.HasPrefix(operation.HTTPPath, "/v") {
			t.Fatalf("invalid HTTP path for %s: %q", command, operation.HTTPPath)
		}
		if !strings.HasPrefix(operation.DocumentationURL, "https://developers.uqpay.com/") ||
			!strings.HasSuffix(operation.DocumentationURL, "/"+operation.APIOperationID) {
			t.Fatalf("invalid documentation URL for %s: %q", command, operation.DocumentationURL)
		}
	}
	if seenCommands["uqpay beneficiary list"] {
		t.Fatal("shortcut command must not be duplicated in the canonical manifest")
	}
}

func TestOperationManifestRepresentativeContracts(t *testing.T) {
	manifest, err := buildOperationManifest(NewRootCmd())
	if err != nil {
		t.Fatal(err)
	}
	byCommand := map[string]manifestOperation{}
	for _, operation := range manifest.Operations {
		byCommand[strings.Join(operation.Command, " ")] = operation
	}

	virtualAccount := byCommand["uqpay banking virtual-account create"]
	if virtualAccount.OperationID != "banking.create-virtual-account" ||
		virtualAccount.HTTPMethod != "POST" || virtualAccount.HTTPPath != "/v1/virtual/accounts" ||
		virtualAccount.Risk != riskWrite || !virtualAccount.SupportsOnBehalf ||
		!virtualAccount.Idempotency.Required || virtualAccount.Idempotency.Mode != "caller_required" {
		t.Fatalf("unexpected VA Create contract: %#v", virtualAccount)
	}
	assertBodyField(t, virtualAccount, "country", "string", true)
	assertBodyField(t, virtualAccount, "currency", "string", true)
	assertBodyField(t, virtualAccount, "nickname", "string", false)

	settlements := byCommand["uqpay payment settlement list"]
	if settlements.HTTPMethod != "GET" || settlements.Risk != riskRead ||
		!settlements.Idempotency.Required || settlements.Idempotency.Mode != "auto_generated" {
		t.Fatalf("unexpected settlement read contract: %#v", settlements)
	}

	beneficiaryDelete := byCommand["uqpay banking beneficiary delete"]
	if beneficiaryDelete.Risk != riskDestructive {
		t.Fatalf("beneficiary delete risk = %q", beneficiaryDelete.Risk)
	}

	upload := byCommand["uqpay file upload"]
	assertBodyField(t, upload, "file", "string", true)

	terminal := byCommand["uqpay payment terminal register"]
	if !terminal.Body.Supported || !terminal.Body.AdditionalProperties {
		t.Fatalf("terminal dynamic body contract = %#v", terminal.Body)
	}

	additionalDocuments := byCommand["uqpay account additional-documents"]
	assertParameterRequired(t, additionalDocuments, "business-code")
	assertParameterRequired(t, additionalDocuments, "country")
}

func TestOperationManifestCommandOutputsJSONAndSchema(t *testing.T) {
	root := NewRootCmd()
	var output bytes.Buffer
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs([]string{"operation-manifest"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var manifest operationManifest
	if err := json.Unmarshal(output.Bytes(), &manifest); err != nil {
		t.Fatalf("manifest is not JSON: %v\n%s", err, output.String())
	}
	if len(manifest.Operations) != 100 {
		t.Fatalf("operation count = %d", len(manifest.Operations))
	}
	for _, forbidden := range []string{"api-key", "client-id", "YOUR_API_KEY", "secret"} {
		if strings.Contains(strings.ToLower(output.String()), strings.ToLower(forbidden)) {
			t.Fatalf("manifest leaks credential/presentation field %q", forbidden)
		}
	}

	root = NewRootCmd()
	output.Reset()
	root.SetOut(&output)
	root.SetErr(&output)
	root.SetArgs([]string{"operation-manifest", "--schema"})
	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(output.Bytes(), &schema); err != nil {
		t.Fatalf("schema is not JSON: %v\n%s", err, output.String())
	}
	if schema["$id"] != "urn:uqpay:cli:operation-manifest:v1" {
		t.Fatalf("schema $id = %#v", schema["$id"])
	}
}

func assertBodyField(t *testing.T, operation manifestOperation, name, fieldType string, required bool) {
	t.Helper()
	for _, field := range operation.Body.Fields {
		if field.Name == name {
			if field.Type != fieldType || field.Required != required {
				t.Fatalf("%s body field %s = %#v", operation.OperationID, name, field)
			}
			return
		}
	}
	t.Fatalf("%s missing body field %s", operation.OperationID, name)
}

func assertParameterRequired(t *testing.T, operation manifestOperation, name string) {
	t.Helper()
	for _, parameter := range operation.Parameters {
		if parameter.Name == name {
			if !parameter.Required {
				t.Fatalf("%s parameter %s should be required", operation.OperationID, name)
			}
			return
		}
	}
	t.Fatalf("%s missing parameter %s", operation.OperationID, name)
}
