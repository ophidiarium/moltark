package filefmt

import (
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
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

	got, err := MutateTOMLFile(raw, desiredValues, []string{"project.name"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	if !strings.Contains(got, "name = \"renamed\"\n") {
		t.Fatalf("expected updated project.name, got:\n%s", got)
	}
	if !strings.Contains(got, "description = \"\"\"\nDocker [images]\n# still part of the string\n\"\"\"\n") {
		t.Fatalf("expected multiline string to be preserved, got:\n%s", got)
	}
}

func TestMutateTOMLFileSupportsRootLevelKeys(t *testing.T) {
	got, err := MutateTOMLFile("", map[string]any{"schema-version": 1}, []string{"schema-version"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	if got != "schema-version = 1\n" {
		t.Fatalf("unexpected rendered TOML: %q", got)
	}
}

func TestMutateTOMLFileRejectsMissingOwnedPaths(t *testing.T) {
	_, err := MutateTOMLFile("", map[string]any{
		"tool": map[string]any{},
	}, []string{"tool.moltark.schema-version"})
	if err == nil {
		t.Fatal("expected missing owned path to fail")
	}
	if !strings.Contains(err.Error(), `missing desired value for owned path "tool.moltark.schema-version"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMutateTOMLFileSkipsKeysInsideMultilineStrings(t *testing.T) {
	raw := `[project]
description = """
name = "not-the-real-name"
"""
name = "real"
version = "0.1.0"
`

	desiredValues := map[string]any{
		"project": map[string]any{
			"name": "updated",
		},
	}

	got, err := MutateTOMLFile(raw, desiredValues, []string{"project.name"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	if !strings.Contains(got, "name = \"updated\"\n") {
		t.Fatalf("expected updated project.name, got:\n%s", got)
	}
	if !strings.Contains(got, "name = \"not-the-real-name\"") {
		t.Fatalf("expected multiline content to be preserved, got:\n%s", got)
	}
}

func TestMutateTOMLFileSkipsTableHeadersInsideMultilineStrings(t *testing.T) {
	raw := `[project]
description = """
[build-system]
fake content
"""
name = "demo"

[build-system]
requires = ["hatchling"]
`

	desiredValues := map[string]any{
		"build-system": map[string]any{
			"requires": []string{"flit_core"},
		},
	}

	got, err := MutateTOMLFile(raw, desiredValues, []string{"build-system.requires"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	if !strings.Contains(got, `requires = ["flit_core"]`) {
		t.Fatalf("expected updated build-system.requires, got:\n%s", got)
	}
	if !strings.Contains(got, "[build-system]\nfake content") {
		t.Fatalf("expected multiline content to be preserved, got:\n%s", got)
	}
}

func TestMutateTOMLFileSkipsLiteralMultilineStrings(t *testing.T) {
	raw := `[project]
description = '''
name = "not-the-real-name"
'''
name = "real"
`

	desiredValues := map[string]any{
		"project": map[string]any{
			"name": "updated",
		},
	}

	got, err := MutateTOMLFile(raw, desiredValues, []string{"project.name"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	if !strings.Contains(got, "name = \"updated\"\n") {
		t.Fatalf("expected updated project.name, got:\n%s", got)
	}
	if !strings.Contains(got, "name = \"not-the-real-name\"") {
		t.Fatalf("expected literal multiline content to be preserved, got:\n%s", got)
	}
}

func TestMutateTOMLFileSkipsMultilineStringsInsideArrays(t *testing.T) {
	raw := `[project]
classifiers = ["""
[fake-table]
name = "not-real"
"""]
name = "demo"
`

	desiredValues := map[string]any{
		"project": map[string]any{
			"name": "updated",
		},
	}

	got, err := MutateTOMLFile(raw, desiredValues, []string{"project.name"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	// The output must be valid TOML with the correct project.name value.
	var parsed struct {
		Project struct {
			Name        string   `toml:"name"`
			Classifiers []string `toml:"classifiers"`
		} `toml:"project"`
	}
	if err := toml.Unmarshal([]byte(got), &parsed); err != nil {
		t.Fatalf("output is not valid TOML: %v\ngot:\n%s", err, got)
	}
	if parsed.Project.Name != "updated" {
		t.Fatalf("expected project.name = %q, got %q\nfull output:\n%s", "updated", parsed.Project.Name, got)
	}
	if len(parsed.Project.Classifiers) != 1 || !strings.Contains(parsed.Project.Classifiers[0], "[fake-table]") {
		t.Fatalf("expected classifiers to be preserved, got %v\nfull output:\n%s", parsed.Project.Classifiers, got)
	}
}

func TestMutateTOMLFileStopsAtArrayOfTablesBoundary(t *testing.T) {
	raw := `[tool.uv]
dev-dependencies = ["pytest"]

[[tool.uv.index]]
name = "custom"
url = "https://example.com/simple"
`

	desiredValues := map[string]any{
		"tool": map[string]any{
			"uv": map[string]any{
				"managed": true,
			},
		},
	}

	got, err := MutateTOMLFile(raw, desiredValues, []string{"tool.uv.managed"})
	if err != nil {
		t.Fatalf("MutateTOMLFile: %v", err)
	}

	// New key must land inside [tool.uv], before [[tool.uv.index]].
	uvIdx := strings.Index(got, "[tool.uv]")
	managedIdx := strings.Index(got, "managed = true")
	arrayIdx := strings.Index(got, "[[tool.uv.index]]")

	if managedIdx < 0 {
		t.Fatalf("expected managed key to be inserted, got:\n%s", got)
	}
	if arrayIdx < 0 {
		t.Fatalf("expected array-of-tables to be preserved, got:\n%s", got)
	}
	if managedIdx < uvIdx || managedIdx > arrayIdx {
		t.Fatalf("expected managed key between [tool.uv] and [[tool.uv.index]], got:\n%s", got)
	}
	if !strings.Contains(got, `name = "custom"`) {
		t.Fatalf("expected array-of-tables content to be preserved, got:\n%s", got)
	}
}

func TestRenderTomlValueRendersValidInlineTables(t *testing.T) {
	rendered, err := RenderTomlValue(map[string]any{
		"enabled": true,
		"tool":    "uv",
	})
	if err != nil {
		t.Fatalf("RenderTomlValue: %v", err)
	}

	values := map[string]any{}
	if err := toml.Unmarshal([]byte("value = "+rendered+"\n"), &values); err != nil {
		t.Fatalf("rendered value must be valid TOML: %v", err)
	}
}

func TestRenderTomlValueRejectsUnsupportedTypes(t *testing.T) {
	type custom struct{ X int }
	_, err := RenderTomlValue(custom{X: 1})
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
	if !strings.Contains(err.Error(), "unsupported TOML value type") {
		t.Fatalf("unexpected error: %v", err)
	}
}
