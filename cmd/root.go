package cmd

import (
	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/cmd/banking"
	"github.com/uqpay/uqpay-cli/cmd/connect"
	"github.com/uqpay/uqpay-cli/cmd/issuing"
	"github.com/uqpay/uqpay-cli/cmd/payment"
	"github.com/uqpay/uqpay-cli/cmd/simulator"
	"github.com/uqpay/uqpay-cli/cmd/supporting"
	"github.com/uqpay/uqpay-cli/internal/build"
	"github.com/uqpay/uqpay-cli/internal/cmdutil"
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "uqpay",
		Short:         "UQPAY command-line tool",
		Version:       build.Version + " (" + build.Date + ")",
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	root.PersistentFlags().StringVar(&cmdutil.FlagEnv, "env", "", "Environment: sandbox | production")
	root.PersistentFlags().StringVar(&cmdutil.FlagClientID, "client-id", "", "Override client ID")
	root.PersistentFlags().StringVar(&cmdutil.FlagAPIKey, "api-key", "", "Override API key")
	root.PersistentFlags().StringVarP(&cmdutil.FlagOutput, "output", "o", "", "Output format: table | json | yaml")
	root.PersistentFlags().BoolVar(&cmdutil.FlagDebug, "debug", false, "Print HTTP request and response details")

	// Groups control how commands appear in `uqpay --help`
	root.AddGroup(
		&cobra.Group{ID: "domains", Title: "API Domains:"},
		&cobra.Group{ID: "other", Title: "Other Commands:"},
		&cobra.Group{ID: "shortcuts", Title: "Shortcuts:"},
	)

	// Domain groups
	bankingCmd := banking.NewCmd()
	bankingCmd.GroupID = "domains"
	root.AddCommand(bankingCmd)

	issuingCmd := issuing.NewCmd()
	issuingCmd.GroupID = "domains"
	root.AddCommand(issuingCmd)

	paymentCmd := payment.NewCmd()
	paymentCmd.GroupID = "domains"
	root.AddCommand(paymentCmd)

	// Other top-level commands
	accountCmd := connect.NewAccountCmd()
	accountCmd.GroupID = "other"
	root.AddCommand(accountCmd)

	configCmd := newConfigCmd()
	configCmd.GroupID = "other"
	root.AddCommand(configCmd)

	fileCmd := supporting.NewFileCmd()
	fileCmd.GroupID = "other"
	root.AddCommand(fileCmd)

	simulateCmd := simulator.NewSimulateCmd()
	simulateCmd.GroupID = "other"
	root.AddCommand(simulateCmd)

	setupCompletionCmd := newSetupCompletionCmd()
	setupCompletionCmd.GroupID = "other"
	root.AddCommand(setupCompletionCmd)

	// Hide cobra's built-in completion command (users should use setup-completion instead)
	root.CompletionOptions.HiddenDefaultCmd = true

	// Shortcuts: top-level aliases — each is an independent instance
	shortcuts := []struct {
		cmd   *cobra.Command
		short string
	}{
		{banking.NewBeneficiaryCmd(), "Manage banking beneficiaries (shortcut for 'banking beneficiary')"},
		{banking.NewConversionCmd(), "Manage currency conversions (shortcut for 'banking conversion')"},
		{banking.NewExchangeRateCmd(), "Query banking exchange rates (shortcut for 'banking exchange-rate')"},
		{banking.NewPayoutCmd(), "Manage banking payouts (shortcut for 'banking payout')"},
		{issuing.NewCardCmd(), "Manage issuing cards (shortcut for 'issuing card')"},
		{issuing.NewCardholderCmd(), "Manage issuing cardholders (shortcut for 'issuing cardholder')"},
	}
	for _, s := range shortcuts {
		s.cmd.Short = s.short
		s.cmd.GroupID = "shortcuts"
		root.AddCommand(s.cmd)
	}

	return root
}
