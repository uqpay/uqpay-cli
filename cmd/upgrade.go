package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/uqpay/uqpay-cli/internal/build"
	"github.com/uqpay/uqpay-cli/internal/update"
)

func newUpgradeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade CLI to the latest version",
		RunE: func(cmd *cobra.Command, args []string) error {
			current := build.Version
			fmt.Fprintf(os.Stderr, "Current version: %s\n", current)

			latest, err := update.LatestVersion()
			if err != nil {
				return fmt.Errorf("failed to check latest version: %w", err)
			}

			if current == latest {
				fmt.Fprintf(os.Stderr, "Already up to date.\n")
				return nil
			}

			fmt.Fprintf(os.Stderr, "Latest version:  %s\n", latest)
			fmt.Fprintf(os.Stderr, "Upgrading...\n\n")

			npmCmd := exec.Command("npm", "install", "-g", update.NpmPackage()+"@latest")
			npmCmd.Stdout = os.Stdout
			npmCmd.Stderr = os.Stderr
			npmCmd.Stdin = os.Stdin

			if err := npmCmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "\nUpgrade failed. Try manually:\n  npm install -g %s@latest\n", update.NpmPackage())
				return err
			}

			fmt.Fprintf(os.Stderr, "\nUpgrade complete.\n")
			return nil
		},
	}
}
