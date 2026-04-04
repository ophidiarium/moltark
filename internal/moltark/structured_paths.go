package moltark

import (
	"fmt"
	"sort"
	"strings"
)

func lookupStructuredValue(values map[string]any, format string, path string) (any, bool) {
	current := any(values)
	for _, part := range structuredPathParts(format, path) {
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

func requireStructuredValue(values map[string]any, format string, path string) (any, error) {
	value, ok := lookupStructuredValue(values, format, path)
	if !ok {
		return nil, fmt.Errorf("missing desired value for owned path %q", path)
	}
	return value, nil
}

func setStructuredValue(values map[string]any, format string, path string, value any) {
	parts := structuredPathParts(format, path)
	if len(parts) == 0 {
		return
	}

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

func structuredPathParts(format string, path string) []string {
	switch format {
	case FileFormatJSON, FileFormatYAML:
		return jsonPointerParts(path)
	default:
		return strings.Split(path, ".")
	}
}

func jsonPointerParts(pointer string) []string {
	if pointer == "" {
		return nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return []string{pointer}
	}

	rawParts := strings.Split(pointer[1:], "/")
	parts := make([]string, 0, len(rawParts))
	for _, part := range rawParts {
		part = strings.ReplaceAll(part, "~1", "/")
		part = strings.ReplaceAll(part, "~0", "~")
		parts = append(parts, part)
	}
	return parts
}

func encodeJSONPointer(parts ...string) string {
	if len(parts) == 0 {
		return ""
	}

	encoded := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.ReplaceAll(part, "~", "~0")
		part = strings.ReplaceAll(part, "/", "~1")
		encoded = append(encoded, part)
	}
	return "/" + strings.Join(encoded, "/")
}

func inferOwnedPaths(format string, value any) ([]string, error) {
	switch format {
	case FileFormatJSON:
		root, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("json file values must be an object")
		}
		return inferJSONOwnedPaths(root, nil), nil
	case FileFormatYAML:
		root, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("yaml file values must be an object")
		}
		return inferJSONOwnedPaths(root, nil), nil
	case FileFormatTOML:
		root, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("toml file values must be an object")
		}
		if err := validateNoLiteralDotKeys(root, nil); err != nil {
			return nil, err
		}
		return inferDottedOwnedPaths(root, nil), nil
	default:
		return nil, fmt.Errorf("unsupported file format %q", format)
	}
}

func inferJSONOwnedPaths(value any, prefix []string) []string {
	object, ok := value.(map[string]any)
	if !ok {
		return []string{encodeJSONPointer(prefix...)}
	}
	if len(object) == 0 {
		return []string{encodeJSONPointer(prefix...)}
	}

	paths := []string{}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		nested := object[key]
		paths = append(paths, inferJSONOwnedPaths(nested, append(append([]string{}, prefix...), key))...)
	}
	return paths
}

func inferDottedOwnedPaths(value any, prefix []string) []string {
	object, ok := value.(map[string]any)
	if !ok {
		return []string{strings.Join(prefix, ".")}
	}
	if len(object) == 0 {
		return []string{strings.Join(prefix, ".")}
	}

	paths := []string{}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		nested := object[key]
		paths = append(paths, inferDottedOwnedPaths(nested, append(append([]string{}, prefix...), key))...)
	}
	return paths
}

func validateNoLiteralDotKeys(value any, prefix []string) error {
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if strings.Contains(key, ".") {
			location := "<root>"
			if len(prefix) > 0 {
				location = strings.Join(prefix, ".")
			}
			return fmt.Errorf("toml key %q under %s contains a literal dot; use nested maps for TOML paths", key, location)
		}
		if err := validateNoLiteralDotKeys(object[key], append(append([]string{}, prefix...), key)); err != nil {
			return err
		}
	}

	return nil
}
