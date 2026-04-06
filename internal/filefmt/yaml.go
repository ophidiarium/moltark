package filefmt

import (
	"bytes"

	"github.com/ophidiarium/moltark/internal/model"
	"gopkg.in/yaml.v3"
)

func ParseYAMLValues(raw []byte) (map[string]any, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace(raw)) == 0 {
		return values, nil
	}
	if err := yaml.Unmarshal(raw, &values); err != nil {
		return nil, err
	}
	return values, nil
}

func MutateYAMLFile(raw string, desiredValues map[string]any, ownedPaths []string) (string, error) {
	values := map[string]any{}
	if len(bytes.TrimSpace([]byte(raw))) > 0 {
		parsed, err := ParseYAMLValues([]byte(raw))
		if err != nil {
			return "", err
		}
		values = parsed
	}

	for _, ownedPath := range ownedPaths {
		value, err := RequireStructuredValue(desiredValues, model.FileFormatYAML, ownedPath)
		if err != nil {
			return "", err
		}
		SetStructuredValue(values, model.FileFormatYAML, ownedPath, value)
	}

	return RenderYAMLFile(values)
}

func RenderYAMLFile(values map[string]any) (string, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(values); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	return EnsureTrailingNewline(buf.String()), nil
}
