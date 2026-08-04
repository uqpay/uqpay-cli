package issuing

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newMerchantBrandCmd() *cobra.Command {
	var displayName, merchantCode, pageSize, pageNum string
	cmd := &cobra.Command{Use: "merchant-brand", Short: "List issuing merchant brands",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Get(context.Background(), "/v1/issuing/merchant_brands", map[string]string{
				"display_name": displayName, "merchant_code": merchantCode,
				"page_size": pageSize, "page_number": pageNum,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringVar(&displayName, "display-name", "", "Merchant display-name prefix")
	cmd.Flags().StringVar(&merchantCode, "merchant-code", "", "Exact merchant code")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number")
	return cmd
}
