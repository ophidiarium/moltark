package moltark

import (
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
)

func TestMutateTOMLFileUpdatesCompactAssignmentsAndPreservesMultilineStrings(t *testing.T) {
	raw := `[project]
name="demo"
description = """
Docker [images]
# still part of the string
"""
version = "0.1.0"
`

	desiredValues := map[string]any{
		"project": map[string]any{
			"name": "renamed",
		},
	}

	got, err := mutateTOMLFile(raw, desiredValues, []string{"project.name"})
	if err != nil {
		t.Fatalf("mutateTOMLFile: %v", err)
	}

	if !strings.Contains(got, "name = \"renamed\"\n") {
		t.Fatalf("expected updated project.name, got:\n%s", got)
	}
	if !strings.Contains(got, "description = \"\"\"\nDocker [images]\n# still part of the string\n\"\"\"\n") {
		t.Fatalf("expected multiline string to be preserved, got:\n%s", got)
	}
}

func TestMutateTOMLFileSupportsRootLevelKeys(t *testing.T) {
	got, err := mutateTOMLFile("", map[string]any{"schema-version": 1}, []string{"schema-version"})
	if err != nil {
		t.Fatalf("mutateTOMLFile: %v", err)
	}

	if got != "schema-version = 1\n" {
		t.Fatalf("unexpected rendered TOML: %q", got)
	}
}

func TestMutateTOMLFileRejectsMissingOwnedPaths(t *testing.T) {
	_, err := mutateTOMLFile("", map[string]any{
		"tool": map[string]any{},
	}, []string{"tool.moltark.schema-version"})
	if err == nil {
		t.Fatal("expected missing owned path to fail")
	}
	if !strings.Contains(err.Error(), `missing desired value for owned path "tool.moltark.schema-version"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRenderTomlValueRendersValidInlineTables(t *testing.T) {
	rendered := renderTomlValue(map[string]any{
		"enabled": true,
		"tool":    "uv",
	})

	values := map[string]any{}
	if err := toml.Unmarshal([]byte("value = "+rendered+"\n"), &values); err != nil {
		t.Fatalf("rendered value must be valid TOML: %v", err)
	}
}
