package moltark

import (
	"strings"
	"testing"
)

func TestMutateJSONFileMergesOwnedPathsIntoExistingDocument(t *testing.T) {
	raw := `{
  "vscode": {
    "extensions": [
      "golang.go"
    ]
  },
  "keep": true
}
`

	desiredValues := map[string]any{
		"vscode": map[string]any{
			"extensions": []any{"golang.go", "charliermarsh.ruff"},
		},
	}

	got, err := mutateJSONFile(raw, desiredValues, []string{"/vscode/extensions"})
	if err != nil {
		t.Fatalf("mutateJSONFile: %v", err)
	}

	if !strings.Contains(got, `"charliermarsh.ruff"`) {
		t.Fatalf("expected updated json list, got:\n%s", got)
	}
	if !strings.Contains(got, `"keep": true`) {
		t.Fatalf("expected unrelated keys to be preserved, got:\n%s", got)
	}
}

func TestMutateJSONFileRejectsMissingOwnedPaths(t *testing.T) {
	_, err := mutateJSONFile("{}", map[string]any{
		"vscode": map[string]any{},
	}, []string{"/vscode/extensions"})
	if err == nil {
		t.Fatal("expected missing owned path to fail")
	}
	if !strings.Contains(err.Error(), `missing desired value for owned path "/vscode/extensions"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
