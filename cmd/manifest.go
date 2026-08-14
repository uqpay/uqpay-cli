package cmd

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const operationManifestSchema = "uqpay.operation-manifest.v1"

//go:embed operation-manifest.schema.json
var operationManifestJSONSchema []byte

type operationManifest struct {
	Schema     string              `json:"schema"`
	Operations []manifestOperation `json:"operations"`
}

type manifestOperation struct {
	OperationID      string              `json:"operation_id"`
	APIOperationID   string              `json:"api_operation_id"`
	Command          []string            `json:"command"`
	Domain           string              `json:"domain"`
	Resource         string              `json:"resource"`
	Action           string              `json:"action"`
	HTTPMethod       string              `json:"http_method"`
	HTTPPath         string              `json:"http_path"`
	Risk             operationRisk       `json:"risk"`
	Parameters       []manifestParameter `json:"parameters"`
	Body             manifestBody        `json:"body"`
	SupportsOnBehalf bool                `json:"supports_on_behalf_of"`
	Idempotency      manifestIdempotency `json:"idempotency"`
	DocumentationURL string              `json:"documentation_url"`
}

type operationRisk string

const (
	riskRead        operationRisk = "read"
	riskWrite       operationRisk = "write"
	riskDestructive operationRisk = "destructive"
)

type manifestParameter struct {
	Name        string `json:"name"`
	Location    string `json:"location"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Repeatable  bool   `json:"repeatable,omitempty"`
	Default     string `json:"default,omitempty"`
	Description string `json:"description,omitempty"`
}

type manifestBody struct {
	Supported            bool                `json:"supported"`
	Fields               []manifestBodyField `json:"fields"`
	AdditionalProperties bool                `json:"additional_properties"`
}

type manifestBodyField struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
}

type manifestIdempotency struct {
	Required bool   `json:"required"`
	Mode     string `json:"mode"`
}

type operationBinding struct {
	Command            string
	OperationID        string
	Domain             string
	DocsScope          string
	Method             string
	Path               string
	Risk               operationRisk
	FixedBodyFields    []manifestBodyField
	AdditionalBodyData bool
}

func newOperationManifestCmd(root *cobra.Command) *cobra.Command {
	var schemaOnly bool
	command := &cobra.Command{
		Use:   "operation-manifest",
		Short: "Export the versioned machine-readable API operation contract",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if schemaOnly {
				if _, err := cmd.OutOrStdout().Write(operationManifestJSONSchema); err != nil {
					return err
				}
				_, err := fmt.Fprintln(cmd.OutOrStdout())
				return err
			}
			manifest, err := buildOperationManifest(root)
			if err != nil {
				return err
			}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			encoder.SetEscapeHTML(false)
			return encoder.Encode(manifest)
		},
	}
	command.Flags().BoolVar(&schemaOnly, "schema", false, "Output the JSON Schema for the manifest instead of operation data")
	return command
}

func buildOperationManifest(root *cobra.Command) (operationManifest, error) {
	commands := canonicalRunnableCommands(root)
	bindings := operationBindings()
	if len(commands) != len(bindings) {
		return operationManifest{}, fmt.Errorf(
			"operation registry drift: %d canonical API commands, %d bindings",
			len(commands), len(bindings),
		)
	}

	operations := make([]manifestOperation, 0, len(bindings))
	seenIDs := make(map[string]string, len(bindings))
	for commandPath, binding := range bindings {
		command := commands[commandPath]
		if command == nil {
			return operationManifest{}, fmt.Errorf("operation registry command not found: uqpay %s", commandPath)
		}
		stableID := binding.Domain + "." + binding.OperationID
		if previous, exists := seenIDs[stableID]; exists {
			return operationManifest{}, fmt.Errorf("duplicate operation_id %q for %s and %s", stableID, previous, commandPath)
		}
		seenIDs[stableID] = commandPath

		parts := strings.Split(commandPath, " ")
		resource, action := commandResourceAction(parts)
		parameters, supportsOnBehalf, explicitIdempotency := commandParameters(command)
		bodyFields := mergeBodyFields(extractBodyFields(command.Long), binding.FixedBodyFields)
		hasDataFlag := command.Flags().Lookup("data") != nil
		bodySupported := binding.Method != "GET" || hasDataFlag || len(bodyFields) > 0
		idempotencyRequired := binding.Method != "GET" || explicitIdempotency || commandUsesIdempotencyOnRead(commandPath)
		idempotencyMode := "none"
		if idempotencyRequired {
			idempotencyMode = "auto_generated"
			if explicitIdempotency {
				idempotencyMode = "caller_required"
			}
		}

		operations = append(operations, manifestOperation{
			OperationID:    stableID,
			APIOperationID: binding.OperationID,
			Command:        append([]string{"uqpay"}, parts...),
			Domain:         binding.Domain,
			Resource:       resource,
			Action:         action,
			HTTPMethod:     binding.Method,
			HTTPPath:       binding.Path,
			Risk:           binding.Risk,
			Parameters:     parameters,
			Body: manifestBody{
				Supported:            bodySupported,
				Fields:               bodyFields,
				AdditionalProperties: hasDataFlag || binding.AdditionalBodyData,
			},
			SupportsOnBehalf: supportsOnBehalf,
			Idempotency: manifestIdempotency{
				Required: idempotencyRequired,
				Mode:     idempotencyMode,
			},
			DocumentationURL: "https://developers.uqpay.com/" + binding.DocsScope + "/v1.6/api-reference/" + binding.OperationID,
		})
	}

	sort.Slice(operations, func(i, j int) bool {
		return strings.Join(operations[i].Command, " ") < strings.Join(operations[j].Command, " ")
	})
	return operationManifest{Schema: operationManifestSchema, Operations: operations}, nil
}

func canonicalRunnableCommands(root *cobra.Command) map[string]*cobra.Command {
	allowedRoots := map[string]bool{
		"account": true, "rfi": true, "file": true, "simulate": true,
		"banking": true, "issuing": true, "payment": true,
	}
	commands := make(map[string]*cobra.Command)
	var visit func(*cobra.Command, []string)
	visit = func(parent *cobra.Command, path []string) {
		for _, child := range parent.Commands() {
			childPath := append(append([]string{}, path...), child.Name())
			if len(childPath) == 1 && !allowedRoots[childPath[0]] {
				continue
			}
			if child.Runnable() {
				commands[strings.Join(childPath, " ")] = child
			}
			visit(child, childPath)
		}
	}
	visit(root, nil)
	return commands
}

func commandResourceAction(parts []string) (string, string) {
	if len(parts) == 1 {
		return parts[0], "execute"
	}
	if parts[0] == "banking" || parts[0] == "issuing" || parts[0] == "payment" {
		if len(parts) == 2 {
			return parts[1], "list"
		}
		return parts[1], strings.Join(parts[2:], ".")
	}
	return parts[0], strings.Join(parts[1:], ".")
}

func commandParameters(command *cobra.Command) ([]manifestParameter, bool, bool) {
	parameters := argumentParameters(command.Use)
	supportsOnBehalf := false
	explicitIdempotency := false
	command.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Hidden || flag.Name == "data" || isCredentialOrPresentationFlag(flag.Name) {
			return
		}
		location := "option"
		if flag.Name == "on-behalf-of" {
			location = "header"
			supportsOnBehalf = true
		}
		if flag.Name == "idempotency-key" {
			location = "header"
			explicitIdempotency = true
		}
		parameters = append(parameters, manifestParameter{
			Name:        flag.Name,
			Location:    location,
			Type:        flag.Value.Type(),
			Required:    flagRequired(flag),
			Repeatable:  flag.Value.Type() == "stringArray" || flag.Value.Type() == "stringSlice",
			Default:     defaultIfMeaningful(flag.DefValue),
			Description: flag.Usage,
		})
	})
	sort.Slice(parameters, func(i, j int) bool {
		if parameters[i].Location == parameters[j].Location {
			return parameters[i].Name < parameters[j].Name
		}
		return parameters[i].Location < parameters[j].Location
	})
	return parameters, supportsOnBehalf, explicitIdempotency
}

func argumentParameters(use string) []manifestParameter {
	fields := strings.Fields(use)
	parameters := make([]manifestParameter, 0)
	for _, field := range fields[1:] {
		required := strings.HasPrefix(field, "<")
		if !required && !strings.HasPrefix(field, "[") {
			continue
		}
		name := strings.Trim(field, "<>[]")
		repeatable := strings.HasSuffix(name, "...")
		name = strings.TrimSuffix(name, "...")
		parameters = append(parameters, manifestParameter{
			Name: name, Location: "argument", Type: "string", Required: required, Repeatable: repeatable,
		})
	}
	return parameters
}

var bodyFieldPattern = regexp.MustCompile(`^\s{4,}([A-Za-z][A-Za-z0-9_.\[\]-]*)\s+(string|number|integer|bool|boolean|object|array)\b`)

func extractBodyFields(longHelp string) []manifestBodyField {
	fields := make([]manifestBodyField, 0)
	seen := make(map[string]bool)
	required := false
	for _, line := range strings.Split(longHelp, "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(trimmed, "Required"):
			required = true
			continue
		case strings.HasPrefix(trimmed, "Optional"):
			required = false
			continue
		case strings.HasPrefix(trimmed, "Example"):
			required = false
		}
		match := bodyFieldPattern.FindStringSubmatch(line)
		if match == nil || seen[match[1]] {
			continue
		}
		seen[match[1]] = true
		fieldType := match[2]
		if fieldType == "bool" {
			fieldType = "boolean"
		}
		fields = append(fields, manifestBodyField{Name: match[1], Type: fieldType, Required: required})
	}
	return fields
}

func mergeBodyFields(discovered, fixed []manifestBodyField) []manifestBodyField {
	byName := make(map[string]manifestBodyField, len(discovered)+len(fixed))
	for _, field := range discovered {
		byName[field.Name] = field
	}
	for _, field := range fixed {
		byName[field.Name] = field
	}
	fields := make([]manifestBodyField, 0, len(byName))
	for _, field := range byName {
		fields = append(fields, field)
	}
	sort.Slice(fields, func(i, j int) bool { return fields[i].Name < fields[j].Name })
	return fields
}

func flagRequired(flag *pflag.Flag) bool {
	values := flag.Annotations[cobra.BashCompOneRequiredFlag]
	return (len(values) > 0 && values[0] == "true") ||
		strings.Contains(strings.ToLower(flag.Usage), "required")
}

func isCredentialOrPresentationFlag(name string) bool {
	switch name {
	case "api-key", "client-id", "env", "output", "debug":
		return true
	default:
		return false
	}
}

func defaultIfMeaningful(value string) string {
	if value == "" || value == "[]" || value == "false" {
		return ""
	}
	return value
}

func commandUsesIdempotencyOnRead(commandPath string) bool {
	return strings.HasPrefix(commandPath, "payment balance ") ||
		strings.HasPrefix(commandPath, "payment bank-account ") ||
		strings.HasPrefix(commandPath, "payment payout ") ||
		commandPath == "payment settlement list"
}
