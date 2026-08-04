package payment

import (
	"context"
	"os"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/client"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
	"github.com/uqpay/uqpay-cli/internal/dotparam"
	"github.com/uqpay/uqpay-cli/internal/output"
)

func newTerminalCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "terminal", Short: "Manage payment terminals"}
	cmd.AddCommand(newTerminalRegisterCmd(), newTerminalGetPinKeyCmd())
	return cmd
}

func newTerminalRegisterCmd() *cobra.Command {
	return newTerminalPostCmd("register", "Register a terminal", "/v2/terminal/register")
}

func newTerminalGetPinKeyCmd() *cobra.Command {
	return newTerminalPostCmd("get-pin-key", "Get an encrypted terminal PIN key", "/v2/terminal/getPinKey")
}

func newTerminalPostCmd(use, short, path string) *cobra.Command {
	var data []string
	var onBehalfOf string
	cmd := &cobra.Command{Use: use, Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).PostH(context.Background(), path, body,
				map[string]string{
					"x-client-id":    cfg.ClientID,
					"x-on-behalf-of": onBehalfOf,
				})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "Request key=value pairs")
	cmd.Flags().StringVar(&onBehalfOf, "on-behalf-of", "", "Sub-account ID")
	return cmd
}
