package banking

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func NewExchangeRateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exchange-rate",
		Short: "Query banking exchange rates",
	}
	cmd.AddCommand(newExchangeRateListCmd())
	return cmd
}

func newExchangeRateListCmd() *cobra.Command {
	var onBehalfOf, currencyPairs string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List current exchange rates",
		Long: `List current exchange rates.

Flags:
  --currency-pairs   Comma-separated 6-letter currency pairs (e.g. USDSGD,EURUSD)
  --on-behalf-of     Sub-account ID to act on behalf of

Examples:
  uqpay exchange-rate list
  uqpay exchange-rate list --currency-pairs USDSGD,EURUSD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/exchange/rates", map[string]string{
				"currency_pairs": currencyPairs,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currencyPairs, "currency-pairs", "", "Comma-separated 6-letter currency pairs (e.g. USDSGD,EURUSD)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
