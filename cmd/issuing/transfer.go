package issuing

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

const issuingTransferCreateHelp = `Create a fund transfer between master account and sub-accounts.

Parameters:
  Required:
    source_account_id       string   Source account ID (UUID)
    destination_account_id  string   Destination account ID (UUID)
    currency                string   ISO 4217 currency code (e.g. SGD, USD)
    amount                  number   Transfer amount

  Optional:
    remark                  string   Transfer remark

Examples:
  uqpay issuing transfer create \
    -d source_account_id=acc_xxx \
    -d destination_account_id=acc_yyy \
    -d currency=SGD \
    -d amount=100 \
    -d remark="top-up"`

func newIssuingTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer",
		Short: "Manage issuing transfers between accounts",
	}
	cmd.AddCommand(
		newIssuingTransferCreateCmd(),
		newIssuingTransferGetCmd(),
	)
	return cmd
}

func newIssuingTransferCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an issuing transfer",
		Long:  issuingTransferCreateHelp,
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
			dotparam.CoerceNumbers(body, "amount")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/transfers", body)
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

func newIssuingTransferGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <transfer-id>",
		Short: "Retrieve an issuing transfer",
		Long: `Retrieve details of an issuing transfer by its transfer ID.

The transfer ID is returned in the response of "issuing transfer create".

Examples:
  uqpay issuing transfer get tf_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/transfers/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}
