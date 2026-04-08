package payment

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

const paymentPayoutCreateHelp = `Create a payment payout.

Parameters:
  Required:
    payout_currency       string   ISO 4217 payout currency
    payout_amount         string   Payout amount
    statement_descriptor  string   Statement descriptor for the payout

  Optional:
    internal_note         string   Internal note
    payout_account_id     string   Target bank account ID

Examples:
  uqpay payment payout create \
    -d payout_currency=USD \
    -d payout_amount=500 \
    -d statement_descriptor="Monthly payout"`

func newPaymentPayoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payout",
		Short: "Manage payment payouts",
	}
	cmd.AddCommand(
		newPaymentPayoutListCmd(),
		newPaymentPayoutGetCmd(),
		newPaymentPayoutCreateCmd(),
	)
	return cmd
}

func newPaymentPayoutListCmd() *cobra.Command {
	var onBehalfOf, status, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment payouts",
		Long: `List payment payouts.

Flags:
  --status        Filter by status: INITIATED | PROCESSING | COMPLETED | FAILED | FAILED_REFUNDED
  --start-time    Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time      Filter by created time end (ISO 8601)
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
			data, err := c.GetHI(context.Background(), "/v2/payment/payout", map[string]string{
				"payout_status": status,
				"start_time":    startTime,
				"end_time":      endTime,
				"page_size":     pageSize,
				"page_number":   pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: INITIATED | PROCESSING | COMPLETED | FAILED | FAILED_REFUNDED")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPaymentPayoutGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <payout-id>",
		Short: "Retrieve a payment payout",
		Long: `Retrieve a payment payout by its ID.

The payout ID is returned in the response of "payment payout create" or "payment payout list".

Examples:
  uqpay payment payout get po_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetHI(context.Background(), "/v2/payment/payout/"+args[0], nil,
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

func newPaymentPayoutCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment payout",
		Long:  paymentPayoutCreateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v2/payment/payout/create", body,
				map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
