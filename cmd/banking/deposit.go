package banking

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newDepositCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit",
		Short: "Manage banking deposits",
	}
	cmd.AddCommand(
		newDepositListCmd(),
		newDepositGetCmd(),
	)
	return cmd
}

func newDepositListCmd() *cobra.Command {
	var onBehalfOf, status, startTime, endTime, pageSize, pageNum string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List deposits",
		Long: `List banking deposits.

Flags:
  --status        Filter by status: PENDING | COMPLETED | FAILED
  --start-time    Filter by created time start (ISO 8601, e.g. 2026-01-01T00:00:00Z)
  --end-time      Filter by created time end (ISO 8601)
  --page-size     Results per page (default 10)
  --page-num      Page number (default 1)
  --on-behalf-of  Sub-account ID to act on behalf of`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/deposit", map[string]string{
				"status":      status,
				"start_time":  startTime,
				"end_time":    endTime,
				"page_size":   pageSize,
				"page_number": pageNum,
			}, map[string]string{"x-on-behalf-of": onBehalfOf})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&status, "status", "", "Filter by status: PENDING | COMPLETED | FAILED")
	cmd.Flags().StringVar(&startTime, "start-time", "", "Filter start time (ISO 8601)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "Filter end time (ISO 8601)")
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page (default 10)")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number (default 1)")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID to act on behalf of")
	return cmd
}

func newDepositGetCmd() *cobra.Command {
	var onBehalfOf string
	cmd := &cobra.Command{
		Use:   "get <deposit-id>",
		Short: "Retrieve a deposit",
		Long: `Retrieve a deposit by its ID.

The deposit ID is returned in the response of "deposit list".

Examples:
  uqpay deposit get dep_xxx`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.GetH(context.Background(), "/v1/deposit/"+args[0], nil,
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
