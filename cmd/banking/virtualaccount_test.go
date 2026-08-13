package banking

import (
	"strings"
	"testing"
)

func TestVirtualAccountCreateDocumentsApplicationContract(t *testing.T) {
	cmd := newVirtualAccountCreateCmd()
	help := cmd.Long
	for _, required := range []string{
		"country", "One ISO 4217 currency", "--idempotency-key", "LOCAL | SWIFT",
		"nickname", "-d country=SG", "-d currency=USD",
	} {
		if !strings.Contains(help, required) {
			t.Errorf("create help missing %q", required)
		}
	}
	if flag := cmd.Flags().Lookup("idempotency-key"); flag == nil {
		t.Fatal("create command does not expose --idempotency-key")
	}
}

func TestVirtualAccountApplicationCommandsRemainSeparateFromIssuedList(t *testing.T) {
	root := newVirtualAccountCmd()
	issuedList, _, err := root.Find([]string{"list"})
	if err != nil || issuedList == nil || issuedList.Use != "list" {
		t.Fatalf("issued Virtual Account list missing: %v", err)
	}
	application, _, err := root.Find([]string{"application"})
	if err != nil || application == nil {
		t.Fatalf("application command missing: %v", err)
	}
	for _, name := range []string{"list", "retrieve"} {
		child, _, findErr := application.Find([]string{name})
		if findErr != nil || child == nil || !strings.HasPrefix(child.Use, name) {
			t.Errorf("application %s command missing: %v", name, findErr)
		}
	}
	list, _, _ := application.Find([]string{"list"})
	for _, flag := range []string{"page-num", "page-size", "status", "country", "currency", "on-behalf-of"} {
		if list.Flags().Lookup(flag) == nil {
			t.Errorf("application list missing --%s", flag)
		}
	}
}
