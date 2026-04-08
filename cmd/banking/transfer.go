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

const bankingTransferCreateHelp = `Create a banking transfer between a master account and a sub-account.

Use account UUIDs (from "account list"), not balance IDs.

Parameters:
  Required:
    source_account_id   string   Source account UUID (from "account list")
    target_account_id   string   Target account UUID (from "account list")
    currency            string   ISO 4217 currency code (e.g. USD, SGD)
    amount              string   Transfer amount
    reason              string   Reason for the transfer

Examples:
  uqpay banking transfer create \
    -d source_account_id=e95e0692-22b3-41b5-9dba-8ffef502d97a \
    -d target_account_id=f07d1878-523a-4267-aa7d-a2286ae836c6 \
    -d currency=SGD \
    -d amount=100 \
    -d reason="Internal settlement"`

func newBankingTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Manage banking transfers between accounts",
	}
	cmd.AddCommand(
		newBankingTransferListCmd(),
		newBankingTransferGetCmd(),
		newBankingTransferCreateCmd(),
	)
	return cmd
}

func newBankingTransferListCmd() *cobra.Command {
	var status, currency, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List banking transfers",
		Long: `List banking transfers.

Flags:
  --status      Filter by status: completed | failed
  --currency    Filter by currency (ISO 4217)
  --start-time  Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time    Filter by created time end (ISO 8601)
  --page-size   Results per page (default 10)
  --page-num    Page number (default 1)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/transfer", map[string]string{
				"transfer_status": status,
				"currency":        currency,
				"start_time":      startTime,
				"end_time":        endTime,
				"page_size":       pageSize,
				"page_number":     pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: completed | failed")
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newBankingTransferGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transfer-id>",
		Short: "Retrieve a banking transfer",
		Long: `Retrieve a banking transfer by its ID.

The transfer ID is returned in the response of "banking transfer create" or "banking transfer list".

Examples:
  uqpay banking transfer get txf_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/transfer/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newBankingTransferCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a banking transfer",
		Long:  bankingTransferCreateHelp,
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
			resp, err := c.Post(context.Background(), "/v1/transfer", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable)")
	return cmd
}
