package dotparam

import (
	"encoding/base64"
	"fmt"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// coerce converts a string value to the appropriate Go type for JSON serialization.
//
// Rules (applied in order):
//   - "num:<v>" prefix → parse as int64 or float64 (prefix stripped).
//   - "str:<v>" prefix → string (prefix stripped), kept for backward compatibility.
//   - Values wrapped in double-quotes → string (quotes stripped).
//   - "true" / "false" → bool.
//   - Everything else → string. No automatic number coercion.
//
// This "default string" design avoids the need for callers to know whether an API
// field expects a JSON string or number — the server handles the conversion.
// The rare fields that require a JSON number (e.g. inherit=-1) use the num: prefix.
func coerce(s string) any {
	if strings.HasPrefix(s, "num:") {
		v := s[4:]
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
		return v // fallback to string if not a valid number
	}
	if strings.HasPrefix(s, "str:") {
		return s[4:]
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}
	return s
}

// CoerceNumbers walks a parsed map and converts string values to numbers
// for fields whose leaf name matches the given set. This allows commands to
// declare which fields the API expects as JSON numbers, so users never need
// to type the num: prefix manually.
//
// Example:
//
//	body, _ := dotparam.Parse(data)
//	dotparam.CoerceNumbers(body, "amount", "transaction_amount", "card_limit")
func CoerceNumbers(m map[string]any, fields ...string) {
	set := make(map[string]bool, len(fields))
	for _, f := range fields {
		set[f] = true
	}
	coerceNumbersWalk(m, set)
}

func coerceNumbersWalk(m map[string]any, fields map[string]bool) {
	for k, v := range m {
		switch val := v.(type) {
		case string:
			if fields[k] {
				if n, err := strconv.ParseInt(val, 10, 64); err == nil {
					m[k] = n
				} else if f, err := strconv.ParseFloat(val, 64); err == nil {
					m[k] = f
				}
			}
		case map[string]any:
			coerceNumbersWalk(val, fields)
		case []any:
			for _, item := range val {
				if sub, ok := item.(map[string]any); ok {
					coerceNumbersWalk(sub, fields)
				}
			}
		}
	}
}

// Parse converts a slice of "key=value" strings into a nested map[string]any.
// Supports:
//
//	flat:    "currency=USD"
//	dot:     "bank_details.account_number=12345"
//	indexed: "spending_controls[0].amount=1000"
//	append:  "payment_method_types[]=card"
//	file:    "front_file=@/path/to/image.png"  (reads file and encodes as data URI)
func Parse(pairs []string) (map[string]any, error) {
	result := map[string]any{}
	for _, pair := range pairs {
		idx := strings.IndexByte(pair, '=')
		if idx < 0 {
			return nil, fmt.Errorf("invalid -d value %q: expected key=value", pair)
		}
		key, val := pair[:idx], pair[idx+1:]
		if strings.HasPrefix(val, "@") {
			encoded, err := encodeFile(val[1:])
			if err != nil {
				return nil, fmt.Errorf("invalid -d key %q: %w", key, err)
			}
			val = encoded
		}
		if err := setNested(result, key, val); err != nil {
			return nil, fmt.Errorf("invalid -d key %q: %w", key, err)
		}
	}
	return result, nil
}

// encodeFile reads a file and returns its content as a base64 string.
// Use path prefixed with "+" (e.g. "@+/path/to/file.png") for data URI format
// ("data:<mime>;base64,<b64>"), which is required by some endpoints (e.g. account create-sub).
// The default (plain "@path") returns pure base64 with no prefix.
func encodeFile(path string) (string, error) {
	dataURI := false
	if strings.HasPrefix(path, "+") {
		dataURI = true
		path = path[1:]
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", path, err)
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	if !dataURI {
		return encoded, nil
	}

	mimeType := mime.TypeByExtension(filepath.Ext(path))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}
	if semi := strings.IndexByte(mimeType, ';'); semi >= 0 {
		mimeType = strings.TrimSpace(mimeType[:semi])
	}
	return "data:" + mimeType + ";base64," + encoded, nil
}

// segRe matches one path segment: field name and optional [n] or [].
var segRe = regexp.MustCompile(`^([^.\[]+)(\[(\d*)\])?$`)

type segment struct {
	field string
	isArr bool
	index int // -1 means append ([] syntax)
}

func parseKey(key string) ([]segment, error) {
	parts := strings.Split(key, ".")
	segs := make([]segment, 0, len(parts))
	for _, part := range parts {
		m := segRe.FindStringSubmatch(part)
		if m == nil {
			return nil, fmt.Errorf("invalid segment %q", part)
		}
		seg := segment{field: m[1]}
		if m[2] != "" { // has [...]
			seg.isArr = true
			if m[3] == "" {
				seg.index = -1 // append
			} else {
				n, err := strconv.Atoi(m[3])
				if err != nil {
					return nil, err
				}
				seg.index = n
			}
		}
		segs = append(segs, seg)
	}
	return segs, nil
}

func setNested(root map[string]any, key, val string) error {
	segs, err := parseKey(key)
	if err != nil {
		return err
	}
	return applySegments(root, segs, val)
}

func applySegments(node map[string]any, segs []segment, val string) error {
	seg := segs[0]
	rest := segs[1:]

	if len(rest) == 0 {
		// Terminal: set value
		if seg.isArr {
			arr := getOrMakeSlice(node, seg.field)
			if seg.index == -1 {
				node[seg.field] = append(arr, coerce(val))
			} else {
				node[seg.field] = setSliceIndex(arr, seg.index, coerce(val))
			}
		} else {
			node[seg.field] = coerce(val)
		}
		return nil
	}

	// Non-terminal: descend
	if seg.isArr {
		arr := getOrMakeSlice(node, seg.field)
		if seg.index == -1 {
			return fmt.Errorf("append [] syntax not supported in non-terminal position")
		}
		for len(arr) <= seg.index {
			arr = append(arr, map[string]any{})
		}
		child, ok := arr[seg.index].(map[string]any)
		if !ok {
			child = map[string]any{}
		}
		if err := applySegments(child, rest, val); err != nil {
			return err
		}
		arr[seg.index] = child
		node[seg.field] = arr
	} else {
		child := getOrMakeMap(node, seg.field)
		if err := applySegments(child, rest, val); err != nil {
			return err
		}
		node[seg.field] = child
	}
	return nil
}

func getOrMakeSlice(node map[string]any, key string) []any {
	v, ok := node[key]
	if !ok {
		return []any{}
	}
	arr, ok := v.([]any)
	if !ok {
		return []any{}
	}
	return arr
}

func setSliceIndex(arr []any, index int, val any) []any {
	for len(arr) <= index {
		arr = append(arr, map[string]any{})
	}
	arr[index] = val
	return arr
}

func getOrMakeMap(node map[string]any, key string) map[string]any {
	v, ok := node[key]
	if !ok {
		return map[string]any{}
	}
	m, ok := v.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return m
}
