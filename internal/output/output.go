package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/jedib0t/go-pretty/v6/table"
	"gopkg.in/yaml.v3"
)

// Print writes data (raw API JSON bytes) to w in the given format.
// format: "table" (default), "json", "yaml"
func Print(w io.Writer, data []byte, format string) error {
	switch format {
	case "json":
		return printJSON(w, data)
	case "yaml":
		return printYAML(w, data)
	default:
		return printTable(w, data)
	}
}

// Stdout is a convenience wrapper that writes to os.Stdout.
func Stdout(data []byte, format string) error {
	return Print(os.Stdout, data, format)
}

func printJSON(w io.Writer, data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		_, err = w.Write(data)
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func printYAML(w io.Writer, data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	return yaml.NewEncoder(w).Encode(v)
}

func printTable(w io.Writer, data []byte) error {
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Fprintln(w, string(data))
		return nil
	}
	switch typed := v.(type) {
	case []any:
		return renderList(w, typed)
	case map[string]any:
		if arr, ok := typed["data"].([]any); ok {
			if err := renderList(w, arr); err != nil {
				return err
			}
			if len(arr) >= 10 {
				fmt.Fprintf(w, "\n(%d results — use --page-num 2 for the next page, --page-size to change limit)\n", len(arr))
			}
			return nil
		}
		return renderRecord(w, typed)
	default:
		fmt.Fprintln(w, string(data))
	}
	return nil
}

func renderList(w io.Writer, rows []any) error {
	if len(rows) == 0 {
		fmt.Fprintln(w, "No results.")
		return nil
	}
	firstRow, ok := rows[0].(map[string]any)
	if !ok {
		b, _ := json.MarshalIndent(rows, "", "  ")
		fmt.Fprintln(w, string(b))
		return nil
	}
	headers := sortedKeys(firstRow)

	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateHeader = false

	// Header row
	headerRow := make(table.Row, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	t.AppendHeader(headerRow)

	for _, row := range rows {
		m, ok := row.(map[string]any)
		if !ok {
			continue
		}
		cells := make(table.Row, len(headers))
		for i, h := range headers {
			cells[i] = stringify(m[h])
		}
		t.AppendRow(cells)
	}
	t.Render()
	return nil
}

func renderRecord(w io.Writer, record map[string]any) error {
	t := table.NewWriter()
	t.SetOutputMirror(w)
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateHeader = false

	for _, k := range sortedKeys(record) {
		t.AppendRow(table.Row{k, stringify(record[k])})
	}
	t.Render()
	return nil
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func stringify(v any) string {
	if v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "true"
		}
		return "false"
	case map[string]any, []any:
		b, _ := json.Marshal(v)
		s := string(b)
		if len(s) > 60 {
			return s[:57] + "..."
		}
		return s
	default:
		return fmt.Sprintf("%v", v)
	}
}
