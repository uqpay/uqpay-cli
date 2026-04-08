package banking

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

const payoutCreateHelp = `Create a payout to a beneficiary.

Parameters:
  Required:
    currency          string   ISO 4217 source currency (e.g. USD)
    amount            string   Amount to send in source currency
    purpose_code      string   Purpose code (e.g. GOODS_SERVICES, SALARY, INVOICE, etc.)
    payout_reference  string   Bank reference shown in beneficiary's bank statement
    fee_paid_by       string   SHARED | OURS
    payout_date       string   Scheduled date (YYYY-MM-DD)

  Optional:
    quote_id          string   Pre-created quote ID (from "conversion quote")
    payout_currency   string   ISO 4217 currency the beneficiary receives
    payout_amount     number   Amount the beneficiary receives in payout_currency
    beneficiary_id    string   Existing beneficiary UUID
    beneficiary       object   Inline beneficiary (if no beneficiary_id)
    payer_id          string   Payer UUID (deprecated)

Examples:
  uqpay banking payout create \
    -d currency=USD \
    -d amount=1000 \
    -d purpose_code=GOODS_SERVICES \
    -d payout_reference="Invoice #001" \
    -d fee_paid_by=SHARED \
    -d payout_date=2026-04-10 \
    -d beneficiary_id=ben_xxx`

func NewPayoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "payout",
		Short: "Manage banking payouts",
	}
	cmd.AddCommand(
		newPayoutListCmd(),
		newPayoutGetCmd(),
		newPayoutCreateCmd(),
	)
	return cmd
}

func newPayoutListCmd() *cobra.Command {
	var onBehalfOf, status, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payouts",
		Long: `List payouts.

Flags:
  --status      Filter by status: READY_TO_SEND | PENDING | REJECTED | FAILED | COMPLETED
  --start-time  Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time    Filter by created time end (ISO 8601)
  --page-size   Results per page (default 10)
  --page-num    Page number (default 1)
  --on-behalf-of  Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/payouts", map[string]string{
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
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: READY_TO_SEND | PENDING | REJECTED | FAILED | COMPLETED")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPayoutGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <payout-id>",
		Short: "Retrieve a payout",
		Long: `Retrieve a payout by its ID.

The payout ID is returned in the response of "payout create" or "payout list".

Examples:
  uqpay payout get pay_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/payouts/"+args[0], nil,
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

func newPayoutCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payout",
		Long:  payoutCreateHelp,
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
			dotparam.CoerceNumbers(body, "payout_amount")
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v1/payouts", body,
				map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
