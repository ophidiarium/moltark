package moltark

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadStateRejectsIncompatibleSchemaVersion(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, StateDirName), 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	body := `{
  "schema_version": 999,
  "template_version": "python/v2",
  "managed_files": [],
  "last_applied_model": {
    "projects": [],
    "components": []
  }
}
`
	if err := os.WriteFile(statePath(root), []byte(body), 0o644); err != nil {
		t.Fatalf("write state file: %v", err)
	}

	_, err := loadState(root)
	if err == nil {
		t.Fatal("expected schema-version mismatch to fail")
	}
	if !strings.Contains(err.Error(), "unsupported schema_version 999") {
		t.Fatalf("unexpected error: %v", err)
	}
}
