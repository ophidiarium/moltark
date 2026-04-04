package moltark

import (
	"strings"
	"testing"
)

func TestMutateYAMLFileMergesOwnedPathsIntoExistingDocument(t *testing.T) {
	raw := `docs:
  changed-files:
    any-glob-to-any-file:
      - docs/**
keep: true
`

	desiredValues := map[string]any{
		"docs": map[string]any{
			"changed-files": map[string]any{
				"any-glob-to-any-file": []any{"docs/**", "README.md"},
			},
		},
	}

	got, err := mutateYAMLFile(raw, desiredValues, []string{"/docs/changed-files/any-glob-to-any-file"})
	if err != nil {
		t.Fatalf("mutateYAMLFile: %v", err)
	}

	if !strings.Contains(got, "- README.md\n") {
		t.Fatalf("expected updated yaml list, got:\n%s", got)
	}
	if !strings.Contains(got, "keep: true\n") {
		t.Fatalf("expected unrelated keys to be preserved, got:\n%s", got)
	}
}
