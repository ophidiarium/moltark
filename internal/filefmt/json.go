package filefmt

import (
	"bytes"
	"encoding/json"

	"github.com/ophidiarium/moltark/internal/model"
)

func ParseJSONValues(raw []byte) (map[string]any, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 {
		return values, nil
	}
	if err := json.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func MutateJSONFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace([]byte(raw))) > 0 {
		parsed, err := ParseJSONValues([]byte(raw))
		if err != nil {
			return "", err
		}
		values = parsed
	}

	for _, ownedPath := range ownedPaths {
		value, err := RequireStructuredValue(desiredValues, model.FileFormatJSON, ownedPath)
		if err != nil {
			return "", err
		}
		SetStructuredValue(values, model.FileFormatJSON, ownedPath, value)
	}

	return RenderJSONFile(values)
}

func RenderJSONFile(values map[string]any) (string, error) {
	encoded, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return "", err
	}
	return EnsureTrailingNewline(string(encoded)), nil
}

func RenderJSONValue(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `"<invalid-json>"`
	}
	return string(encoded)
}
