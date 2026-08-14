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
		"nickname", "-d country=SG", "-d currency=USD", "account_id", "direct_id", `"0" for main`,
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
	for _, outputField := range []string{"account_id", "direct_id"} {
		if !strings.Contains(list.Long, outputField) {
			t.Errorf("application list help missing required output field %q", outputField)
		}
	}
	retrieve, _, _ := application.Find([]string{"retrieve"})
	for _, outputField := range []string{"account_id", "direct_id", `"0" for`} {
		if !strings.Contains(retrieve.Long, outputField) {
			t.Errorf("application retrieve help missing output contract %q", outputField)
		}
	}
	for _, flag := range []string{"page-num", "page-size", "status", "country", "currency", "on-behalf-of"} {
		if list.Flags().Lookup(flag) == nil {
			t.Errorf("application list missing --%s", flag)
		}
	}
}
