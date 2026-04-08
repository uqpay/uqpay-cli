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

func newVirtualAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "virtual-account",
		Short: "Manage banking virtual accounts",
	}
	cmd.AddCommand(
		newVirtualAccountListCmd(),
		newVirtualAccountCreateCmd(),
	)
	return cmd
}

func newVirtualAccountListCmd() *cobra.Command {
	var onBehalfOf, currency, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List virtual accounts",
		Long: `List virtual accounts.

Flags:
  --currency      Filter by currency (ISO 4217)
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
			data, err := c.GetH(context.Background(), "/v1/virtual/accounts", map[string]string{
				"currency":    currency,
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
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newVirtualAccountCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a virtual account",
		Long: `Create a new virtual account.

Parameters:
  Required:
    currency         string   ISO 4217 currency code(s), comma-separated (e.g. USD,SGD)

  Optional:
    payment_method   string   LOCAL | SWIFT

Examples:
  uqpay virtual-account create -d currency=SGD
  uqpay virtual-account create -d currency=USD,SGD -d payment_method=LOCAL`,
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
			resp, err := c.PostH(context.Background(), "/v1/virtual/accounts", body,
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
