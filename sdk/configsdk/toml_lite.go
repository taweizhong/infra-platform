package configsdk

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

func decodeTOML(content string, out any) error {
	m, err := parseTOML(content)
	if err != nil {
		return err
	}
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}

func encodeTOML(m map[string]any) (string, error) {
	var b strings.Builder
	if err := writeTOMLMap(&b, nil, m); err != nil {
		return "", err
	}
	return b.String(), nil
}

func parseTOML(content string) (map[string]any, error) {
	root := map[string]any{}
	current := root
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section := strings.TrimSpace(line[1 : len(line)-1])
			parts := strings.Split(section, ".")
			current = root
			for _, p := range parts {
				p = strings.TrimSpace(p)
				n, ok := current[p].(map[string]any)
				if !ok {
					n = map[string]any{}
					current[p] = n
				}
				current = n
			}
			continue
		}
		idx := strings.Index(line, "=")
		if idx <= 0 {
			return nil, fmt.Errorf("invalid toml line: %s", line)
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		parsed, err := parseValue(v)
		if err != nil {
			return nil, err
		}
		current[k] = parsed
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return root, nil
}

func parseValue(v string) (any, error) {
	if strings.HasPrefix(v, `"`) && strings.HasSuffix(v, `"`) {
		return strings.Trim(v, `"`), nil
	}
	if v == "true" {
		return true, nil
	}
	if v == "false" {
		return false, nil
	}
	if strings.HasPrefix(v, "[") && strings.HasSuffix(v, "]") {
		inner := strings.TrimSpace(v[1 : len(v)-1])
		if inner == "" {
			return []any{}, nil
		}
		parts := splitCommaAware(inner)
		arr := make([]any, 0, len(parts))
		for _, p := range parts {
			pv, err := parseValue(strings.TrimSpace(p))
			if err != nil {
				return nil, err
			}
			arr = append(arr, pv)
		}
		return arr, nil
	}
	if i, err := strconv.ParseInt(v, 10, 64); err == nil {
		return i, nil
	}
	if f, err := strconv.ParseFloat(v, 64); err == nil {
		return f, nil
	}
	return v, nil
}

func splitCommaAware(s string) []string {
	var out []string
	var buf bytes.Buffer
	inStr := false
	for _, r := range s {
		switch r {
		case '"':
			inStr = !inStr
			buf.WriteRune(r)
		case ',':
			if inStr {
				buf.WriteRune(r)
			} else {
				out = append(out, buf.String())
				buf.Reset()
			}
		default:
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		out = append(out, buf.String())
	}
	return out
}

func writeTOMLMap(b *strings.Builder, path []string, m map[string]any) error {
	keys := sortedKeys(m)
	for _, k := range keys {
		v := m[k]
		if _, ok := v.(map[string]any); ok {
			continue
		}
		line, err := encodeKV(k, v)
		if err != nil {
			return err
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	for _, k := range keys {
		child, ok := m[k].(map[string]any)
		if !ok {
			continue
		}
		sub := append(append([]string{}, path...), k)
		if b.Len() > 0 {
			b.WriteString("\n")
		}
		b.WriteString("[")
		b.WriteString(strings.Join(sub, "."))
		b.WriteString("]\n")
		if err := writeTOMLMap(b, sub, child); err != nil {
			return err
		}
	}
	return nil
}

func encodeKV(k string, v any) (string, error) {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("%s = \"%s\"", k, t), nil
	case bool:
		return fmt.Sprintf("%s = %t", k, t), nil
	case int:
		return fmt.Sprintf("%s = %d", k, t), nil
	case int64:
		return fmt.Sprintf("%s = %d", k, t), nil
	case float64:
		return fmt.Sprintf("%s = %v", k, t), nil
	case []any:
		parts := make([]string, 0, len(t))
		for _, item := range t {
			p, err := encodeArrayVal(item)
			if err != nil {
				return "", err
			}
			parts = append(parts, p)
		}
		return fmt.Sprintf("%s = [%s]", k, strings.Join(parts, ", ")), nil
	default:
		return "", fmt.Errorf("unsupported value type %T", v)
	}
}

func encodeArrayVal(v any) (string, error) {
	switch t := v.(type) {
	case string:
		return fmt.Sprintf("\"%s\"", t), nil
	case bool:
		return fmt.Sprintf("%t", t), nil
	case int:
		return fmt.Sprintf("%d", t), nil
	case int64:
		return fmt.Sprintf("%d", t), nil
	case float64:
		return fmt.Sprintf("%v", t), nil
	default:
		return "", fmt.Errorf("unsupported array value type %T", v)
	}
}

func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
