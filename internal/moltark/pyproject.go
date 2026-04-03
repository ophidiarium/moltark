package moltark

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

func pythonProjectFileValues(project ProjectSpec) map[string]any {
	if project.Python == nil {
		return map[string]any{}
	}

	return map[string]any{
		"build-system": map[string]any{
			"requires":      append([]string(nil), project.Python.BuildSystem.Requires...),
			"build-backend": project.Python.BuildSystem.Backend,
		},
		"project": map[string]any{
			"name":            project.Name,
			"version":         project.Python.Version,
			"requires-python": project.Python.RequiresPython,
		},
		"tool": map[string]any{
			"moltark": map[string]any{
				"schema-version":   SchemaVersion,
				"template-version": project.Python.TemplateVersion,
			},
		},
	}
}

func parseTomlValues(raw []byte) (map[string]any, error) {
	values := map[string]any{}
	if err := toml.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func lookupPath(values map[string]any, path string) (any, bool) {
	current := any(values)
	for _, part := range strings.Split(path, ".") {
		nested, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = nested[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

func setPathValue(values map[string]any, path string, value any) {
	parts := strings.Split(path, ".")
	current := values
	for _, part := range parts[:len(parts)-1] {
		nested, ok := current[part].(map[string]any)
		if !ok {
			nested = map[string]any{}
			current[part] = nested
		}
		current = nested
	}
	current[parts[len(parts)-1]] = value
}

func mutateTOMLFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return renderTOMLFile(desiredValues, ownedPaths), nil
	}

	updates := make([]tomlUpdate, 0, len(ownedPaths))
	for _, ownedPath := range ownedPaths {
		table, key := splitOwnedPath(ownedPath)
		value, _ := lookupPath(desiredValues, ownedPath)
		updates = append(updates, tomlUpdate{
			Table: table,
			Key:   key,
			Value: renderTomlValue(value),
		})
	}

	out := raw
	for _, update := range updates {
		var err error
		out, err = upsertTomlKey(out, update.Table, update.Key, update.Value)
		if err != nil {
			return "", err
		}
	}

	return ensureTrailingNewline(out), nil
}

func renderTOMLFile(desiredValues map[string]any, ownedPaths []string) string {
	sections := map[string][]string{}
	tableOrder := []string{}
	for _, ownedPath := range ownedPaths {
		table, key := splitOwnedPath(ownedPath)
		if _, ok := sections[table]; !ok {
			tableOrder = append(tableOrder, table)
		}
		if !containsString(sections[table], key) {
			sections[table] = append(sections[table], key)
		}
	}
	lines := []string{}
	for _, table := range tableOrder {
		keys := sections[table]
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "["+table+"]")
		for _, key := range keys {
			ownedPath := table + "." + key
			value, _ := lookupPath(desiredValues, ownedPath)
			lines = append(lines, fmt.Sprintf("%s = %s", key, renderTomlValue(value)))
		}
	}

	return strings.Join(lines, "\n") + "\n"
}

type tomlUpdate struct {
	Table string
	Key   string
	Value string
}

func splitOwnedPath(path string) (string, string) {
	index := strings.LastIndex(path, ".")
	return path[:index], path[index+1:]
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func renderTomlValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strconv.Quote(typed)
	case int:
		return strconv.Itoa(typed)
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, strconv.Quote(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	default:
		encoded, _ := json.Marshal(typed)
		return string(encoded)
	}
}

func upsertTomlKey(raw string, table string, key string, value string) (string, error) {
	lines, newline := splitText(raw)
	keyLine := fmt.Sprintf("%s = %s", key, value)
	start, end := findTable(lines, table)
	if start == -1 {
		insertAt := len(lines)
		for insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) == "" {
			insertAt--
		}
		block := []string{}
		if insertAt > 0 {
			block = append(block, "")
		}
		block = append(block, "["+table+"]", keyLine)
		lines = insertLines(lines, insertAt, block)
		return joinLines(lines, newline), nil
	}

	keyStart, keyEnd := findKey(lines, start, end, key)
	if keyStart == -1 {
		insertAt := end
		for insertAt > start+1 && strings.TrimSpace(lines[insertAt-1]) == "" {
			insertAt--
		}
		lines = insertLines(lines, insertAt, []string{keyLine})
		return joinLines(lines, newline), nil
	}

	lines = append(lines[:keyStart], append([]string{keyLine}, lines[keyEnd:]...)...)
	return joinLines(lines, newline), nil
}

func splitText(raw string) ([]string, string) {
	newline := "\n"
	if strings.Contains(raw, "\r\n") {
		newline = "\r\n"
		raw = strings.ReplaceAll(raw, "\r\n", "\n")
	}
	raw = strings.TrimSuffix(raw, "\n")
	if raw == "" {
		return []string{}, newline
	}
	return strings.Split(raw, "\n"), newline
}

func joinLines(lines []string, newline string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, newline) + newline
}

func findTable(lines []string, table string) (int, int) {
	start := -1
	for i, line := range lines {
		name, ok := parseTableHeader(line)
		if !ok {
			continue
		}
		if name == table {
			start = i
			continue
		}
		if start != -1 {
			return start, i
		}
	}
	if start == -1 {
		return -1, -1
	}
	return start, len(lines)
}

func parseTableHeader(line string) (string, bool) {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "[") || !strings.HasSuffix(trimmed, "]") || strings.HasPrefix(trimmed, "[[") {
		return "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(trimmed, "["), "]"), true
}

func findKey(lines []string, start int, end int, key string) (int, int) {
	prefix := key + " ="
	for i := start + 1; i < end; i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, prefix) {
			return i, findValueEnd(lines, i, end)
		}
	}
	return -1, -1
}

func findValueEnd(lines []string, keyIndex int, end int) int {
	line := lines[keyIndex]
	eqIndex := strings.Index(line, "=")
	if eqIndex < 0 {
		return keyIndex + 1
	}

	depth := bracketDelta(line[eqIndex+1:])
	if depth <= 0 {
		return keyIndex + 1
	}

	for i := keyIndex + 1; i < end; i++ {
		depth += bracketDelta(lines[i])
		if depth <= 0 {
			return i + 1
		}
	}

	return end
}

func bracketDelta(text string) int {
	return strings.Count(text, "[") - strings.Count(text, "]")
}

func insertLines(lines []string, at int, inserts []string) []string {
	if at >= len(lines) {
		return append(lines, inserts...)
	}
	result := append([]string{}, lines[:at]...)
	result = append(result, inserts...)
	result = append(result, lines[at:]...)
	return result
}

func ensureTrailingNewline(raw string) string {
	if raw == "" || strings.HasSuffix(raw, "\n") {
		return raw
	}
	return raw + "\n"
}
