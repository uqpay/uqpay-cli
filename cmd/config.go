package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/apierr"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	cmd.AddCommand(newConfigSetCmd(), newConfigGetCmd())
	return cmd
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a config value",
		Long: `Set a configuration value. Keys: client-id, api-key, env, output.

Examples:
  uqpay config set client-id your_client_id
  uqpay config set api-key your_api_key
  uqpay config set env sandbox
  uqpay config set output json`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			if err := cfg.Set(args[0], args[1]); err != nil {
				cmdutil.WriteError(&apierr.ConfigError{Message: err.Error()}, cfg.Output)
				return err
			}
			fmt.Printf("config: %s set to %q\n", args[0], args[1])
			return nil
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [key]",
		Short: "Get config value(s)",
		Long: `Get one or all configuration values.

Examples:
  uqpay config get            # show all
  uqpay config get api-key`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				cmdutil.WriteError(err, "table")
				return err
			}
			if len(args) == 0 {
				fmt.Printf("client-id: %s\n", cfg.ClientID)
				fmt.Printf("api-key:   %s\n", maskSecret(cfg.APIKey))
				fmt.Printf("env:       %s\n", cfg.Env)
				fmt.Printf("output:    %s\n", cfg.Output)
				return nil
			}
			switch args[0] {
			case "client-id":
				fmt.Println(cfg.ClientID)
			case "api-key":
				fmt.Println(maskSecret(cfg.APIKey))
			case "env":
				fmt.Println(cfg.Env)
			case "output":
				fmt.Println(cfg.Output)
			default:
				err := &apierr.ConfigError{Message: "unknown key: " + args[0]}
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return nil
		},
	}
}

// maskSecret masks all but the last 4 characters of a secret.
func maskSecret(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return "****" + s[len(s)-4:]
}
