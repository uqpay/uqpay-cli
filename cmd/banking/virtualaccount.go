package banking

import (
	"context"
	"fmt"
	"net/url"
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
		newVirtualAccountApplicationCmd(),
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
	var onBehalfOf, idempotencyKey string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Submit a Virtual Account application",
		Long: `Submit a Virtual Account application.

Parameters:
  Required:
    country          string   ISO 3166-1 alpha-2 country code
    currency         string   One ISO 4217 currency code
    --idempotency-key         Stable x-idempotency-key value (1-64 characters)

  Optional:
    payment_method   string   LOCAL | SWIFT | omitted/empty to evaluate both
    nickname         string   Application label (maximum 255 characters)
    --on-behalf-of            Sub-account ID sent as x-on-behalf-of

Examples:
  uqpay banking virtual-account create --idempotency-key va-001 -d country=SG -d currency=USD
  uqpay banking virtual-account create --idempotency-key va-002 -d country=BH -d currency=GBP -d payment_method=SWIFT -d nickname=Collections`,
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
			if len(idempotencyKey) == 0 || len(idempotencyKey) > 64 {
				err := fmt.Errorf("--idempotency-key must contain between 1 and 64 characters")
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v1/virtual/accounts", body,
				map[string]string{
					"x-idempotency-key": idempotencyKey,
					"x-on-behalf-of":    onBehalfOf,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	cmd.Flags().StringVar(&idempotencyKey, "idempotency-key", "", "Stable x-idempotency-key value (required, maximum 64 characters)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID sent as x-on-behalf-of")
	_ = cmd.MarkFlagRequired("idempotency-key")
	return cmd
}

func newVirtualAccountApplicationCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application",
		Short: "Track Virtual Account applications (separate from issued accounts)",
	}
	cmd.AddCommand(
		newVirtualAccountApplicationListCmd(),
		newVirtualAccountApplicationRetrieveCmd(),
	)
	return cmd
}

func newVirtualAccountApplicationListCmd() *cobra.Command {
	var pageNumber, pageSize, status, country, currency, onBehalfOf string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Virtual Account application summaries",
		Long: `List Virtual Account applications, newest first.

This is not the issued Virtual Account list. Page number and page size are sent
on every request. Optional status, country, and currency filters are combined.
Status: SUBMITTED | PARTIALLY_COMPLETED | COMPLETED | FAILED | CLOSED.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/virtual/applications", map[string]string{
				"page_number": pageNumber,
				"page_size":   pageSize,
				"status":      status,
				"country":     country,
				"currency":    currency,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&pageNumber, "page-num", "1", "Page number (minimum 1)")
	cmd.Flags().StringVar(&pageSize, "page-size", "50", "Applications per page (1-100)")
	cmd.Flags().StringVar(&status, "status", "", "Filter by application status")
	cmd.Flags().StringVar(&country, "country", "", "Filter by ISO-2 country")
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by ISO-3 currency")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newVirtualAccountApplicationRetrieveCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:     "retrieve <application-id>",
		Aliases: []string{"get"},
		Short:   "Retrieve the complete current application",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(
				context.Background(),
				"/v1/virtual/applications/"+url.PathEscape(args[0]),
				nil,
				map[string]string{"x-on-behalf-of": onBehalfOf},
			)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID used for the application")
	return cmd
}
