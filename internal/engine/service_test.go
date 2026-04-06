package engine

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ophidiarium/moltark/internal/model"
	"github.com/ophidiarium/moltark/internal/testrepo"
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

	pyprojectPath := filepath.Join(root, model.PyprojectFileName)
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
	return testrepo.PrepareFixture(name)
}
