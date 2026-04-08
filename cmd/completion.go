package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

func newSetupCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup-completion",
		Short: "Install shell completion for uqpay",
		Long: `Install shell tab-completion for uqpay.

Detects your current shell (zsh, bash, fish, or powershell) and appends
the completion loader to your shell config file.

Run this once after installing uqpay. Restart your shell or source your
config file to activate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := detectShell()
			if shell == "" {
				return fmt.Errorf("unsupported shell — supported: zsh, bash, fish, powershell\nManual setup: run 'uqpay completion --help'")
			}

			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("cannot find home directory: %w", err)
			}

			var rcFile, line, sourceHint string
			switch shell {
			case "zsh":
				rcFile = filepath.Join(home, ".zshrc")
				line = `eval "$(uqpay completion zsh)"`
				sourceHint = "source " + rcFile
			case "bash":
				rcFile = filepath.Join(home, ".bashrc")
				line = `eval "$(uqpay completion bash)"`
				sourceHint = "source " + rcFile
			case "fish":
				rcFile = filepath.Join(home, ".config", "fish", "completions", "uqpay.fish")
				line = "" // fish uses a generated file, not eval
				sourceHint = "restart your fish shell"
			case "powershell":
				rcFile = filepath.Join(home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
				line = `uqpay completion powershell | Out-String | Invoke-Expression`
				sourceHint = "restart PowerShell"
			}

			// Fish: write completion file directly
			if shell == "fish" {
				if err := os.MkdirAll(filepath.Dir(rcFile), 0755); err != nil {
					return fmt.Errorf("cannot create %s: %w", filepath.Dir(rcFile), err)
				}
				out, err := exec.Command("uqpay", "completion", "fish").Output()
				if err != nil {
					// fallback: generate from current binary
					out, err = exec.Command(os.Args[0], "completion", "fish").Output()
					if err != nil {
						return fmt.Errorf("failed to generate fish completions: %w", err)
					}
				}
				if err := os.WriteFile(rcFile, out, 0644); err != nil {
					return fmt.Errorf("cannot write to %s: %w", rcFile, err)
				}
				fmt.Printf("Completion installed in %s\n", rcFile)
				fmt.Printf("Restart your fish shell to activate.\n")
				return nil
			}

			// zsh/bash/powershell: append to rc file
			data, _ := os.ReadFile(rcFile)
			if strings.Contains(string(data), "uqpay completion") {
				fmt.Printf("Completion already installed in %s\n", rcFile)
				return nil
			}

			if err := os.MkdirAll(filepath.Dir(rcFile), 0755); err != nil {
				return fmt.Errorf("cannot create %s: %w", filepath.Dir(rcFile), err)
			}
			f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return fmt.Errorf("cannot write to %s: %w", rcFile, err)
			}
			defer f.Close()

			if _, err := fmt.Fprintf(f, "\n# uqpay shell completion\n%s\n", line); err != nil {
				return fmt.Errorf("failed to write: %w", err)
			}

			fmt.Printf("Completion installed in %s\n", rcFile)
			fmt.Printf("Run '%s' to activate.\n", sourceHint)
			return nil
		},
	}
}

func detectShell() string {
	// Check SHELL env var
	supported := map[string]bool{"zsh": true, "bash": true, "fish": true}

	sh := filepath.Base(os.Getenv("SHELL"))
	if supported[sh] {
		return sh
	}
	// Windows PowerShell
	if runtime.GOOS == "windows" {
		return "powershell"
	}
	// Fallback: check parent process on macOS/Linux
	ppid := os.Getppid()
	out, err := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=").Output()
	if err == nil {
		name := filepath.Base(strings.TrimSpace(string(out)))
		if supported[name] {
			return name
		}
	}
	return ""
}
