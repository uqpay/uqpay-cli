package cmd

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commandManifestEntry is intentionally transport-oriented: it exposes the
// public Cobra contract without credentials or runtime configuration.
type commandManifestEntry struct {
	Path     []string      `json:"path"`
	Use      string        `json:"use"`
	Short    string        `json:"short"`
	Long     string        `json:"long,omitempty"`
	Example  string        `json:"example,omitempty"`
	Runnable bool          `json:"runnable"`
	MinArgs  int           `json:"min_args"`
	MaxArgs  *int          `json:"max_args,omitempty"`
	Flags    []commandFlag `json:"flags,omitempty"`
}

type commandFlag struct {
	Name      string `json:"name"`
	Shorthand string `json:"shorthand,omitempty"`
	Usage     string `json:"usage"`
	Default   string `json:"default,omitempty"`
	Type      string `json:"type"`
}

func newCommandManifestCmd(root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "command-manifest",
		Short:  "Export the machine-readable CLI command contract",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			entries := collectCommandManifest(root)
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			return encoder.Encode(map[string]any{
				"schema_version": 1,
				"commands":       entries,
			})
		},
	}
	return cmd
}

func collectCommandManifest(root *cobra.Command) []commandManifestEntry {
	entries := make([]commandManifestEntry, 0)
	var visit func(*cobra.Command, []string)
	visit = func(current *cobra.Command, parent []string) {
		children := current.Commands()
		for _, child := range children {
			if child.Hidden || child.Name() == "command-manifest" {
				continue
			}
			path := append(append([]string{}, parent...), child.Name())
			entry := commandManifestEntry{
				Path:     path,
				Use:      child.Use,
				Short:    child.Short,
				Long:     child.Long,
				Example:  child.Example,
				Runnable: child.Run != nil || child.RunE != nil,
				Flags:    collectCommandFlags(child),
			}
			entry.MinArgs, entry.MaxArgs = inferArgBounds(child.Use)
			entries = append(entries, entry)
			visit(child, path)
		}
	}
	visit(root, nil)
	sort.Slice(entries, func(i, j int) bool {
		return strings.Join(entries[i].Path, " ") < strings.Join(entries[j].Path, " ")
	})
	return entries
}

func collectCommandFlags(cmd *cobra.Command) []commandFlag {
	flags := make([]commandFlag, 0)
	seen := make(map[string]struct{})
	appendFlag := func(flag *pflag.Flag) {
		if flag.Hidden || flag.Name == "api-key" || flag.Name == "client-id" || flag.Name == "env" || flag.Name == "output" || flag.Name == "debug" {
			return
		}
		if _, exists := seen[flag.Name]; exists {
			return
		}
		seen[flag.Name] = struct{}{}
		flags = append(flags, commandFlag{
			Name: flag.Name, Shorthand: flag.Shorthand, Usage: flag.Usage,
			Default: flag.DefValue, Type: flag.Value.Type(),
		})
	}
	cmd.NonInheritedFlags().VisitAll(appendFlag)
	cmd.InheritedFlags().VisitAll(appendFlag)
	sort.Slice(flags, func(i, j int) bool { return flags[i].Name < flags[j].Name })
	return flags
}

func inferArgBounds(use string) (int, *int) {
	fields := strings.Fields(use)
	if len(fields) <= 1 {
		zero := 0
		return 0, &zero
	}
	minArgs, maxArgs := 0, 0
	unbounded := false
	for _, field := range fields[1:] {
		if strings.Contains(field, "...") {
			unbounded = true
		}
		if strings.HasPrefix(field, "<") {
			minArgs++
			maxArgs++
		} else if strings.HasPrefix(field, "[") {
			maxArgs++
		}
	}
	if unbounded {
		return minArgs, nil
	}
	return minArgs, &maxArgs
}
