package moltark

import (
	"bytes"

	"gopkg.in/yaml.v3"
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
		value, err := requireStructuredValue(desiredValues, FileFormatYAML, ownedPath)
		if err != nil {
			return "", err
		}
		setStructuredValue(values, FileFormatYAML, ownedPath, value)
	}

	return renderYAMLFile(values)
}

func renderYAMLFile(values map[string]any) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(values); err != nil {
		return "", err
	}
	return ensureTrailingNewline(buf.String()), nil
}
