package payment

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payment",
		Short: "Payment API — intents, attempts, refunds, settlements",
	}
	cmd.AddCommand(
		newPaymentIntentCmd(),
		newPaymentAttemptCmd(),
		newPaymentBankAccountCmd(),
		newPaymentPayoutCmd(),
		newRefundCmd(),
		newPaymentBalanceCmd(),
		newSettlementCmd(),
	)
	return cmd
}
