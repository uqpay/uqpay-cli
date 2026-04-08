package issuing

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

const reportCreateHelp = `Create a transaction or settlement report.

Parameters:
  Required:
    report_type   string   SETTLEMENT | LEDGER
    start_time    string   Start time (ISO 8601, e.g. 2026-01-01T00:00:00Z)
    end_time      string   End time (ISO 8601, e.g. 2026-01-31T23:59:59Z)

Examples:
  uqpay report create \
    -d report_type=SETTLEMENT \
    -d start_time=2026-01-01T00:00:00Z \
    -d end_time=2026-01-31T23:59:59Z`

func newReportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Manage issuing reports",
	}
	cmd.AddCommand(
		newReportCreateCmd(),
		newReportDownloadCmd(),
	)
	return cmd
}

func newReportCreateCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a report",
		Long:  reportCreateHelp,
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
			resp, err := c.Post(context.Background(), "/v1/issuing/reports", body)
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

func newReportDownloadCmd() *cobra.Command {
	var outFile string
	cmd := &cobra.Command{
		Use:   "download <report-id>",
		Short: "Download a report file",
		Long: `Download a report file by report ID. Saves to a local file.

Examples:
  uqpay report download rpt_xxx
  uqpay report download rpt_xxx --out my-report.csv`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			c := client.New(cfg)
			data, err := c.Get(context.Background(), "/v1/issuing/reports/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			if outFile == "" {
				outFile = "report-" + args[0] + ".csv"
			}
			if err := os.WriteFile(outFile, data, 0644); err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			fmt.Fprintf(os.Stdout, "Report saved to %s (%d bytes)\n", outFile, len(data))
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFile, "out", "O", "", "Output file path (default: report-<id>.csv)")
	return cmd
}
