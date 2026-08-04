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

func NewRfiCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "rfi", Short: "Manage Connect requests for information"}
	cmd.AddCommand(newRfiListCmd(), newRfiGetCmd(), newRfiAnswerCmd())
	return cmd
}

func newRfiListCmd() *cobra.Command {
	var pageSize, pageNum, status string
	cmd := &cobra.Command{
		Use: "list", Short: "List RFIs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			data, err := client.New(cfg).Get(context.Background(), "/v1/rfis", map[string]string{
				"page_size": pageSize, "page_number": pageNum, "status": status,
			})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
	cmd.Flags().StringVar(&pageSize, "page-size", "10", "Results per page")
	cmd.Flags().StringVar(&pageNum, "page-num", "1", "Page number")
	cmd.Flags().StringVar(&status, "status", "", "RFI status filter")
	return cmd
}

func newRfiGetCmd() *cobra.Command {
	return &cobra.Command{
		Use: "get <rfi-id>", Short: "Retrieve an RFI", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			data, err := client.New(cfg).Get(context.Background(), "/v1/rfis/"+args[0], nil)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, data, cfg.Output)
		},
	}
}

func newRfiAnswerCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{
		Use: "answer", Short: "Answer an RFI",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Post(context.Background(), "/v1/rfis/answer", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		},
	}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Key=value pairs; include rfi_id and answer[n].*")
	return cmd
}
