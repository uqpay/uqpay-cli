package issuing

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newTransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "Query issuing transactions",
	}
	cmd.AddCommand(newTransactionListCmd(), newTransactionGetCmd())
	return cmd
}

func newTransactionListCmd() *cobra.Command {
	var cardID, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List transactions",
		Long: `List card transactions with optional filters.

Flags:
  --card-id      Filter by card ID
  --start-time   Start time in ISO 8601 format (e.g. 2026-01-01T00:00:00Z)
  --end-time     End time in ISO 8601 format (e.g. 2026-01-31T23:59:59Z)
  --page-size    Results per page, 10-100 (default 10)
  --page-num     Page number (default 1)

Examples:
  uqpay transaction list
  uqpay transaction list --card-id c03a7ac3-42f2-437d-8f5d-c328c28f1012
  uqpay transaction list --start-time 2026-01-01T00:00:00Z --end-time 2026-01-31T23:59:59Z`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/transactions", map[string]string{
				"card_id":     cardID,
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
	cmd.Flags().StringVar(&cardID, "card-id", "", "Filter by card ID")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "End time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newTransactionGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transaction-id>",
		Short: "Retrieve a transaction",
		Long: `Retrieve full details of a card transaction by its transaction ID.

Examples:
  uqpay transaction get txn_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/transactions/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}
