package moltark

import (
	"bytes"

	yaml "github.com/goccy/go-yaml"
)

func parseYAMLValues(raw []byte) (map[string]any, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 {
		return values, nil
	}
	if err := yaml.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func mutateYAMLFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace([]byte(raw))) > 0 {
		parsed, err := parseYAMLValues([]byte(raw))
		if err != nil {
			return "", err
		}
		values = parsed
	}

	for _, ownedPath := range ownedPaths {
		value, _ := lookupStructuredValue(desiredValues, FileFormatYAML, ownedPath)
		setStructuredValue(values, FileFormatYAML, ownedPath, value)
	}

	return renderYAMLFile(values)
}

func renderYAMLFile(values map[string]any) (string, error) {
	encoded, err := yaml.Marshal(values)
	if err != nil {
		return "", err
	}
	return ensureTrailingNewline(string(encoded)), nil
}
