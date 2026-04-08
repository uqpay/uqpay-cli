package issuing

import (
	"context"
	"net/url"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

// ── Help text constants ───────────────────────────────────────────────────────

const cardCreateHelp = `Create a new issuing card.

Parameters:
  Required:
    card_currency                                           string   SGD | USD
    cardholder_id                                          string   Cardholder UUID
    card_product_id                                        string   Card product ID

  Conditional:
    card_limit                                             number   Required for BINs 527735/555071/555243 (>=0.01); optional for others
    auto_cancel_trigger                                    string   ON_AUTH | ON_CAPTURE — required when usage_type=ONE_TIME

  Optional:
    usage_type                                             string   NORMAL (default) | ONE_TIME
    expiry_at                                              string   ISO 8601 datetime
    spending_controls[n].amount                            number   Max spend per interval
    spending_controls[n].interval                          string   PER_TRANSACTION
    risk_controls.allow_3ds_transactions                   string   Y (default) | N
    risk_controls.allowed_mcc[n]                           string   Allowed MCC whitelist (mutually exclusive with blocked_mcc)
    risk_controls.blocked_mcc[n]                           string   Blocked MCC blacklist (mutually exclusive with allowed_mcc)
    metadata.<key>                                         string   Custom key-value pairs (max 3200 bytes total)
    cardholder_required_fields.gender                      string   MALE | FEMALE
    cardholder_required_fields.nationality                 string   ISO 3166-1 alpha-2
    cardholder_required_fields.phone_number                string
    cardholder_required_fields.date_of_birth               string   YYYY-MM-DD
    cardholder_required_fields.residential_address.country string   ISO 3166-1 alpha-2 (required if address provided)
    cardholder_required_fields.residential_address.city    string   max 128 chars (required if address provided)
    cardholder_required_fields.residential_address.line1   string   max 255 chars (required if address provided)
    cardholder_required_fields.residential_address.state   string   max 128 chars
    cardholder_required_fields.residential_address.line2   string   max 255 chars
    cardholder_required_fields.residential_address.postal_code string max 16 chars
    cardholder_required_fields.identity.type               string   ID_CARD | PASSPORT (required if identity provided)
    cardholder_required_fields.identity.number             string   (required if identity provided)
    cardholder_required_fields.identity.front_file         string   base64 string or @filepath (required if identity provided)
    cardholder_required_fields.identity.back_file          string   base64 string or @filepath (required if type=ID_CARD)
    cardholder_required_fields.identity.hand_file          string   base64 string or @filepath, holding document
    cardholder_required_fields.kyc_verification.method     string   THIRD_PARTY | SUMSUB_REDIRECT (required if kyc_verification provided)
    cardholder_required_fields.kyc_verification.kyc_proof.provider     string   e.g. SUMSUB (required if method=THIRD_PARTY)
    cardholder_required_fields.kyc_verification.kyc_proof.reference_id string   >=10 chars, globally unique

Examples:
  uqpay issuing card create -d card_currency=USD -d cardholder_id=ch_xxx -d card_product_id=prod_xxx

  uqpay issuing card create \
    -d card_currency=USD \
    -d cardholder_id=ch_xxx \
    -d card_product_id=prod_xxx \
    -d spending_controls[0].amount=1000 \
    -d spending_controls[0].interval=PER_TRANSACTION \
    -d risk_controls.allow_3ds_transactions=Y`

const cardUpdateHelp = `Update an existing issuing card.

Parameters (all optional):
    card_limit                              number   Credit limit (omit if card mode_type is SINGLE)
    no_pin_payment_amount                   number   Max amount for PIN-less transactions (default 200 SGD)
    spending_controls[n].amount             number   Max spend per interval
    spending_controls[n].interval           string   PER_TRANSACTION
    risk_controls.allow_3ds_transactions    string   Y | N
    risk_controls.allowed_mcc[n]            string   Allowed MCC whitelist
    risk_controls.blocked_mcc[n]            string   Blocked MCC blacklist
    metadata.<key>                          string   Custom key-value pairs (max 3200 bytes total)

Examples:
  uqpay issuing card update card_xxx -d spending_controls[0].amount=500 -d spending_controls[0].interval=PER_TRANSACTION`

const cardUpdateStatusHelp = `Update the status of a card.

Parameters:
  Required:
    card_status     string   ACTIVE | FROZEN | CANCELLED

  Optional:
    update_reason   string   Reason for the status change (max 100 chars)

Examples:
  uqpay issuing card update-status card_xxx -d card_status=FROZEN
  uqpay issuing card update-status card_xxx -d card_status=ACTIVE -d update_reason="Cardholder request"`

const cardActivateHelp = `Activate a physical card.

Parameters:
  Required:
    card_id           string   Card ID
    activation_code   string   Activation code printed on the card mailer
    pin               string   PIN to set on the card — must be a 6-digit numeric value

Examples:
  uqpay issuing card activate -d card_id=card_xxx -d activation_code=123456 -d pin=123456`

const cardSetPinHelp = `Set or reset the PIN of a card.

Parameters:
  Required:
    card_id   string   Card ID
    pin       string   New PIN — must be a 6-digit numeric value

Examples:
  uqpay issuing card set-pin -d card_id=card_xxx -d pin=123456`

const cardAssignHelp = `Assign an unassigned physical card to a cardholder.

Parameters:
  Required:
    cardholder_id   string   Cardholder ID
    card_number     string   Full card number (PAN) printed on the card
    card_currency   string   ISO 4217 currency code (e.g. SGD, USD)
    card_mode       string   SINGLE | SHARE

Examples:
  uqpay issuing card assign \
    -d cardholder_id=ch_xxx \
    -d card_number=5550710000000001 \
    -d card_currency=SGD \
    -d card_mode=SINGLE`

// ── Command constructors ──────────────────────────────────────────────────────

func NewCardCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "card",
		Short: "Manage issuing cards",
	}
	cmd.AddCommand(
		newCardListCmd(),
		newCardGetCmd(),
		newCardGetOrderCmd(),
		newCardGetSecureCmd(),
		newCardIframeURLCmd(),
		newCardCreateCmd(),
		newCardUpdateCmd(),
		newCardUpdateStatusCmd(),
		newCardActivateCmd(),
		newCardSetPinCmd(),
		newCardAssignCmd(),
		newCardRechargeCmd(),
		newCardWithdrawCmd(),
	)
	return cmd
}

func newCardGetOrderCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-order <order-id>",
		Short: "Retrieve a card order by order ID",
		Long: `Retrieve the status and details of a card order.

The <order-id> is the card_order_id returned in the response of "card create".
Use this to poll async card creation status (PENDING → PROCESSING → SUCCESS/FAILED).

Examples:
  uqpay issuing card get-order 11233ad9-c080-4bd8-8df9-630dfeb487a7`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cards/"+args[0]+"/order", nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newCardListCmd() *cobra.Command {
	var status, cardNumber, cardholderID, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List issuing cards",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cards", map[string]string{
				"card_status":   status,
				"card_number":   cardNumber,
				"cardholder_id": cardholderID,
				"page_size":     pageSize,
				"page_number":   pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by card status: ACTIVE | FROZEN | BLOCKED | PENDING | CANCELLED | LOST | STOLEN | FAILED")
	cmd.Flags().StringVar(&cardNumber, "card-number", "", "Filter by card number (masked or full)")
	cmd.Flags().StringVar(&cardholderID, "cardholder-id", "", "Filter by cardholder ID")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newCardGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <card-id>",
		Short: "Retrieve a card",
		Long: `Retrieve full details of a card by its card ID.

Examples:
  uqpay issuing card get c03a7ac3-42f2-437d-8f5d-c328c28f1012`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cards/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newCardGetSecureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get-secure <card-id>",
		Short: "Retrieve sensitive card data (licensed institutions only)",
		Long: `Retrieve the full unmasked card number, expiry date, and CVV.

Requires your account to have a licensed institution permission. If you receive
a 403 error, use "card iframe-url" instead to display card details via a
PCI-compliant embedded iframe.

Examples:
  uqpay issuing card get-secure c03a7ac3-42f2-437d-8f5d-c328c28f1012`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cards/"+args[0]+"/secure", nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newCardIframeURLCmd() *cobra.Command {
	var lang, styles string
	cmd := &cobra.Command{
		Use:   "iframe-url <card-id>",
		Short: "Get a PCI-compliant iframe URL to display sensitive card data",
		Long: `Get a PCI-compliant iframe URL hosted by UQPAY. Open the URL in a browser or embed in your frontend.
Use this instead of get-secure if your account is not a licensed institution.

Examples:
  uqpay issuing card iframe-url card_xxx
  uqpay issuing card iframe-url card_xxx --lang zh`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			c := client.New(cfg)
			// 1. Get PAN token
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/token", map[string]any{})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			var tokenResp struct {
				Token     string `json:"token"`
				ExpiresAt string `json:"expires_at"`
				ExpiresIn int    `json:"expires_in"`
			}
			if err := cmdutil.ParseJSON(resp, &tokenResp); err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			// 2. Build iframe URL
			params := url.Values{}
			params.Set("token", tokenResp.Token)
			params.Set("cardId", args[0])
			if lang != "" {
				params.Set("lang", lang)
			}
			if styles != "" {
				params.Set("styles", styles)
			}
			iframeURL := c.IframeBase() + "/iframe/card?" + params.Encode()
			// 3. Output
			result := map[string]any{
				"iframe_url": iframeURL,
				"expires_at": tokenResp.ExpiresAt,
				"expires_in": tokenResp.ExpiresIn,
			}
			b, err := cmdutil.MarshalJSON(result)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, b, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&lang, "lang", "", "Display language (e.g. zh, en)")
	cmd.Flags().StringVar(&styles, "styles", "", "Custom CSS styles as JSON")
	return cmd
}

func newCardCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new issuing card",
		Long:  cardCreateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			dotparam.CoerceNumbers(body, "card_limit", "amount", "no_pin_payment_amount")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards", body)
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

func newCardUpdateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "update <card-id>",
		Short: "Update a card",
		Long:  cardUpdateHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			dotparam.CoerceNumbers(body, "card_limit", "amount", "no_pin_payment_amount")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/"+args[0], body)
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

func newCardUpdateStatusCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "update-status <card-id>",
		Short: "Update card status",
		Long:  cardUpdateStatusHelp,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/status", body)
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

func newCardActivateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activate a physical card",
		Long:  cardActivateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/activate", body)
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

func newCardSetPinCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "set-pin",
		Short: "Set or reset card PIN",
		Long:  cardSetPinHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/pin", body)
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

func newCardAssignCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign a physical card to a cardholder",
		Long:  cardAssignHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/assign", body)
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

func newCardRechargeCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "recharge <card-id>",
		Short: "Recharge card balance",
		Long: `Recharge the balance of a card.

Parameters:
  Required:
    amount   number   Amount to add

Examples:
  uqpay issuing card recharge card_xxx -d amount=100`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			dotparam.CoerceNumbers(body, "amount")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/recharge", body)
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

func newCardWithdrawCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "withdraw <card-id>",
		Short: "Withdraw from card balance",
		Long: `Withdraw funds from a card balance.

Parameters:
  Required:
    amount   number   Amount to withdraw

Examples:
  uqpay issuing card withdraw card_xxx -d amount=50`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			dotparam.CoerceNumbers(body, "amount")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/withdraw", body)
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
