package moltark

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestApplyReplansAgainstFreshFileBodies(t *testing.T) {
	root, err := prepareFixture("upgraded_template_v1_to_v2")
	if err != nil {
		t.Fatalf("prepare fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(root)
	})

	service := NewService()
	plan, err := service.Plan(root)
	if err != nil {
		t.Fatalf("plan: %v", err)
	}

	pyprojectPath := filepath.Join(root, PyprojectFileName)
	raw, err := os.ReadFile(pyprojectPath)
	if err != nil {
		t.Fatalf("read pyproject: %v", err)
	}

	updatedRaw := strings.Replace(
		string(raw),
		"[project]\n",
		"[project]\ndependencies = [\n    \"httpx>=0.27.0\",\n]\n",
		1,
	)
	if err := os.WriteFile(pyprojectPath, []byte(updatedRaw), 0o644); err != nil {
		t.Fatalf("write updated pyproject: %v", err)
	}

	result, err := service.Apply(root, plan)
	if err != nil {
		t.Fatalf("apply: %v", err)
	}
	if len(result.Wrote) == 0 {
		t.Fatal("expected apply to write managed files")
	}

	body, err := os.ReadFile(pyprojectPath)
	if err != nil {
		t.Fatalf("read final pyproject: %v", err)
	}

	if !strings.Contains(string(body), "\"httpx>=0.27.0\"") {
		t.Fatalf("expected user-managed dependency to be preserved, got:\n%s", body)
	}
	if !strings.Contains(string(body), "template-version = \"python/v2\"") {
		t.Fatalf("expected managed template version update, got:\n%s", body)
	}
}

func prepareFixture(name string) (string, error) {
	root, err := os.MkdirTemp("", "moltark-fixture-*")
	if err != nil {
		return "", err
	}
	src := filepath.Join(repoRoot(), "tests", "fixtures", name)
	if err := copyDir(src, root); err != nil {
		_ = os.RemoveAll(root)
		return "", err
	}
	return root, nil
}

func copyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.Create(target)
		if err != nil {
			return err
		}
		defer out.Close()

		if _, err := io.Copy(out, in); err != nil {
			return err
		}
		return out.Close()
	})
}

func repoRoot() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return filepath.Clean(".")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
