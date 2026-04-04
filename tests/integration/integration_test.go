package integration_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gkampitakis/go-snaps/snaps"
	"github.com/ophidiarium/moltark/internal/testutil"
)

var integrationSnaps = snaps.WithConfig(
	snaps.Dir(testutil.RepoPath("tests/integration/__snapshots__")),
)

func TestBootstrapAndReapply(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "demo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("create repo dir: %v", err)
	}

	initResult := testutil.RunCLI(t, dir, "", "init")
	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")

	if initResult.ExitCode != 0 || planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: init=%d plan=%d apply=%d replan=%d", initResult.ExitCode, planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark init", initResult),
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestPreserveUserManagedDependencies(t *testing.T) {
	dir := testutil.CopyFixture(t, "bare_python_repo_with_uv_deps")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	doctorResult := testutil.RunCLI(t, dir, "", "doctor")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || doctorResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d doctor=%d", planResult.ExitCode, applyResult.ExitCode, doctorResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark doctor", doctorResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestDriftDetection(t *testing.T) {
	dir := testutil.CopyFixture(t, "drifted_owned_fields")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	doctorResult := testutil.RunCLI(t, dir, "", "doctor")

	if planResult.ExitCode != 0 || doctorResult.ExitCode != 1 {
		t.Fatalf("unexpected exit codes: plan=%d doctor=%d", planResult.ExitCode, doctorResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark doctor", doctorResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestConflictSurfacing(t *testing.T) {
	dir := testutil.CopyFixture(t, "conflicting_owned_fields")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")

	if planResult.ExitCode != 1 || applyResult.ExitCode != 1 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d", planResult.ExitCode, applyResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestTemplateVersionUpgrade(t *testing.T) {
	dir := testutil.CopyFixture(t, "upgraded_template_v1_to_v2")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestUVWorkspaceWithParentRelativeMembers(t *testing.T) {
	dir := testutil.CopyFixture(t, "uv_workspace_with_nested_root")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")
	showResult := testutil.RunCLI(t, dir, "", "show")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 || showResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d show=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode, showResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderCommand("moltark show", showResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestCoreTasksAndTriggerBindings(t *testing.T) {
	dir := testutil.CopyFixture(t, "core_tasks_with_triggers")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")
	showResult := testutil.RunCLI(t, dir, "", "show")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 || showResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d show=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode, showResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderCommand("moltark show", showResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestGoLintVscodeComponentOnly(t *testing.T) {
	dir := testutil.CopyFixture(t, "go_lint_vscode_component")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")
	showResult := testutil.RunCLI(t, dir, "", "show")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 || showResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d show=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode, showResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderCommand("moltark show", showResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestCoreStructuredFilePrimitives(t *testing.T) {
	dir := testutil.CopyFixture(t, "core_structured_file_primitives")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")
	showResult := testutil.RunCLI(t, dir, "", "show")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 || showResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d show=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode, showResult.ExitCode)
	}

	integrationSnaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderCommand("moltark show", showResult),
		testutil.RenderRepoState(t, dir),
	))
}

func TestOpaqueRootProjectWithPythonChild(t *testing.T) {
	dir := testutil.CopyFixture(t, "opaque_root_with_python_child")

	planResult := testutil.RunCLI(t, dir, "", "plan")
	applyResult := testutil.RunCLI(t, dir, "", "apply", "-auto-approve")
	replanResult := testutil.RunCLI(t, dir, "", "plan")
	showResult := testutil.RunCLI(t, dir, "", "show")

	if planResult.ExitCode != 0 || applyResult.ExitCode != 0 || replanResult.ExitCode != 0 || showResult.ExitCode != 0 {
		t.Fatalf("unexpected exit codes: plan=%d apply=%d replan=%d show=%d", planResult.ExitCode, applyResult.ExitCode, replanResult.ExitCode, showResult.ExitCode)
	}

	snaps.MatchSnapshot(t, renderSession(
		testutil.RenderCommand("moltark plan", planResult),
		testutil.RenderCommand("moltark apply -auto-approve", applyResult),
		testutil.RenderCommand("moltark plan", replanResult),
		testutil.RenderCommand("moltark show", showResult),
		testutil.RenderRepoState(t, dir),
	))
}

func renderSession(parts ...string) string {
	return strings.Join(parts, "\n---\n")
}
