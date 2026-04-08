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

const refundCreateHelp = `Create a refund for a payment intent.

Parameters:
  Required:
    payment_intent_id   string   Payment intent ID to refund
    amount              string   Amount to refund
    reason              string   Reason for the refund

  Optional:
    payment_attempt_id  string   Specific payment attempt ID to refund
    metadata            object   Key-value metadata

Examples:
  uqpay refund create \
    -d payment_intent_id=pi_xxx \
    -d amount=50 \
    -d reason=requested_by_customer`

func newRefundCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refund",
		Short: "Manage payment refunds",
	}
	cmd.AddCommand(
		newRefundListCmd(),
		newRefundGetCmd(),
		newRefundCreateCmd(),
	)
	return cmd
}

func newRefundListCmd() *cobra.Command {
	var clientID, paymentIntentID, merchantOrderID, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List refunds",
		Long: `List payment refunds.

Flags:
  --payment-intent-id   Filter by payment intent ID
  --merchant-order-id   Filter by merchant order ID
  --start-time          Filter by created time start (ISO 8601)
  --end-time            Filter by created time end (ISO 8601)
  --page-size           Results per page (default 10)
  --page-num            Page number (default 1)
  --client-id           Client ID (defaults to configured client ID)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			if clientID == "" {
				clientID = cfg.ClientID
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v2/payment/refunds", map[string]string{
				"payment_intent_id": paymentIntentID,
				"merchant_order_id": merchantOrderID,
				"start_time":        startTime,
				"end_time":          endTime,
				"page_size":         pageSize,
				"page_number":       pageNum,
			}, map[string]string{"x-client-id": clientID})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&paymentIntentID, "payment-intent-id", "", "Filter by payment intent ID")
	cmd.Flags().StringVar(&merchantOrderID, "merchant-order-id", "", "Filter by merchant order ID")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	return cmd
}

func newRefundGetCmd() *cobra.Command {
	var clientID string
	cmd := &cobra.Command{
		Use:   "get <refund-id>",
		Short: "Retrieve a refund",
		Long: `Retrieve a refund by its ID.

The refund ID is returned in the response of "refund create" or "refund list".

Examples:
  uqpay refund get ref_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			if clientID == "" {
				clientID = cfg.ClientID
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v2/payment/refunds/"+args[0], nil,
				map[string]string{"x-client-id": clientID})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	return cmd
}

func newRefundCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a refund",
		Long:  refundCreateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			if clientID == "" {
				clientID = cfg.ClientID
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v2/payment/refunds", body,
				map[string]string{
					"x-client-id":    clientID,
					"x-on-behalf-of": onBehalfOf,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}
