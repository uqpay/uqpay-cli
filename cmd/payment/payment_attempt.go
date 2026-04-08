package payment

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newPaymentAttemptCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attempt",
		Short: "Query payment attempts",
	}
	cmd.AddCommand(
		newPaymentAttemptListCmd(),
		newPaymentAttemptGetCmd(),
	)
	return cmd
}

func newPaymentAttemptListCmd() *cobra.Command {
	var clientID, paymentIntentID, status, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List payment attempts",
		Long: `List payment attempts.

Flags:
  --payment-intent-id   Filter by payment intent ID
  --status              Filter by attempt status: INITIATED | AUTHENTICATION_REDIRECTED | PENDING_AUTHORIZATION | AUTHORIZED | CAPTURE_REQUESTED | SETTLED | SUCCEEDED | CANCELLED | EXPIRED | FAILED
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
			data, err := c.GetH(context.Background(), "/v2/payment/payment_attempts", map[string]string{
				"payment_intent_id": paymentIntentID,
				"attempt_status":    status,
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
	cmd.Flags().StringVar(&status, "status", "", "Filter by attempt status: INITIATED | AUTHENTICATION_REDIRECTED | PENDING_AUTHORIZATION | AUTHORIZED | CAPTURE_REQUESTED | SETTLED | SUCCEEDED | CANCELLED | EXPIRED | FAILED")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&clientID, "client-id", "", "Client ID (defaults to configured client ID)")
	return cmd
}

func newPaymentAttemptGetCmd() *cobra.Command {
	var clientID string
	cmd := &cobra.Command{
		Use:   "get <attempt-id>",
		Short: "Retrieve a payment attempt",
		Long: `Retrieve a payment attempt by its ID.

The attempt ID is returned in the response of "payment attempt list".

Examples:
  uqpay payment attempt get pa_xxx`,
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
			data, err := c.GetH(context.Background(), "/v2/payment/payment_attempts/"+args[0], nil,
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
