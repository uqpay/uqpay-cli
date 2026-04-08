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

const bankAccountCreateHelp = `Create a payment bank account.

Parameters:
  Required:
    currency           string   ISO 4217 currency code

  Optional:
    account_number     string   Bank account number
    bank_name          string   Name of the bank
    swift_code         string   SWIFT/BIC code
    bank_country_code  string   Two-letter country code (ISO 3166-1 alpha-2)
    bank_address       string   Bank address
    bank_code_type     string   aba | bank_code | sort_code | bsb_code | ifsc | cnaps_number
    bank_code_value    string   Routing code value
    bank_branch_code   string   Bank branch code

Examples:
  uqpay payment bank-account create \
    -d currency=USD \
    -d account_number=123456789 \
    -d bank_name="Chase Bank" \
    -d swift_code=CHASUS33`

const bankAccountUpdateHelp = `Update a payment bank account.

The bank account ID is returned in the response of "payment bank-account create" or "payment bank-account list".

Parameters (all optional — only include fields to update):
    account_number     string   Bank account number
    bank_name          string   Name of the bank
    swift_code         string   SWIFT/BIC code
    bank_country_code  string   Two-letter country code (ISO 3166-1 alpha-2)
    bank_address       string   Bank address
    bank_code_type     string   aba | bank_code | sort_code | bsb_code | ifsc | cnaps_number
    bank_code_value    string   Routing code value
    bank_branch_code   string   Bank branch code

Examples:
  uqpay payment bank-account update ba_xxx -d bank_name="New Bank Name"`

func newPaymentBankAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bank-account",
		Short: "Manage payment bank accounts",
	}
	cmd.AddCommand(
		newPaymentBankAccountListCmd(),
		newPaymentBankAccountGetCmd(),
		newPaymentBankAccountCreateCmd(),
		newPaymentBankAccountUpdateCmd(),
	)
	return cmd
}

func newPaymentBankAccountListCmd() *cobra.Command {
	var onBehalfOf, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment bank accounts",
		Long: `List payment bank accounts.

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
			data, err := c.GetHI(context.Background(), "/v2/payment/bankaccount", map[string]string{
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

func newPaymentBankAccountGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <bank-account-id>",
		Short: "Retrieve a payment bank account",
		Long: `Retrieve a payment bank account by its ID.

The bank account ID is returned in the response of "payment bank-account create" or "payment bank-account list".

Examples:
  uqpay payment bank-account get ba_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetHI(context.Background(), "/v2/payment/bankaccount/"+args[0], nil,
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

func newPaymentBankAccountCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment bank account",
		Long:  bankAccountCreateHelp,
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
			resp, err := c.PostH(context.Background(), "/v2/payment/bankaccount/create", body,
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

func newPaymentBankAccountUpdateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "update <bank-account-id>",
		Short: "Update a payment bank account",
		Long:  bankAccountUpdateHelp,
		Args:  cobra.ExactArgs(1),
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
			resp, err := c.PostH(context.Background(), "/v2/payment/bankaccount/"+args[0], body,
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
