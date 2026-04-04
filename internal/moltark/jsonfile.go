package moltark

import (
	"bytes"
	"encoding/json"
)

func parseJSONValues(raw []byte) (map[string]any, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 {
		return values, nil
	}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func mutateJSONFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace([]byte(raw))) > 0 {
		parsed, err := parseJSONValues([]byte(raw))
		if err != nil {
			return "", err
		}
		values = parsed
	}

	for _, ownedPath := range ownedPaths {
		value, err := requireStructuredValue(desiredValues, FileFormatJSON, ownedPath)
		if err != nil {
			return "", err
		}
		setStructuredValue(values, FileFormatJSON, ownedPath, value)
	}

	return renderJSONFile(values)
}

func renderJSONFile(values map[string]any) (string, error) {
	encoded, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return "", err
	}
	return ensureTrailingNewline(string(encoded)), nil
}

func renderJSONValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `"<invalid-json>"`
	}
	return string(encoded)
}
