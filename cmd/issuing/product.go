package issuing

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newProductCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "product",
		Short: "List issuing card products",
	}
	cmd.AddCommand(newProductListCmd())
	return cmd
}

func newProductListCmd() *cobra.Command {
	var pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available card products",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/products", map[string]string{
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
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page, 10-100 (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	return cmd
}
