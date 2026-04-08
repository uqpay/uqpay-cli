package payment

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newSettlementCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settlement",
		Short: "Query payment settlements",
	}
	cmd.AddCommand(
		newSettlementListCmd(),
	)
	return cmd
}

func newSettlementListCmd() *cobra.Command {
	var onBehalfOf, paymentIntentID, batchID, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment settlements",
		Long: `List payment settlements.

Flags:
  --payment-intent-id   Filter by payment intent ID
  --batch-id            Filter by settlement batch ID
  --start-time          Filter by settled time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time            Filter by settled time end (ISO 8601)
  --page-size           Results per page (default 10)
  --page-num            Page number (default 1)
  --on-behalf-of        Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetHI(context.Background(), "/v2/payment/settlements", map[string]string{
				"payment_intent_id":   paymentIntentID,
				"settlement_batch_id": batchID,
				"settled_start_time":  startTime,
				"settled_end_time":    endTime,
				"page_size":           pageSize,
				"page_number":         pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&paymentIntentID, "payment-intent-id", "", "Filter by payment intent ID")
	cmd.Flags().StringVar(&batchID, "batch-id", "", "Filter by settlement batch ID")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter settled start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter settled end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
