package moltark

import (
	"reflect"
	"testing"
)

func TestInferOwnedPathsForTOMLUsesDottedPaths(t *testing.T) {
	paths, err := inferOwnedPaths(FileFormatTOML, map[string]any{
		"default": map[string]any{
			"extend-words": map[string]any{
				"teh": "teh",
			},
		},
		"files": map[string]any{
			"extend-exclude": []any{"vendor/**"},
		},
	})
	if err != nil {
		t.Fatalf("inferOwnedPaths: %v", err)
	}

	want := []string{
		"default.extend-words.teh",
		"files.extend-exclude",
	}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("unexpected paths: got %#v want %#v", paths, want)
	}
}

func TestInferOwnedPathsForYAMLUsesJSONPointers(t *testing.T) {
	paths, err := inferOwnedPaths(FileFormatYAML, map[string]any{
		"docs": map[string]any{
			"changed-files": map[string]any{
				"any-glob-to-any-file": []any{"docs/**"},
			},
		},
	})
	if err != nil {
		t.Fatalf("inferOwnedPaths: %v", err)
	}

	want := []string{"/docs/changed-files/any-glob-to-any-file"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("unexpected paths: got %#v want %#v", paths, want)
	}
}
