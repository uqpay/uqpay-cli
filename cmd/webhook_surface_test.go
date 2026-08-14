package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// The CLI is an API command client and raw structured-output tool. It does not
// claim to parse or type webhook deliveries; SDKs and customer webhook handlers
// own that concern.
func TestCLIExposesNoWebhookParserCommand(t *testing.T) {
	var visit func(*cobra.Command)
	visit = func(command *cobra.Command) {
		if strings.Contains(strings.ToLower(command.Name()), "webhook") {
			t.Errorf("unexpected webhook parser command in CLI surface: %s", command.CommandPath())
		}
		command.LocalNonPersistentFlags().VisitAll(func(flag *pflag.Flag) {
			if strings.Contains(strings.ToLower(flag.Name), "webhook") {
				t.Errorf("unexpected webhook parser flag in CLI surface: %s --%s", command.CommandPath(), flag.Name)
			}
		})
		for _, child := range command.Commands() {
			visit(child)
		}
	}
	visit(NewRootCmd())
}
