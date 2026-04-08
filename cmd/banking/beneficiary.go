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

const beneficiaryCreateHelp = `Create a new beneficiary.

Parameters:
  Required (COMPANY entity):
    entity_type                   string   COMPANY | INDIVIDUAL
    company_name                  string   Company name
    payment_method                string   LOCAL | SWIFT
    bank_details.bank_name        string   Name of the bank
    bank_details.bank_address     string   Address of the bank
    bank_details.bank_country_code string  Two-letter country code (ISO 3166-1 alpha-2)
    bank_details.account_holder   string   Account holder name
    bank_details.account_currency_code string  ISO 4217 currency code
    bank_details.swift_code       string   SWIFT/BIC code
    bank_details.clearing_system  string   Clearing system
    address.country               string   Two-letter country code
    address.city                  string   City
    address.street_address        string   Street address
    address.postal_code           string   Postal code
    address.state                 string   State or province

  Required (INDIVIDUAL entity):
    entity_type                   string   COMPANY | INDIVIDUAL
    first_name                    string   First name
    last_name                     string   Last name
    payment_method                string   LOCAL | SWIFT
    bank_details.bank_name        string   Name of the bank
    bank_details.bank_address     string   Address of the bank
    bank_details.bank_country_code string  Two-letter country code
    bank_details.account_holder   string   Account holder name
    bank_details.account_currency_code string  ISO 4217 currency code
    bank_details.swift_code       string   SWIFT/BIC code
    bank_details.clearing_system  string   Clearing system
    address.country               string   Two-letter country code
    address.city                  string   City
    address.street_address        string   Street address
    address.postal_code           string   Postal code
    address.state                 string   State or province

  Optional:
    nickname                           string   Nickname for the beneficiary
    email                              string   Email address
    bank_details.account_number        string   Account number (or iban)
    bank_details.iban                  string   IBAN (or account_number)
    bank_details.routing_code_type1    string   Routing code type (e.g. ach, sort_code)
    bank_details.routing_code_value1   string   Routing code value
    bank_details.routing_code_type2    string   Routing code sub-type (e.g. branch_code)
    bank_details.routing_code_value2   string   Routing code sub-type value
    address.nationality                string   Two-letter nationality code
    additional_info.organization_code  string   Unified Social Credit Identifier
    additional_info.proxy_id           string   PayNow proxy ID (SGD)
    additional_info.id_type            string   PASSPORT | NATIONAL_ID | DRIVERS_LICENSE
    additional_info.id_number          string   ID number
    additional_info.tax_id             string   Tax identification number
    additional_info.msisdn             string   Mobile phone number (+country_code...)

Examples:
  uqpay beneficiary create \
    -d entity_type=COMPANY \
    -d company_name="Acme Corp" \
    -d payment_method=SWIFT \
    -d bank_details.bank_name="DBS Bank" \
    -d bank_details.bank_address="12 Marina Blvd" \
    -d bank_details.bank_country_code=SG \
    -d bank_details.account_holder="Acme Corp" \
    -d bank_details.account_currency_code=SGD \
    -d bank_details.account_number=1234567890 \
    -d bank_details.swift_code=DBSSSGSG \
    -d bank_details.clearing_system=SWIFT \
    -d address.country=SG \
    -d address.city=Singapore \
    -d address.street_address="12 Marina Blvd" \
    -d address.postal_code=018982 \
    -d address.state=Singapore`

const beneficiaryUpdateHelp = `Update an existing beneficiary.

The beneficiary ID is returned in the response of "beneficiary create" or "beneficiary list".

Parameters (all optional — only include fields to update):
    entity_type                        string   COMPANY | INDIVIDUAL
    company_name                       string   Company name
    first_name                         string   First name (INDIVIDUAL only)
    last_name                          string   Last name (INDIVIDUAL only)
    payment_method                     string   LOCAL | SWIFT
    nickname                           string   Nickname
    email                              string   Email address
    bank_details.bank_name             string   Name of the bank
    bank_details.bank_address          string   Address of the bank
    bank_details.bank_country_code     string   Two-letter country code
    bank_details.account_holder        string   Account holder name
    bank_details.account_currency_code string   ISO 4217 currency code
    bank_details.account_number        string   Account number
    bank_details.iban                  string   IBAN
    bank_details.swift_code            string   SWIFT/BIC code
    bank_details.clearing_system       string   Clearing system
    bank_details.routing_code_type1    string   Routing code type
    bank_details.routing_code_value1   string   Routing code value
    bank_details.routing_code_type2    string   Routing code sub-type
    bank_details.routing_code_value2   string   Routing code sub-type value
    address.country                    string   Two-letter country code
    address.city                       string   City
    address.street_address             string   Street address
    address.postal_code                string   Postal code
    address.state                      string   State or province
    address.nationality                string   Two-letter nationality code
    additional_info.organization_code  string   Unified Social Credit Identifier
    additional_info.proxy_id           string   PayNow proxy ID
    additional_info.id_type            string   PASSPORT | NATIONAL_ID | DRIVERS_LICENSE
    additional_info.id_number          string   ID number
    additional_info.tax_id             string   Tax identification number
    additional_info.msisdn             string   Mobile phone number

Examples:
  uqpay beneficiary update ben_xxx -d nickname="New Name"`

const beneficiaryCheckHelp = `Check if a beneficiary bank account is valid.

Parameters:
  Required:
    entity_type      string   COMPANY | INDIVIDUAL
    account_number   string   Bank account number (or iban)
    payment_method   string   LOCAL | SWIFT
    currency         string   ISO 4217 currency code

  Optional:
    first_name       string   First name (INDIVIDUAL only)
    last_name        string   Last name (INDIVIDUAL only)
    company_name     string   Company name (COMPANY only)
    clearing_system  string   LOCAL | SWIFT | ACH | FAST | MEPS | GIRO | Fedwire | Faster Payments | RTGS | FPS | EFT | Interac e-Transfer | Bill Payment | CHAPS | Bank Transfer | NPP | BPAY
    iban             string   IBAN (for European countries)
    additional_info  object   Additional information

Examples:
  uqpay beneficiary check \
    -d entity_type=COMPANY \
    -d account_number=1234567890 \
    -d payment_method=SWIFT \
    -d currency=USD`

func NewBeneficiaryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beneficiary",
		Short: "Manage banking beneficiaries",
	}
	cmd.AddCommand(
		newBeneficiaryListCmd(),
		newBeneficiaryGetCmd(),
		newBeneficiaryCreateCmd(),
		newBeneficiaryUpdateCmd(),
		newBeneficiaryDeleteCmd(),
		newBeneficiaryCheckCmd(),
		newBeneficiaryPaymentMethodsCmd(),
	)
	return cmd
}

func newBeneficiaryListCmd() *cobra.Command {
	var onBehalfOf, entityType, nickname, currency, companyName, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List beneficiaries",
		Long: `List beneficiaries.

Flags:
  --entity-type    Filter by entity type: COMPANY | INDIVIDUAL
  --nickname       Filter by nickname
  --currency       Filter by currency (ISO 4217)
  --company-name   Filter by company name
  --page-size      Results per page (default 10)
  --page-num       Page number (default 1)
  --on-behalf-of   Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/beneficiaries", map[string]string{
				"entity_type":  entityType,
				"nickname":     nickname,
				"currency":     currency,
				"company_name": companyName,
				"page_size":    pageSize,
				"page_number":  pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&entityType, "entity-type", "", "Filter by entity type: COMPANY | INDIVIDUAL")
	cmd.Flags().StringVar(&nickname, "nickname", "", "Filter by nickname")
	cmd.Flags().StringVar(&currency, "currency", "", "Filter by currency (ISO 4217)")
	cmd.Flags().StringVar(&companyName, "company-name", "", "Filter by company name")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newBeneficiaryGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <beneficiary-id>",
		Short: "Retrieve a beneficiary",
		Long: `Retrieve a beneficiary by its ID.

The beneficiary ID is returned in the response of "beneficiary create" or "beneficiary list".

Examples:
  uqpay beneficiary get ben_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/beneficiaries/"+args[0], nil,
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

func newBeneficiaryCreateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a beneficiary",
		Long:  beneficiaryCreateHelp,
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
			resp, err := c.PostH(context.Background(), "/v1/beneficiaries", body,
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

func newBeneficiaryUpdateCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "update <beneficiary-id>",
		Short: "Update a beneficiary",
		Long:  beneficiaryUpdateHelp,
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
			resp, err := c.PostH(context.Background(), "/v1/beneficiaries/"+args[0], body,
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

func newBeneficiaryDeleteCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "delete <beneficiary-id>",
		Short: "Delete a beneficiary",
		Long: `Delete a beneficiary by its ID.

The beneficiary ID is returned in the response of "beneficiary create" or "beneficiary list".

Examples:
  uqpay beneficiary delete ben_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			resp, err := c.PostH(context.Background(), "/v1/beneficiaries/"+args[0]+"/delete", nil,
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

func newBeneficiaryCheckCmd() *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check if a beneficiary bank account is valid",
		Long:  beneficiaryCheckHelp,
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
			resp, err := c.PostH(context.Background(), "/v1/beneficiaries/check", body,
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

func newBeneficiaryPaymentMethodsCmd() *cobra.Command {
	var currency, country string
	cmd := &cobra.Command{
		Use:   "payment-methods",
		Short: "List available payment methods for beneficiaries",
		Long: `List available payment methods for beneficiaries by currency and country.

Flags:
  --currency   ISO 4217 currency code (required)
  --country    Two-letter country code ISO 3166-1 alpha-2 (required)

Examples:
  uqpay beneficiary payment-methods --currency USD --country US`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/beneficiaries/paymentmethods", map[string]string{
				"currency": currency,
				"country":  country,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&currency, "currency", "", "ISO 4217 currency code (required)")
	cmd.Flags().StringVar(&country, "country", "", "Two-letter country code (required)")
	return cmd
}
