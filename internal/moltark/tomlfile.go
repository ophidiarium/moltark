package moltark

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

func decodeToml(raw []byte, v any) error {
	return toml.Unmarshal(raw, v)
}

func mutateTOMLFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	if strings.TrimSpace(raw) == "" {
		return renderTOMLFile(desiredValues, ownedPaths)
	}

	updates := make([]tomlUpdate, 0, len(ownedPaths))
	for _, ownedPath := range ownedPaths {
		table, key := splitOwnedPath(ownedPath)
		value, err := requireStructuredValue(desiredValues, FileFormatTOML, ownedPath)
		if err != nil {
			return "", err
		}
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

func renderTOMLFile(desiredValues map[string]any, ownedPaths []string) (string, error) {
	sections := map[string][]string{}
	rootKeys := []string{}
	tableOrder := []string{}
	for _, ownedPath := range ownedPaths {
		table, key := splitOwnedPath(ownedPath)
		if table == "" {
			if !containsString(rootKeys, key) {
				rootKeys = append(rootKeys, key)
			}
			continue
		}
		if _, ok := sections[table]; !ok {
			tableOrder = append(tableOrder, table)
		}
		if !containsString(sections[table], key) {
			sections[table] = append(sections[table], key)
		}
	}
	lines := []string{}
	for _, key := range rootKeys {
		value, err := requireStructuredValue(desiredValues, FileFormatTOML, key)
		if err != nil {
			return "", err
		}
		lines = append(lines, fmt.Sprintf("%s = %s", key, renderTomlValue(value)))
	}
	for _, table := range tableOrder {
		keys := sections[table]
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, "["+table+"]")
		for _, key := range keys {
			ownedPath := table + "." + key
			value, err := requireStructuredValue(desiredValues, FileFormatTOML, ownedPath)
			if err != nil {
				return "", err
			}
			lines = append(lines, fmt.Sprintf("%s = %s", key, renderTomlValue(value)))
		}
	}

	return strings.Join(lines, "\n") + "\n", nil
}

type tomlUpdate struct {
	Table string
	Key   string
	Value string
}

func splitOwnedPath(path string) (string, string) {
	index := strings.LastIndex(path, ".")
	if index < 0 {
		return "", path
	}
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
	case nil:
		return `""`
	case string:
		return strconv.Quote(typed)
	case bool:
		return strconv.FormatBool(typed)
	case int:
		return strconv.Itoa(typed)
	case int8, int16, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(typed).Int(), 10)
	case uint, uint8, uint16, uint32, uint64, uintptr:
		return strconv.FormatUint(reflect.ValueOf(typed).Uint(), 10)
	case float32:
		return strconv.FormatFloat(float64(typed), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	case []string:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, strconv.Quote(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case []any:
		items := make([]string, 0, len(typed))
		for _, item := range typed {
			items = append(items, renderTomlValue(item))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]any:
		return renderTomlInlineTable(typed)
	case map[string]string:
		items := make(map[string]any, len(typed))
		for key, item := range typed {
			items[key] = item
		}
		return renderTomlInlineTable(items)
	}

	valueRef := reflect.ValueOf(value)
	switch valueRef.Kind() {
	case reflect.Slice, reflect.Array:
		items := make([]string, 0, valueRef.Len())
		for i := 0; i < valueRef.Len(); i++ {
			items = append(items, renderTomlValue(valueRef.Index(i).Interface()))
		}
		return "[" + strings.Join(items, ", ") + "]"
	case reflect.Map:
		if valueRef.Type().Key().Kind() == reflect.String {
			items := make(map[string]any, valueRef.Len())
			iter := valueRef.MapRange()
			for iter.Next() {
				items[iter.Key().String()] = iter.Value().Interface()
			}
			return renderTomlInlineTable(items)
		}
	}

	return strconv.Quote(fmt.Sprint(value))
}

func renderTomlInlineTable(values map[string]any) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	items := make([]string, 0, len(keys))
	for _, key := range keys {
		items = append(items, fmt.Sprintf("%s = %s", key, renderTomlValue(values[key])))
	}
	return "{ " + strings.Join(items, ", ") + " }"
}

func upsertTomlKey(raw string, table string, key string, value string) (string, error) {
	lines, newline := splitText(raw)
	mask := buildMultilineMask(lines)
	keyLine := fmt.Sprintf("%s = %s", key, value)
	if table == "" {
		start, end := findRootTable(lines, mask)
		keyStart, keyEnd := findKey(lines, start, end, key, mask)
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
	start, end := findTable(lines, table, mask)
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

	keyStart, keyEnd := findKey(lines, start, end, key, mask)
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

// buildMultilineMask returns a boolean slice where mask[i] is true when
// line i is a continuation of a multi-line basic (""") or literal (''')
// string value. The opening key=value line itself is not masked.
func buildMultilineMask(lines []string) []bool {
	mask := make([]bool, len(lines))
	inMultiline := false
	closer := ""
	for i, line := range lines {
		if inMultiline {
			mask[i] = true
			if strings.Contains(line, closer) {
				inMultiline = false
			}
			continue
		}
		idx := assignmentIndex(line)
		if idx < 0 {
			continue
		}
		value := strings.TrimSpace(line[idx+1:])
		if strings.HasPrefix(value, `"""`) {
			if rest := value[3:]; !strings.Contains(rest, `"""`) {
				inMultiline = true
				closer = `"""`
			}
		} else if strings.HasPrefix(value, `'''`) {
			if rest := value[3:]; !strings.Contains(rest, `'''`) {
				inMultiline = true
				closer = `'''`
			}
		}
	}
	return mask
}

func findRootTable(lines []string, mask []bool) (int, int) {
	for i, line := range lines {
		if mask[i] {
			continue
		}
		if _, ok := parseTableHeader(line); ok {
			return -1, i
		}
	}
	return -1, len(lines)
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

func findTable(lines []string, table string, mask []bool) (int, int) {
	start := -1
	for i, line := range lines {
		if mask[i] {
			continue
		}
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
	if !strings.HasPrefix(trimmed, "[") || strings.HasPrefix(trimmed, "[[") {
		return "", false
	}
	end := strings.Index(trimmed, "]")
	if end <= 0 {
		return "", false
	}

	rest := strings.TrimSpace(trimmed[end+1:])
	if rest != "" && !strings.HasPrefix(rest, "#") {
		return "", false
	}

	return strings.TrimSpace(trimmed[1:end]), true
}

func findKey(lines []string, start int, end int, key string, mask []bool) (int, int) {
	for i := start + 1; i < end; i++ {
		if mask[i] {
			continue
		}
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parsedKey, ok := parseAssignmentKey(lines[i])
		if ok && parsedKey == key {
			return i, findValueEnd(lines, i, end)
		}
	}
	return -1, -1
}

func findValueEnd(lines []string, keyIndex int, end int) int {
	for i := keyIndex; i < end; i++ {
		snippet := strings.Join(lines[keyIndex:i+1], "\n")
		if isCompleteTomlAssignment(snippet) {
			return i + 1
		}
	}

	return end
}

func parseAssignmentKey(line string) (string, bool) {
	index := assignmentIndex(line)
	if index < 0 {
		return "", false
	}

	key := strings.TrimSpace(line[:index])
	if key == "" {
		return "", false
	}
	return key, true
}

func assignmentIndex(line string) int {
	inBasicString := false
	inLiteralString := false
	escaped := false
	for i := 0; i < len(line); i++ {
		ch := line[i]
		switch {
		case inBasicString:
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inBasicString = false
			}
		case inLiteralString:
			if ch == '\'' {
				inLiteralString = false
			}
		default:
			switch ch {
			case '#':
				return -1
			case '"':
				inBasicString = true
			case '\'':
				inLiteralString = true
			case '=':
				return i
			}
		}
	}
	return -1
}

func isCompleteTomlAssignment(snippet string) bool {
	var values map[string]any
	return decodeToml([]byte(snippet), &values) == nil
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
