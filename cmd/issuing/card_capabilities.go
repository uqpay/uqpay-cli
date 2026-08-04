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

func newCardElevateLimitCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{Use: "elevate-limit <card-id>", Short: "Elevate a card per-transaction limit", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				return err
			}
			dotparam.CoerceNumbers(body, "limit_amount", "duration_in_days")
			resp, err := client.New(cfg).Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/elevate_limit", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "limit_amount and optional duration_in_days")
	return cmd
}

func newCardEnrollProtectionCmd() *cobra.Command {
	var actionCode string
	cmd := &cobra.Command{Use: "enroll-network-protection <card-id>", Short: "Enroll a card in Network Protection", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			body := map[string]any{"risk_control": "network_protection", "action_code": actionCode}
			resp, err := client.New(cfg).Post(context.Background(), "/v1/issuing/cards/"+args[0]+"/risk", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringVar(&actionCode, "action-code", "", "Network Protection action code")
	_ = cmd.MarkFlagRequired("action-code")
	return cmd
}

func newCardRemoveProtectionCmd() *cobra.Command {
	return &cobra.Command{Use: "remove-network-protection <card-id>", Short: "Remove a card from Network Protection", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Delete(context.Background(), "/v1/issuing/cards/"+args[0]+"/risk", map[string]any{"risk_control": "network_protection"})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
}

func newCardManagePinCmd() *cobra.Command {
	var data []string
	cmd := &cobra.Command{Use: "manage-pin", Short: "Set or reset a Standard virtual card PIN",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			body, err := dotparam.Parse(data)
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Post(context.Background(), "/v1/issuing/cards/manage/pin", body)
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringArrayVarP(&data, "data", "d", nil, "card_id, type, pin, and optional old_pin")
	return cmd
}

func newCardListArtsCmd() *cobra.Command {
	var productID string
	cmd := &cobra.Command{Use: "list-arts", Short: "List available card arts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Get(context.Background(), "/v1/issuing/cards/arts", map[string]string{"card_product_id": productID})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
	cmd.Flags().StringVar(&productID, "card-product-id", "", "Filter by card product ID")
	return cmd
}

func newCardSetDefaultArtCmd() *cobra.Command {
	return &cobra.Command{Use: "set-default-art <card-art-id>", Short: "Set the default card art", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := cmdutil.LoadConfig()
			if err != nil {
				return err
			}
			resp, err := client.New(cfg).Post(context.Background(), "/v1/issuing/cards/arts/default", map[string]any{"card_art_id": args[0]})
			if err != nil {
				cmdutil.WriteError(err, cfg.Output)
				return err
			}
			return output.Print(os.Stdout, resp, cfg.Output)
		}}
}
