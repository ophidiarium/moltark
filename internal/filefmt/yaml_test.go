package filefmt

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

	got, err := MutateYAMLFile(raw, desiredValues, []string{"/docs/changed-files/any-glob-to-any-file"})
	if err != nil {
		t.Fatalf("MutateYAMLFile: %v", err)
	}

	if !strings.Contains(got, "- README.md\n") {
		t.Fatalf("expected updated yaml list, got:\n%s", got)
	}
	if !strings.Contains(got, "keep: true\n") {
		t.Fatalf("expected unrelated keys to be preserved, got:\n%s", got)
	}
}

func TestMutateYAMLFileRejectsMissingOwnedPaths(t *testing.T) {
	_, err := MutateYAMLFile("{}", map[string]any{
		"docs": map[string]any{},
	}, []string{"/docs/changed-files/any-glob-to-any-file"})
	if err == nil {
		t.Fatal("expected missing owned path to fail")
	}
	if !strings.Contains(err.Error(), `missing desired value for owned path "/docs/changed-files/any-glob-to-any-file"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
