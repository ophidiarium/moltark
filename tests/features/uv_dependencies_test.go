package features_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/ophidiarium/moltark/internal/testutil"
)

type uvFeature struct {
	dir   string
	plan  testutil.CommandResult
	apply testutil.CommandResult
}

func (f *uvFeature) aPythonRepositoryBootstrappedByMoltark() error {
	dir, err := testutil.PrepareFixture("bare_python_repo_with_uv_deps")
	if err != nil {
		return err
	}
	f.dir = dir
	return nil
}

func (f *uvFeature) aRepositoryDeclaringANestedUVWorkspace() error {
	dir, err := testutil.PrepareFixture("uv_workspace_with_nested_root")
	if err != nil {
		return err
	}
	f.dir = dir
	return nil
}

func (f *uvFeature) moltarkPlanIsExecuted() error {
	f.plan = testutil.RunCLIInDir(f.dir, "", "plan")
	if f.plan.ExitCode != 0 {
		return fmt.Errorf("plan failed with exit code %d", f.plan.ExitCode)
	}
	return nil
}

func (f *uvFeature) noDependencyDriftIsReported() error {
	if !strings.Contains(f.plan.Stdout, "no change to dependencies (user-managed)") {
		return fmt.Errorf("expected dependency preservation message in plan output")
	}
	if strings.Contains(f.plan.Stdout, "drift detected in project.dependencies") {
		return fmt.Errorf("dependency drift should not be reported")
	}
	return nil
}

func (f *uvFeature) moltarkApplyMakesNoDependencyChanges() error {
	f.apply = testutil.RunCLIInDir(f.dir, "", "apply", "-auto-approve")
	if f.apply.ExitCode != 0 {
		return fmt.Errorf("apply failed with exit code %d", f.apply.ExitCode)
	}
	if strings.Contains(f.apply.Stdout, "wrote pyproject.toml") {
		return fmt.Errorf("apply should not rewrite pyproject.toml when only user-managed dependencies changed")
	}
	return nil
}

func (f *uvFeature) moltarkApplyIsExecuted() error {
	f.apply = testutil.RunCLIInDir(f.dir, "", "apply", "-auto-approve")
	if f.apply.ExitCode != 0 {
		return fmt.Errorf("apply failed with exit code %d", f.apply.ExitCode)
	}
	return nil
}

func (f *uvFeature) theRootProjectWritesUVWorkspaceMembersRelativeToItsPath() error {
	body, err := os.ReadFile(filepath.Join(f.dir, "tally", "pyproject.toml"))
	if err != nil {
		return err
	}
	content := string(body)
	if !strings.Contains(content, `[tool.uv.workspace]`) {
		return fmt.Errorf("expected root project pyproject.toml to contain [tool.uv.workspace]")
	}
	if !strings.Contains(content, `_integrations/vscode-tally`) || !strings.Contains(content, `_integrations/intellij-tally`) {
		return fmt.Errorf("expected root project pyproject.toml to contain parent-relative workspace members")
	}
	if strings.Contains(content, `tally/_integrations/`) {
		return fmt.Errorf("workspace members must be relative to the root project path")
	}
	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	feature := &uvFeature{}
	ctx.Step(`^a Python repository bootstrapped by Moltark$`, feature.aPythonRepositoryBootstrappedByMoltark)
	ctx.Step(`^a repository declaring a nested uv workspace$`, feature.aRepositoryDeclaringANestedUVWorkspace)
	ctx.Step(`^Moltark plan is executed$`, feature.moltarkPlanIsExecuted)
	ctx.Step(`^Moltark apply is executed$`, feature.moltarkApplyIsExecuted)
	ctx.Step(`^no dependency drift is reported$`, feature.noDependencyDriftIsReported)
	ctx.Step(`^Moltark apply makes no dependency changes$`, feature.moltarkApplyMakesNoDependencyChanges)
	ctx.Step(`^the root project writes uv workspace members relative to its path$`, feature.theRootProjectWritesUVWorkspaceMembersRelativeToItsPath)
}

func TestFeatures(t *testing.T) {
	suite := godog.TestSuite{
		Name:                "moltark-features",
		ScenarioInitializer: InitializeScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{"."},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fail()
	}
}
