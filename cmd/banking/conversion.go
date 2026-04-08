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

const conversionQuoteHelp = `Create a currency conversion quote.

Parameters:
  Required:
    sell_currency      string   ISO 4217 currency to sell
    buy_currency       string   ISO 4217 currency to buy
    conversion_date    string   Scheduled conversion date (YYYY-MM-DD)

  Optional:
    sell_amount        string   Amount of sell_currency (provide either sell_amount or buy_amount)
    buy_amount         string   Amount of buy_currency (provide either buy_amount or sell_amount)
    transaction_type   string   conversion | payout

Examples:
  uqpay conversion quote \
    -d sell_currency=USD \
    -d buy_currency=SGD \
    -d sell_amount=1000 \
    -d conversion_date=2026-04-10`

const conversionCreateHelp = `Create a currency conversion from a quote.

Parameters:
  Required:
    quote_id          string   Quote ID obtained from "conversion quote"
    sell_currency     string   ISO 4217 currency to sell
    buy_currency      string   ISO 4217 currency to buy
    conversion_date   string   Scheduled conversion date (YYYY-MM-DD)

  Optional:
    sell_amount       string   Amount of sell_currency
    buy_amount        string   Amount of buy_currency

Examples:
  uqpay conversion create \
    -d quote_id=qte_xxx \
    -d sell_currency=USD \
    -d buy_currency=SGD \
    -d sell_amount=1000 \
    -d conversion_date=2026-04-10`

func NewConversionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "conversion",
		Short: "Manage currency conversions",
	}
	cmd.AddCommand(
		newConversionListCmd(),
		newConversionGetCmd(),
		newConversionQuoteCmd(),
		newConversionCreateCmd(),
		newConversionDatesCmd(),
	)
	return cmd
}

func newConversionListCmd() *cobra.Command {
	var onBehalfOf, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List conversions",
		Long: `List currency conversions.

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
			data, err := c.GetH(context.Background(), "/v1/conversion", map[string]string{
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

func newConversionGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <conversion-id>",
		Short: "Retrieve a conversion",
		Long: `Retrieve a currency conversion by its ID.

The conversion ID is returned in the response of "conversion create" or "conversion list".

Examples:
  uqpay conversion get cvn_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/conversion/"+args[0], nil,
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

func newConversionQuoteCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "quote",
		Short: "Create a conversion quote",
		Long:  conversionQuoteHelp,
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
			resp, err := c.PostH(context.Background(), "/v1/conversion/quote", body,
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

func newConversionCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a currency conversion",
		Long:  conversionCreateHelp,
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
			resp, err := c.PostH(context.Background(), "/v1/conversion", body,
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

func newConversionDatesCmd() *cobra.Command {
	var onBehalfOf, currencyFrom, currencyTo string
	cmd := &cobra.Command{
		Use:   "dates",
		Short: "List available conversion dates",
		Long: `List available dates on which a currency conversion can be scheduled.

Flags:
  --currency-from   ISO 4217 source currency (required)
  --currency-to     ISO 4217 target currency (required)
  --on-behalf-of    Sub-account ID to act on behalf of

Examples:
  uqpay conversion dates --currency-from USD --currency-to SGD`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/conversion/conversion_dates", map[string]string{
				"currency_from": currencyFrom,
				"currency_to":   currencyTo,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currencyFrom, "currency-from", "", "ISO 4217 source currency (required)")
	cmd.Flags().StringVar(&currencyTo, "currency-to", "", "ISO 4217 target currency (required)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
