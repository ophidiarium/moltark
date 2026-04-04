package testrepo

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
)

var repoRoot = detectRepoRoot()

func detectRepoRoot() string {
	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		if workspace := os.Getenv("TEST_WORKSPACE"); workspace != "" {
			if root := findRepoRoot(filepath.Join(srcDir, workspace)); root != "" {
				return root
			}
		}
	}

	if wd, err := os.Getwd(); err == nil {
		if root := findRepoRoot(wd); root != "" {
			return root
		}
	}

	_, filename, _, _ := runtime.Caller(0)
	if filename != "" {
		if !filepath.IsAbs(filename) {
			if wd, err := os.Getwd(); err == nil {
				filename = filepath.Join(wd, filename)
			}
		}
		if root := findRepoRoot(filepath.Dir(filename)); root != "" {
			return root
		}
	}

	return "."
}

func findRepoRoot(start string) string {
	if start == "" {
		return ""
	}

	dir := filepath.Clean(start)
	for {
		if isRepoRoot(dir) {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

func isRepoRoot(dir string) bool {
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return false
	}
	if _, err := os.Stat(filepath.Join(dir, "MODULE.bazel")); err != nil {
		return false
	}
	return true
}

func RepoPath(path string) string {
	rel := filepath.FromSlash(path)

	if filepath.IsAbs(rel) {
		return rel
	}

	if repoRoot != "" {
		candidate := filepath.Join(repoRoot, rel)
		if _, err := os.Stat(candidate); err == nil {
			return filepath.Clean(candidate)
		}
	}

	if srcDir := os.Getenv("TEST_SRCDIR"); srcDir != "" {
		candidates := []string{filepath.Join(srcDir, rel)}
		matches, err := filepath.Glob(filepath.Join(srcDir, "*", rel))
		if err == nil {
			candidates = append(candidates, matches...)
		}
		for _, candidate := range candidates {
			if _, err := os.Stat(candidate); err == nil {
				return filepath.Clean(candidate)
			}
		}
	}

	if repoRoot != "" {
		return filepath.Join(repoRoot, rel)
	}
	return rel
}

func FixturePath(name string) string {
	return RepoPath(filepath.ToSlash(filepath.Join("tests", "fixtures", name)))
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

func CopyDir(src string, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		info, err := d.Info()
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

		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
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
