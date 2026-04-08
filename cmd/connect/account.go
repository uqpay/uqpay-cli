package connect

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

const accountCreateHelp = `Create a connected sub-account.

Parameters vary by entity_type. Use dot notation for nested fields and [] for arrays.

Parameters (COMPANY entity):
  Required:
    entity_type                                    string   COMPANY
    name                                           string   Business display name
    country                                        string   ISO 3166-1 alpha-2 country code (e.g. SG, US, GB)
    contact_details.email                          string   Contact email
    contact_details.phone                          string   Contact phone with country code
    business_details.legal_entity_name_english     string   Legal name in English
    business_details.incorporation_date            string   Incorporation date (YYYY-MM-DD)
    business_details.registration_number           string   Business registration number
    business_details.business_structure             string   LIMITED_COMPANY | SOLE_TRADER | PARTNERSHIP | LISTED | OTHERS
    business_details.merchant_category_code        string   MCC code (e.g. 7372)
    business_details.estimated_worker_count        string   BS001 (<50) | BS002 (50-100) | BS003 (100-500) | BS004 (>500)
    business_details.monthly_estimated_revenue.amount    string   TM001 (<$50K) | TM002 ($50K-$100K) | TM003 | TM004 | TM005
    business_details.monthly_estimated_revenue.currency  string   ISO 4217 currency code
    business_details.account_purpose[]             string   USE_API | USE_PLATFORM | ... (repeatable)
    registration_address.line1                     string   Street address
    registration_address.city                      string   City
    registration_address.state                     string   State/province
    registration_address.postal_code               string   Postal code
    business_address[0].line1                      string   Street address
    business_address[0].city                       string   City
    business_address[0].country                    string   ISO 3166-1 alpha-2 country code
    business_address[0].state                      string   State/province
    business_address[0].postal_code                string   Postal code
    representatives[0].roles                       string   DIRECTOR | BENEFICIAL_OWNER | SHAREHOLDER | ...
    representatives[0].as_applicant                bool     true if this representative is the applicant
    representatives[0].first_name                  string   First name
    representatives[0].last_name                   string   Last name
    representatives[0].nationality                 string   ISO 3166-1 alpha-2 country code
    representatives[0].date_of_birth               string   Date of birth (YYYY-MM-DD)
    representatives[0].identification.type         string   PASSPORT | ID_CARD | DRIVERS_LICENSE
    representatives[0].identification.id_number    string   ID document number
    representatives[0].identification.documents.front  string   ID front image (@+filepath for data URI)
    representatives[0].residential_address.line1   string   Residential street address (required)
    representatives[0].residential_address.city    string   Residential city (required)
    representatives[0].residential_address.country string   ISO 3166-1 alpha-2 (required)
    representatives[0].residential_address.postal_code string Residential postal code
    documents[0].type                              string   CERTIFICATE_OF_INCORPORATION | MEMORANDUM_AND_ARTICLES | ...
    documents[0].front                             string   Document image (@+filepath for data URI)
    tos_acceptance.ip                              string   IPv4 address of user accepting ToS
    tos_acceptance.date                            string   ISO 8601 datetime (e.g. 2026-01-01T00:00:00Z)
    tos_acceptance.user_agent                      string   Browser user agent string
    tos_acceptance.tos_agreement                   number   1 to accept (auto-converted to number)

Parameters (INDIVIDUAL entity):
  Required:
    entity_type              string   INDIVIDUAL
    name                     string   Display name
    country                  string   ISO 3166-1 alpha-2 country code
    contact_details          object   Contact information (email, phone)
    person_details           object   Personal details (first_name, last_name, date_of_birth, nationality)
    residential_address      object   Residential address (line1, city, state, postal_code, country)
    documents                array    Identity documents (type, front)
    tos_acceptance           object   Terms of service acceptance (ip, date, user_agent, tos_agreement)

Example (COMPANY):
  uqpay account create \
    -d entity_type=COMPANY \
    -d name="Acme Corp" \
    -d country=SG \
    -d contact_details.email=admin@acme.com \
    -d contact_details.phone=+6591234567 \
    -d business_details.legal_entity_name_english="Acme Corp Pte Ltd" \
    -d business_details.incorporation_date=2020-01-01 \
    -d business_details.registration_number=T99CC9999Z \
    -d business_details.business_structure=LIMITED_COMPANY \
    -d business_details.merchant_category_code=7372 \
    -d business_details.estimated_worker_count=BS001 \
    -d business_details.monthly_estimated_revenue.amount=TM001 \
    -d business_details.monthly_estimated_revenue.currency=SGD \
    -d "business_details.account_purpose[]=USE_API" \
    -d "registration_address.line1=1 Raffles Place" \
    -d registration_address.city=Singapore \
    -d registration_address.state=SG \
    -d registration_address.postal_code=048616 \
    -d "business_address[0].line1=1 Raffles Place" \
    -d "business_address[0].city=Singapore" \
    -d "business_address[0].country=SG" \
    -d "business_address[0].state=SG" \
    -d "business_address[0].postal_code=048616" \
    -d "representatives[0].roles=DIRECTOR" \
    -d "representatives[0].as_applicant=true" \
    -d "representatives[0].first_name=John" \
    -d "representatives[0].last_name=Doe" \
    -d "representatives[0].nationality=SG" \
    -d "representatives[0].date_of_birth=1990-01-15" \
    -d "representatives[0].identification.type=PASSPORT" \
    -d "representatives[0].identification.id_number=E1234567" \
    -d "representatives[0].identification.documents.front=@+./id_front.png" \
    -d "representatives[0].address.line1=10 Anson Road" \
    -d "representatives[0].address.city=Singapore" \
    -d "representatives[0].address.country=SG" \
    -d "representatives[0].address.postal_code=079903" \
    -d "representatives[0].residential_address.line1=10 Anson Road" \
    -d "representatives[0].residential_address.city=Singapore" \
    -d "representatives[0].residential_address.country=SG" \
    -d "representatives[0].residential_address.postal_code=079903" \
    -d "documents[0].type=CERTIFICATE_OF_INCORPORATION" \
    -d "documents[0].front=@+./cert.png" \
    -d tos_acceptance.ip=192.168.1.1 \
    -d tos_acceptance.date=2026-01-01T00:00:00Z \
    -d tos_acceptance.user_agent=uqpay-cli \
    -d tos_acceptance.tos_agreement=1`

func NewAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage connected accounts",
	}
	cmd.AddCommand(
		newAccountListCmd(),
		newAccountGetCmd(),
		newAccountCreateCmd(),
		newAccountCreateSubCmd(),
		newAccountAdditionalDocumentsCmd(),
	)
	return cmd
}

func newAccountListCmd() *cobra.Command {
	var pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List connected accounts",
		Long: `List all connected sub-accounts.

Flags:
  --page-size   Results per page (default 10)
  --page-num    Page number (default 1)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/accounts", map[string]string{
				"page_size":   pageSize,
				"page_number": pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}

func newAccountGetCmd() *cobra.Command {
	var businessCode string
	cmd := &cobra.Command{
		Use:   "get <account-id>",
		Short: "Retrieve a connected account",
		Long: `Retrieve a connected account by its ID.

The account ID is returned in the response of "account create" or "account list".

Flags:
  --business-code   Filter by business code: BANKING | ACQUIRING | ISSUING

Examples:
  uqpay account get acc_xxx
  uqpay account get acc_xxx --business-code ISSUING`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/accounts/"+args[0], map[string]string{
				"business_code": businessCode,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&businessCode, "business-code", "", "Filter by business code: BANKING | ACQUIRING | ISSUING")
	return cmd
}

func newAccountCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a connected account",
		Long:  accountCreateHelp,
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
			dotparam.CoerceNumbers(body, "tos_agreement")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/accounts", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	return cmd
}

func newAccountAdditionalDocumentsCmd() *cobra.Command {
	var country, businessCode string
	cmd := &cobra.Command{
		Use:   "additional-documents",
		Short: "Get additional documents required for account verification",
		Long: `Get the list of additional documents required for account verification.

Flags:
  --country        Two-letter country code ISO 3166-1 alpha-2 (required)
  --business-code  Business code: BANKING | ACQUIRING (required)

Examples:
  uqpay account additional-documents --country SG --business-code BANKING`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/accounts/get_additional", map[string]string{
				"country":       country,
				"business_code": businessCode,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&country, "country", "", "Two-letter country code (required)")
	cmd.Flags().StringVar(&businessCode, "business-code", "", "Business code: BANKING | ACQUIRING (required)")
	return cmd
}

const accountCreateSubHelp = `Create a sub-account under a connected account.

Parameters vary by entity_type. Use dot notation for nested fields and [] for arrays.

Parameters (INDIVIDUAL entity):
  Required:
    entity_type                                      string   INDIVIDUAL
    nickname                                         string   Display name (max 100 chars)
    individual_info.first_name_english               string   First name in English
    individual_info.last_name_english                string   Last name in English
    individual_info.nationality                      string   ISO 3166-1 alpha-2 country code
    individual_info.phone_number                     string   Phone with country code (e.g. +447911123456)
    individual_info.email_address                    string   Email address
    individual_info.date_of_birth                    string   Date of birth (YYYY-MM-DD)
    individual_info.country_or_territory             string   ISO 3166-1 alpha-2 country of residence
    individual_info.street_address                   string   Street address
    individual_info.city                             string   City
    individual_info.postal_code                      string   Postal code
    individual_info.employment_status                string   Employed | Self-Employed | Unemployed | Student | Retired | Homemaker | Other
    individual_info.industry                         string   Industry (see docs for full list)
    individual_info.job_title                        string   Job title (see docs for full list)
    individual_info.company_name                     string   Company name (max 100 chars)
    identity_verification.identification_type        string   PASSPORT | DRIVERS_LICENSE | NATIONAL_ID
    identity_verification.identification_value       string   ID document number
    identity_verification.identity_docs[]            string   Identity document image (base64 string, @+filepath, or file ID, repeatable)
    identity_verification.face_docs[]               string   Face photo image (base64 string, @+filepath, or file ID, repeatable)
    expected_activity.account_purpose[]              string   PURCHASE | BILL_PAYMENT | EDUCATIONAL_EXPENSES | PERSONAL_REMITTANCE | CHARITABLE_DONATION | LOAN_REPAYMENT | INVESTMENT | OTHERS (repeatable)
    expected_activity.banking_countries[]            string   ISO 3166-1 alpha-2 country codes (repeatable)
    expected_activity.banking_currencies[]           string   ISO 4217 currency codes (repeatable)
    expected_activity.internationally                integer  1 = international, 0 = domestic only
    expected_activity.turnover_monthly               string   TM001 (<$50K) | TM002 ($50K-$100K) | TM003 ($100K-$250K) | TM004 ($250K-$500K) | TM005 (>$500K)
    expected_activity.turnover_monthly_currency      string   ISO 4217 currency code for turnover
    proof_documents.proof_of_address[]              string   Address proof document (base64 string, @+filepath, or file ID, repeatable)
    tos_acceptance.ip                                string   IPv4 address of user accepting ToS
    tos_acceptance.date                              string   Acceptance date (ISO 8601, e.g. 2026-04-07T00:00:00Z)
    tos_acceptance.user_agent                        string   Browser user agent string
    tos_acceptance.tos_agreement                     integer  1 to auto-sign TPSP agreement

  Required (GB/US):
    individual_info.state                            string   State or province (required for GB, US)

  Optional:
    individual_info.apartment_suite_or_floor         string   Apartment/suite/floor
    individual_info.tax_number                       string   Tax identification number
    proof_documents.source_of_funds[]               string   Source of funds document (base64 string, @+filepath, or file ID; required for Virtual Account)
    proof_documents.proof_of_position_and_income[]  string   Position/income proof document (base64 string, @+filepath, or file ID)

Parameters (COMPANY entity):
  Required:
    entity_type                                      string   COMPANY
    nickname                                         string   Display name (max 100 chars)
    inherit                                          integer  1 (inherit from master) | -1 (do not inherit)
    tos_acceptance.ip                                string   IPv4 address of user accepting ToS
    tos_acceptance.date                              string   Acceptance date (ISO 8601)
    tos_acceptance.user_agent                        string   Browser user agent string
    tos_acceptance.tos_agreement                     integer  1 to auto-sign TPSP agreement

  Required (when inherit=-1):
    company_info.legal_business_name                 string   Legal business name in local language
    company_info.legal_business_name_english         string   Legal business name in English (max 255)
    company_info.country_of_incorporation            string   ISO 3166-1 alpha-2 country code
    company_info.company_type                        string   SOLE_PROPRIETOR | LIMITED_COMPANY | PARTNERSHIP | LISTED | OTHERS
    company_info.phone_number                        string   Company phone number
    company_info.email_address                       string   Company email address
    company_info.company_registration_number         string   Business registration number
    company_info.incorparate_date                    string   Incorporation date (YYYY-MM-DD)
    company_info.certification_of_incorporation[]   string   Incorporation certificate (base64 string, @+filepath, or file ID, repeatable)
    company_address.street_address                   string   Street address
    company_address.city                             string   City
    company_address.postal_code                      string   Postal code
    ownership_details.representatives[0].first_name_english   string   Representative first name
    ownership_details.shareholder_docs[]             string   Shareholder document (base64 string, @+filepath, or file ID)

Examples:
  uqpay account create-sub \
    -d entity_type=INDIVIDUAL \
    -d "nickname=My Sub Account" \
    -d individual_info.first_name_english=John \
    -d individual_info.last_name_english=Doe \
    -d individual_info.nationality=GB \
    -d individual_info.phone_number=+447911123456 \
    -d individual_info.email_address=john@example.com \
    -d individual_info.date_of_birth=1990-01-15 \
    -d individual_info.country_or_territory=GB \
    -d "individual_info.street_address=123 Baker Street" \
    -d individual_info.city=London \
    -d individual_info.postal_code=W1U6RS \
    -d individual_info.employment_status=Employed \
    -d "individual_info.industry=Information Technology/IT" \
    -d "individual_info.job_title=Business and administration professionals" \
    -d "individual_info.company_name=Acme Corp." \
    -d identity_verification.identification_type=PASSPORT \
    -d identity_verification.identification_value=P12345678 \
    -d "identity_verification.identity_docs[]=@+./id_front.png" \
    -d "identity_verification.face_docs[]=@+./selfie.png" \
    -d expected_activity.internationally=1 \
    -d "expected_activity.account_purpose[]=PURCHASE" \
    -d "expected_activity.banking_countries[]=GB" \
    -d "expected_activity.banking_currencies[]=GBP" \
    -d expected_activity.turnover_monthly=TM002 \
    -d expected_activity.turnover_monthly_currency=GBP \
    -d "proof_documents.proof_of_address[]=@+./address_proof.png" \
    -d tos_acceptance.ip=192.168.1.1 \
    -d tos_acceptance.date=2026-04-07T00:00:00Z \
    -d "tos_acceptance.user_agent=Mozilla/5.0" \
    -d tos_acceptance.tos_agreement=1`

func newAccountCreateSubCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create-sub",
		Short: "Create a sub-account under a connected account",
		Long:  accountCreateSubHelp,
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
			dotparam.CoerceNumbers(body, "inherit", "internationally", "tos_agreement", "ownership_percentage")
			c := client.New(cfg)
			resp, err := c.Post(context.Background(), "/v1/accounts/create_accounts", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs (repeatable), supports dot notation for nested fields")
	return cmd
}
