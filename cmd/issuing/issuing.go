package issuing

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issuing",
		Short: "Issuing API — cards, cardholders, transactions, balances",
	}
	cmd.AddCommand(
		NewCardCmd(),
		NewCardholderCmd(),
		newTransactionCmd(),
		newBalanceCmd(),
		newIssuingTransferCmd(),
		newProductCmd(),
		newReportCmd(),
	)
	return cmd
}
