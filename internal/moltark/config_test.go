package moltark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDesiredModelRejectsModuleVersions(t *testing.T) {
	root := t.TempDir()
	content := `python = use("moltark/python", version = "v0.1.0")

python.python_project(
    name = "demo",
    version = "0.1.0",
    requires_python = ">=3.12",
)
`
	if err := os.WriteFile(filepath.Join(root, ProjectSpecFileName), []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", ProjectSpecFileName, err)
	}

	_, err := LoadDesiredModel(root)
	if err == nil {
		t.Fatal("expected versioned use() to fail")
	}
	if !strings.Contains(err.Error(), "module version selection is not implemented yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}
