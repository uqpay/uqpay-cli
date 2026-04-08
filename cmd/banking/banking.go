package banking

import "github.com/spf13/cobra"

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "banking",
		Short: "Banking API — balances, transfers, payouts, beneficiaries, conversions",
	}
	cmd.AddCommand(
		newBankingBalanceCmd(),
		newBankingTransferCmd(),
		NewBeneficiaryCmd(),
		NewPayoutCmd(),
		NewConversionCmd(),
		newDepositCmd(),
		NewExchangeRateCmd(),
		newVirtualAccountCmd(),
	)
	return cmd
}
