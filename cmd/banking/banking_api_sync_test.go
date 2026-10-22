package banking

import (
	"bytes"
	"encoding/json"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
	"strings"
	"testing"
)

func TestBankingAPISyncCLIHelpAndJSON(t *testing.T) {
	cmd := newPayoutValidateCmd()
	for _, flag := range []string{"beneficiary-id", "currency", "amount", "fee-paid-by", "on-behalf-of"} {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("missing --%s", flag)
		}
	}
	if !strings.Contains(cmd.Long, "does not reserve funds") {
		t.Error("preflight limitation missing")
	}
	if newBeneficiaryListCmd().Flags().Lookup("payment-method") == nil {
		t.Error("CARD filter missing")
	}
	for _, flag := range []string{"entity-type", "on-behalf-of"} {
		if newBeneficiaryPaymentMethodsCmd().Flags().Lookup(flag) == nil {
			t.Errorf("method query missing --%s", flag)
		}
	}
	required := strings.Split(payoutCreateHelp, "Optional:")[0]
	for _, field := range []string{"payout_reference", "fee_paid_by", "payout_date"} {
		if strings.Contains(required, field) {
			t.Errorf("optional %s documented as required", field)
		}
	}
	if !strings.Contains(beneficiaryCreateHelp, "bank_details.card_expiry") {
		t.Error("CARD nesting missing")
	}
	body, err := dotparam.Parse([]string{"payment_method=CARD", "bank_details.card_expiry=2099-12", "bank_details.clearing_system=MASTERCARD SEND"})
	if err != nil {
		t.Fatal(err)
	}
	if body["bank_details"].(map[string]any)["card_expiry"] != "2099-12" {
		t.Fatal("CARD body lost expiry")
	}
	payload := map[string]any{"channel_code": "MASTERCARD", "sufficient_funds": false, "beneficiary_valid": true, "validation_field": []any{"account_number", map[string]any{"value": "card_expiry"}}}
	var buffer bytes.Buffer
	raw, _ := json.Marshal(payload)
	if err := output.Print(&buffer, raw, "json"); err != nil {
		t.Fatal(err)
	}
	var result map[string]any
	if err := json.Unmarshal(buffer.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result["sufficient_funds"] != false || len(result["validation_field"].([]any)) != 2 {
		t.Fatalf("JSON result: %s", buffer.String())
	}
}
