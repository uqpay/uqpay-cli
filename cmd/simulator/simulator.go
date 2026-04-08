package simulator

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func NewSimulateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate transactions (sandbox only)",
	}
	cmd.AddCommand(
		newSimulateAuthorizationCmd(),
		newSimulateReversalCmd(),
		newSimulateDepositCmd(),
	)
	return cmd
}

func newSimulateAuthorizationCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "authorization",
		Short: "Simulate an issuing card authorization",
		Long: `Simulate an authorization transaction on an issuing card (sandbox only).

Only cards with BIN 40963608 are supported for simulation.
The card's available balance must exceed the simulated transaction amount.
ATM cash withdrawals cannot be simulated.

Parameters:
  Required:
    card_id                  string   Card UUID (from "card list")
    transaction_amount       number   Amount to authorize
    transaction_currency     string   ISO 4217 currency code (e.g. USD)
    merchant_name            string   Merchant name
    merchant_category_code   string   MCC code (only 5734 supported)

Examples:
  uqpay simulate authorization \
    -d card_id=c0cef051-29c5-4796-b86a-cd5b684bfad7 \
    -d transaction_amount=10.00 \
    -d transaction_currency=USD \
    -d "merchant_name=Orchard Central" \
    -d merchant_category_code=5734`,
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
			dotparam.CoerceNumbers(body, "transaction_amount")
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v1/simulation/issuing/authorization", body,
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

func newSimulateReversalCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "reversal <transaction-id>",
		Short: "Simulate a reversal of an approved authorization",
		Long: `Simulate the reversal of an existing approved authorization (sandbox only).

The transaction ID is returned in the response of "simulate authorization".

Examples:
  uqpay simulate reversal 5135e6cc-28b6-4889-81dc-3b86a09e1395`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			body := map[string]any{"transaction_id": args[0]}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v1/simulation/issuing/reversal", body,
				map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newSimulateDepositCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "Simulate a deposit into a banking account",
		Long: `Simulate an inbound deposit into a global banking account (sandbox only).

Parameters:
  Required:
    amount                    number   Amount to deposit
    currency                  string   ISO 4217 currency code (e.g. SGD, USD)
    sender_swift_code         string   Sender's bank SWIFT code (e.g. WELGBE22)

  Optional:
    receiver_account_number   string   Receiver's account number (IBAN or local)
    sender_account_number     string   Sender's account number
    sender_country            string   ISO 3166-1 alpha-2 sender country code
    sender_name               string   Sender's name

Examples:
  uqpay simulate deposit \
    -d amount=1000 \
    -d currency=SGD \
    -d sender_swift_code=WELGBE22 \
    -d receiver_account_number=SG123456789012345678

  uqpay simulate deposit \
    -d amount=500 \
    -d currency=USD \
    -d sender_swift_code=WELGBE22 \
    --on-behalf-of <sub-account-id>`,
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
			resp, err := c.PostH(context.Background(), "/v1/simulation/deposit", body,
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
