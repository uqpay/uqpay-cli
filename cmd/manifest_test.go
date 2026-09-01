package cmd

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestCommandManifestContainsRunnableContract(t *testing.T) {
	root := NewRootCmd()
	buffer := &bytes.Buffer{}
	root.SetOut(buffer)
	root.SetErr(buffer)
	root.SetArgs([]string{"command-manifest"})

	if _, err := root.ExecuteC(); err != nil {
		t.Fatalf("execute command-manifest: %v", err)
	}
	var payload struct {
		SchemaVersion int                    `json:"schema_version"`
		Commands      []commandManifestEntry `json:"commands"`
	}
	if err := json.Unmarshal(buffer.Bytes(), &payload); err != nil {
		t.Fatalf("command-manifest output is not valid JSON: %v\n%s", err, buffer.String())
	}
	if payload.SchemaVersion != 1 {
		t.Fatalf("schema_version = %d, want 1", payload.SchemaVersion)
	}

	found := false
	for _, entry := range payload.Commands {
		if len(entry.Path) == 3 && entry.Path[0] == "payment" && entry.Path[1] == "intent" && entry.Path[2] == "create" {
			found = true
			if !entry.Runnable || entry.MinArgs != 0 || entry.MaxArgs == nil || *entry.MaxArgs != 0 {
				t.Fatalf("unexpected command contract: %+v", entry)
			}
			flagFound := false
			for _, flag := range entry.Flags {
				if flag.Name == "data" && flag.Shorthand == "d" {
					flagFound = true
				}
				if flag.Name == "api-key" || flag.Name == "client-id" {
					t.Fatalf("credential flag leaked into manifest: %+v", flag)
				}
			}
			if !flagFound {
				t.Fatal("payment intent create data flag missing")
			}
		}
	}
	if !found {
		t.Fatal("payment intent create missing from manifest")
	}
}

func TestManifestCommandIsHiddenFromItsOwnOutput(t *testing.T) {
	root := NewRootCmd()
	for _, entry := range collectCommandManifest(root) {
		if len(entry.Path) > 0 && entry.Path[0] == "command-manifest" {
			t.Fatal("manifest command must not expose itself")
		}
	}
}
