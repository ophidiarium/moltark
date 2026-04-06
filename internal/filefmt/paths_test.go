package filefmt

import (
	"reflect"
	"strings"
	"testing"

	"github.com/ophidiarium/moltark/internal/model"
)

func TestInferOwnedPathsForTOMLUsesDottedPaths(t *testing.T) {
	paths, err := InferOwnedPaths(model.FileFormatTOML, map[string]any{
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
		t.Fatalf("InferOwnedPaths: %v", err)
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
	paths, err := InferOwnedPaths(model.FileFormatYAML, map[string]any{
		"docs": map[string]any{
			"changed-files": map[string]any{
				"any-glob-to-any-file": []any{"docs/**"},
			},
		},
	})
	if err != nil {
		t.Fatalf("InferOwnedPaths: %v", err)
	}

	want := []string{"/docs/changed-files/any-glob-to-any-file"}
	if !reflect.DeepEqual(paths, want) {
		t.Fatalf("unexpected paths: got %#v want %#v", paths, want)
	}
}

func TestInferOwnedPathsForTOMLRejectsLiteralDotKeys(t *testing.T) {
	_, err := InferOwnedPaths(model.FileFormatTOML, map[string]any{
		"tool.ruff": map[string]any{
			"line-length": 100,
		},
	})
	if err == nil {
		t.Fatal("expected dotted TOML keys to be rejected")
	}
	if !strings.Contains(err.Error(), `toml key "tool.ruff"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
