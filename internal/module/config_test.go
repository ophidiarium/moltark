package module

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ophidiarium/moltark/internal/model"
)

func TestNormalizeProjectPathNormalizesBackslashes(t *testing.T) {
	got, err := normalizeProjectPath(`sub\dir`)
	if err != nil {
		t.Fatalf("normalizeProjectPath(%q): unexpected error: %v", `sub\dir`, err)
	}
	if got != "sub/dir" {
		t.Errorf("normalizeProjectPath(%q) = %q, want %q", `sub\dir`, got, "sub/dir")
	}
}

func TestNormalizeProjectPathRejectsEscapes(t *testing.T) {
	cases := []string{
		"..",
		"../outside",
		"pkg/../../outside",
		`pkg\..\..\outside`,
		`pkg\..\..\..\outside`,
	}
	for _, input := range cases {
		_, err := normalizeProjectPath(input)
		if err == nil {
			t.Errorf("normalizeProjectPath(%q): expected error, got nil", input)
			continue
		}
		if !strings.Contains(err.Error(), "must not escape") {
			t.Errorf("normalizeProjectPath(%q): got %q, want escape error", input, err)
		}
	}
}

func TestLoadDesiredModelRejectsModuleVersions(t *testing.T) {
	root := t.TempDir()
	content := `python = use("moltark/python", version = "v0.1.0")

python.python_project(
    name = "demo",
    version = "0.1.0",
    requires_python = ">=3.12",
)
`
	if err := os.WriteFile(filepath.Join(root, model.ProjectSpecFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", model.ProjectSpecFileName, err)
	}

	_, err := LoadDesiredModel(root)
	if err == nil {
		t.Fatal("expected versioned use() to fail")
	}
	if !strings.Contains(err.Error(), "module version selection is not implemented yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}
