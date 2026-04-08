package banking

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newBankingBalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Query banking balances",
	}
	cmd.AddCommand(
		newBankingBalanceListCmd(),
		newBankingBalanceGetCmd(),
		newBankingBalanceTransactionsCmd(),
	)
	return cmd
}

func newBankingBalanceListCmd() *cobra.Command {
	var onBehalfOf, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List banking balances",
		Long: `List all banking account balances.

Flags:
  --page-size     Results per page (default 10)
  --page-num      Page number (default 1)
  --on-behalf-of  Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/balances", map[string]string{
				"page_size":   pageSize,
				"page_number": pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newBankingBalanceGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <currency>",
		Short: "Get banking balance for a specific currency",
		Long: `Get the banking balance for a specific currency.

The currency argument is the ISO 4217 currency code, e.g. USD, SGD, EUR.

Examples:
  uqpay banking balance get USD`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/balances/"+args[0], nil,
				map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newBankingBalanceTransactionsCmd() *cobra.Command {
	var onBehalfOf, currency, txnType, txnStatus, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "transactions",
		Short: "List banking balance transactions",
		Long: `List banking balance transaction history.

Flags:
  --currency        Filter by currency (ISO 4217)
  --type            Filter by transaction type: ALL | PAYIN | DEPOSIT | PAYOUT | TRANSFER | CONVERSION | FEE | REFUND | ADJUSTMENT | INVOICE
  --status          Filter by transaction status: ALL | COMPLETED | PENDING | FAILED
  --start-time      Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time        Filter by created time end (ISO 8601)
  --page-size       Results per page (default 10)
  --page-num        Page number (default 1)
  --on-behalf-of    Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/balances/transactions", map[string]string{
				"currency":           currency,
				"transaction_type":   txnType,
				"transaction_status": txnStatus,
				"start_time":         startTime,
				"end_time":           endTime,
				"page_size":          pageSize,
				"page_number":        pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217)")
	cmd.Flags().StringVar(&txnType, "type", "", "Filter by type: ALL | PAYIN | DEPOSIT | PAYOUT | TRANSFER | CONVERSION | FEE | REFUND | ADJUSTMENT | INVOICE")
	cmd.Flags().StringVar(&txnStatus, "status", "", "Filter by status: ALL | COMPLETED | PENDING | FAILED")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
