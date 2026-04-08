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

const paymentIntentCreateHelp = `Create a payment intent.

Parameters:
  Required:
    amount              string   Payment amount
    currency            string   ISO 4217 currency code (e.g. USD, SGD)
    merchant_order_id   string   Merchant reference ID (unique in your system)
    description         string   Payment descriptor shown to customer (max 32 chars)
    return_url          string   URL to redirect customer after payment

  Optional:
    payment_method      object   Payment method details
    ip_address          string   Customer IPv4 or IPv6 address
    payment_orders      object   Purchase order details
    browser_info        object   Browser info for 3DS (required when payment_method.card is used)
    metadata            object   Key-value metadata (max 512 bytes JSON)

Examples:
  uqpay payment intent create \
    -d amount=100 \
    -d currency=SGD \
    -d merchant_order_id=ORDER-001 \
    -d description="Test Payment" \
    -d return_url=https://example.com/return`

const paymentIntentConfirmHelp = `Confirm a payment intent.

The payment intent ID is returned in the response of "payment intent create".

Parameters:
  Optional:
    payment_method   object   Payment method details to confirm with
    ip_address       string   Customer IPv4 or IPv6 address (required when payment_method.card is used)
    browser_info     object   Browser information for 3DS (required when payment_method.card is used)
    return_url       string   URL to redirect customer after payment

Examples:
  uqpay payment intent confirm pi_xxx \
    -d payment_method.card.number=4111111111111111 \
    -d payment_method.card.exp_month=12 \
    -d payment_method.card.exp_year=2028 \
    -d payment_method.card.cvc=123`

func newPaymentIntentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "intent",
		Short: "Manage payment intents",
	}
	cmd.AddCommand(
		newPaymentIntentListCmd(),
		newPaymentIntentGetCmd(),
		newPaymentIntentCreateCmd(),
		newPaymentIntentUpdateCmd(),
		newPaymentIntentConfirmCmd(),
		newPaymentIntentCaptureCmd(),
		newPaymentIntentCancelCmd(),
	)
	return cmd
}

func newPaymentIntentListCmd() *cobra.Command {
	var onBehalfOf, clientID, status, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment intents",
		Long: `List payment intents.

Flags:
  --status        Filter by status: REQUIRES_PAYMENT_METHOD | REQUIRES_CUSTOMER_ACTION | REQUIRES_CAPTURE | PENDING | SUCCEEDED | CANCELLED | FAILED
  --start-time    Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time      Filter by created time end (ISO 8601)
  --page-size     Results per page (default 10)
  --page-num      Page number (default 1)
  --client-id     Override client ID header
  --on-behalf-of  Sub-account ID to act on behalf of`,
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
			data, err := c.GetH(context.Background(), "/v2/payment_intents", map[string]string{
				"payment_intent_status": status,
				"start_time":            startTime,
				"end_time":              endTime,
				"page_size":             pageSize,
				"page_number":           pageNum,
			}, map[string]string{
				"x-on-behalf-of": onBehalfOf,
				"x-client-id":    clientID,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: REQUIRES_PAYMENT_METHOD | REQUIRES_CUSTOMER_ACTION | REQUIRES_CAPTURE | PENDING | SUCCEEDED | CANCELLED | FAILED")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPaymentIntentGetCmd() *cobra.Command {
	var clientID string
	cmd := &cobra.Command{
		Use:   "get <payment-intent-id>",
		Short: "Retrieve a payment intent",
		Long: `Retrieve a payment intent by its ID.

The payment intent ID is returned in the response of "payment intent create" or "payment intent list".

Examples:
  uqpay payment intent get pi_xxx`,
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
			data, err := c.GetH(context.Background(), "/v2/payment_intents/"+args[0], nil,
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

func newPaymentIntentCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a payment intent",
		Long:  paymentIntentCreateHelp,
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
			resp, err := c.PostH(context.Background(), "/v2/payment_intents/create", body,
				map[string]string{
					"x-on-behalf-of": onBehalfOf,
					"x-client-id":    clientID,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPaymentIntentConfirmCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "confirm <payment-intent-id>",
		Short: "Confirm a payment intent",
		Long:  paymentIntentConfirmHelp,
		Args:  cobra.ExactArgs(1),
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
			resp, err := c.PostH(context.Background(), "/v2/payment_intents/"+args[0]+"/confirm", body,
				map[string]string{
					"x-on-behalf-of": onBehalfOf,
					"x-client-id":    clientID,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPaymentIntentCaptureCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "capture <payment-intent-id>",
		Short: "Capture a payment intent",
		Long: `Capture a payment intent that is in REQUIRES_CAPTURE status.

The payment intent ID is returned in the response of "payment intent create".

Parameters:
  Optional:
    amount_to_capture   number   Amount to capture (if different from original amount)

Examples:
  uqpay payment intent capture pi_xxx
  uqpay payment intent capture pi_xxx -d amount_to_capture=50`,
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
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			dotparam.CoerceNumbers(body, "amount_to_capture")
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v2/payment_intents/"+args[0]+"/capture", body,
				map[string]string{
					"x-on-behalf-of": onBehalfOf,
					"x-client-id":    clientID,
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

func newPaymentIntentUpdateCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "update <payment-intent-id>",
		Short: "Update a payment intent",
		Long: `Update a payment intent.

The payment intent ID is returned in the response of "payment intent create".

Parameters (all optional — only include fields to update):
    amount              string   Payment amount
    currency            string   ISO 4217 currency code
    merchant_order_id   string   Merchant reference ID
    description         string   Payment descriptor (max 32 chars)
    return_url          string   Redirect URL after payment
    customer            object   Customer details
    customer_id         string   Customer ID
    payment_orders      object   Purchase order details
    metadata            object   Key-value metadata (max 512 bytes JSON)

Examples:
  uqpay payment intent update pi_xxx -d description="Updated description"`,
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
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v2/payment_intents/"+args[0], body,
				map[string]string{
					"x-on-behalf-of": onBehalfOf,
					"x-client-id":    clientID,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newPaymentIntentCancelCmd() *cobra.Command {
	var data []string
	var onBehalfOf, clientID string
	cmd := &cobra.Command{
		Use:   "cancel <payment-intent-id>",
		Short: "Cancel a payment intent",
		Long: `Cancel a payment intent.

The payment intent ID is returned in the response of "payment intent create".

Parameters:
  Required:
    cancellation_reason   string   duplicate | fraudulent | requested_by_customer | abandoned

Examples:
  uqpay payment intent cancel pi_xxx -d cancellation_reason=requested_by_customer`,
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
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v2/payment_intents/"+args[0]+"/cancel", body,
				map[string]string{
					"x-on-behalf-of": onBehalfOf,
					"x-client-id":    clientID,
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
