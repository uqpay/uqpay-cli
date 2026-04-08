package issuing

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Query issuing balances",
	}
	cmd.AddCommand(
		newBalanceListCmd(),
		newBalanceGetCmd(),
		newBalanceTransactionsCmd(),
	)
	return cmd
}

func newBalanceListCmd() *cobra.Command {
	var currency, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issuing balances",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/balances", map[string]string{
				"currency":    currency,
				"page_size":   pageSize,
				"page_number": pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217, e.g. USD | SGD)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newBalanceGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <currency>",
		Short: "Get balance for a specific currency",
		Long: `Get the issuing balance for a specific currency.

Examples:
  uqpay issuing balance get SGD
  uqpay issuing balance get USD`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			apiData, err := c.Post(context.Background(), "/v1/issuing/balances", map[string]any{"currency": args[0]})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, apiData, cfg.Output)
		},
	}
}

func newBalanceTransactionsCmd() *cobra.Command {
	var currency, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "List balance transaction history",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/balances/transactions", map[string]string{
				"currency":    currency,
				"start_time":  startTime,
				"end_time":    endTime,
				"page_size":   pageSize,
				"page_number": pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217, e.g. USD | SGD)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Start time (ISO 8601, e.g. 2026-01-01T00:00:00Z)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "End time (ISO 8601, e.g. 2026-01-31T23:59:59Z)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}
