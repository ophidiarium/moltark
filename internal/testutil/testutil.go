package testutil

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/ophidiarium/moltark/internal/cliapp"
	"github.com/ophidiarium/moltark/internal/moltark"
	"github.com/ophidiarium/moltark/internal/testrepo"
)

func RepoPath(path string) string {
	return testrepo.RepoPath(path)
}

type CommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

func FixturePath(name string) string {
	return testrepo.FixturePath(name)
}

func PrepareFixture(name string) (string, error) {
	return prepareFixture(name, os.MkdirTemp, CopyDir, os.RemoveAll)
}

func prepareFixture(name string, mkTemp func(string, string) (string, error), copyDir func(string, string) error, removeAll func(string) error) (string, error) {
	dir, err := mkTemp("", "moltark-fixture-*")
	if err != nil {
		return "", err
	}
	if err := copyDir(FixturePath(name), dir); err != nil {
		_ = removeAll(dir)
		return "", err
	}
	return dir, nil
}

func CopyFixture(t *testing.T, name string) string {
	t.Helper()

	dir, err := PrepareFixture(name)
	if err != nil {
		t.Fatalf("prepare fixture %s: %v", name, err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(dir)
	})

	return dir
}

func CopyDir(src string, dst string) error { return testrepo.CopyDir(src, dst) }

func RunCLIInDir(dir string, stdin string, args ...string) CommandResult {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := cliapp.Run(cliapp.Config{
		Args:       args,
		WorkingDir: dir,
		Stdin:      strings.NewReader(stdin),
		Stdout:     &stdout,
		Stderr:     &stderr,
	})

	return CommandResult{
		ExitCode: exitCode,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
	}
}

func RunCLI(t *testing.T, dir string, stdin string, args ...string) CommandResult {
	t.Helper()
	return RunCLIInDir(dir, stdin, args...)
}

func RenderCommand(label string, result CommandResult) string {
	return fmt.Sprintf(
		"$ %s\nexit: %d\nstdout:\n%sstderr:\n%s",
		label,
		result.ExitCode,
		indentBlock(result.Stdout),
		indentBlock(result.Stderr),
	)
}

func RenderRepoState(t *testing.T, root string) string {
	t.Helper()

	paths := []string{
		moltark.ProjectSpecFileName,
		".gitattributes",
		".moltark/state.json",
	}
	structuredFiles := map[string]struct{}{}
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != root && d.Name() == ".moltark" {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		switch filepath.Ext(d.Name()) {
		case ".json", ".toml", ".yaml", ".yml":
			if rel == ".moltark/state.json" {
				return nil
			}
			structuredFiles[rel] = struct{}{}
		}
		return nil
	}); err != nil {
		t.Fatalf("walk repo state: %v", err)
	}
	var structured []string
	for path := range structuredFiles {
		structured = append(structured, path)
	}
	sort.Strings(structured)
	paths = append(paths[:1], append(structured, paths[1:]...)...)

	var b strings.Builder
	for i, path := range paths {
		if i > 0 {
			b.WriteString("\n")
		}

		b.WriteString("== ")
		b.WriteString(path)
		b.WriteString(" ==\n")

		body, err := os.ReadFile(filepath.Join(root, path))
		if err != nil {
			if os.IsNotExist(err) {
				b.WriteString("<absent>\n")
				continue
			}
			t.Fatalf("read %s: %v", path, err)
		}

		b.Write(body)
		if len(body) == 0 || body[len(body)-1] != '\n' {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func indentBlock(text string) string {
	if text == "" {
		return "  <empty>\n"
	}

	lines := strings.Split(text, "\n")
	var b strings.Builder
	for _, line := range lines {
		if line == "" {
			continue
		}
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}
	if b.Len() == 0 {
		return "  <empty>\n"
	}
	return b.String()
}
