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

const cardholderCreateHelp = `Create a new cardholder.

Parameters:
  Required:
    email                                string   Cardholder email
    first_name                           string   1-40 chars, letters and spaces only
    last_name                            string   1-40 chars, letters and spaces only
    country_code                         string   ISO 3166-1 alpha-2 (e.g. GB)
    phone_number                         string

  Optional:
    date_of_birth                        string   YYYY-MM-DD
    gender                               string   MALE | FEMALE
    nationality                          string   ISO 3166-1 alpha-2 (required for STANDARD/ENHANCED KYC)
    residential_address.country          string   ISO 3166-1 alpha-2 (required if address provided)
    residential_address.city             string   max 128 chars (required if address provided)
    residential_address.line1            string   max 255 chars (required if address provided)
    residential_address.state            string   max 128 chars
    residential_address.district         string   max 128 chars
    residential_address.line2            string   max 255 chars
    residential_address.line_en          string   max 255 chars, English address
    residential_address.postal_code      string   max 16 chars
    identity.type                        string   ID_CARD | PASSPORT (required if identity provided)
    identity.number                      string   (required if identity provided)
    identity.front_file                  string   base64 string or @filepath (required if identity provided)
    identity.back_file                   string   base64 string or @filepath (required if type=ID_CARD)
    identity.hand_file                   string   base64 string or @filepath, holding document
    kyc_verification.method              string   THIRD_PARTY | SUMSUB_REDIRECT (required if kyc_verification provided)
    kyc_verification.kyc_proof.provider  string   e.g. SUMSUB (required if method=THIRD_PARTY)
    kyc_verification.kyc_proof.reference_id string >=10 chars, globally unique (required if method=THIRD_PARTY)
    document_type                        string   pdf | png | jpg | jpeg
    document                             string   base64 string or @filepath, max 2MB

Examples:
  uqpay issuing cardholder create \
    -d email=john@example.com \
    -d first_name=John \
    -d last_name=Doe \
    -d country_code=GB \
    -d phone_number=+441234567890

  # Pass files directly — CLI reads and base64-encodes them automatically:
  uqpay cardholder create \
    -d email=john@example.com \
    -d first_name=John \
    -d last_name=Doe \
    -d country_code=GB \
    -d phone_number=+441234567890 \
    -d identity.type=PASSPORT \
    -d identity.number=AB123456 \
    -d identity.front_file=@./passport.jpg`

const cardholderUpdateHelp = `Update an existing cardholder.
Note: first_name and last_name cannot be updated.

Parameters (all optional):
    email                                string
    country_code                         string   ISO 3166-1 alpha-2
    phone_number                         string
    date_of_birth                        string   YYYY-MM-DD
    gender                               string   MALE | FEMALE
    nationality                          string   ISO 3166-1 alpha-2
    residential_address.country          string   ISO 3166-1 alpha-2 (required if address provided)
    residential_address.city             string   max 128 chars (required if address provided)
    residential_address.line1            string   max 255 chars (required if address provided)
    residential_address.state            string   max 128 chars
    residential_address.district         string   max 128 chars
    residential_address.line2            string   max 255 chars
    residential_address.line_en          string   max 255 chars
    residential_address.postal_code      string   max 16 chars
    identity.type                        string   ID_CARD | PASSPORT (required if identity provided)
    identity.number                      string   (required if identity provided)
    identity.front_file                  string   base64 string or @filepath (required if identity provided)
    identity.back_file                   string   base64 string or @filepath (required if type=ID_CARD)
    identity.hand_file                   string   base64 string or @filepath
    kyc_verification.method              string   THIRD_PARTY | SUMSUB_REDIRECT (required if kyc_verification provided)
    kyc_verification.kyc_proof.provider  string   (required if method=THIRD_PARTY)
    kyc_verification.kyc_proof.reference_id string (required if method=THIRD_PARTY)
    document_type                        string   pdf | png | jpg | jpeg
    document                             string   base64 string or @filepath, max 2MB

Examples:
  uqpay issuing cardholder update ch_xxx -d email=new@example.com -d phone_number=+441234567890`

func NewCardholderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cardholder",
		Short: "Manage issuing cardholders",
	}
	cmd.AddCommand(
		newCardholderListCmd(),
		newCardholderGetCmd(),
		newCardholderCreateCmd(),
		newCardholderUpdateCmd(),
	)
	return cmd
}

func newCardholderListCmd() *cobra.Command {
	var status, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cardholders",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cardholders", map[string]string{
				"cardholder_status": status,
				"page_size":         pageSize,
				"page_number":       pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: PENDING | SUCCESS | FAILED | INCOMPLETE")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newCardholderGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <cardholder-id>",
		Short: "Retrieve a cardholder",
		Long: `Retrieve full details of a cardholder by their cardholder ID.

Examples:
  uqpay issuing cardholder get d070383a-bc8f-4955-a1ae-0bd393bc4053`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/cardholders/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newCardholderCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new cardholder",
		Long:  cardholderCreateHelp,
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
			resp, err := c.Post(context.Background(), "/v1/issuing/cardholders", body)
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

func newCardholderUpdateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "update <cardholder-id>",
		Short: "Update a cardholder",
		Long:  cardholderUpdateHelp,
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
			resp, err := c.Post(context.Background(), "/v1/issuing/cardholders/"+args[0], body)
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
