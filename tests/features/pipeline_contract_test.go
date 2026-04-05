package features_test

import (
	"fmt"
	"strings"

	"github.com/ophidiarium/moltark/internal/testutil"
)

type pipelineFeature struct {
	dir    string
	plan   testutil.CommandResult
	apply  testutil.CommandResult
	doctor testutil.CommandResult
}

func (f *pipelineFeature) aFreshRepositoryInitializedByMoltark() error {
	dir, err := testutil.PrepareFixture("bare_python_repo_with_uv_deps")
	if err != nil {
		return err
	}
	f.dir = dir
	return nil
}

func (f *pipelineFeature) aRepositoryWithDriftedOwnedFields() error {
	dir, err := testutil.PrepareFixture("drifted_owned_fields")
	if err != nil {
		return err
	}
	f.dir = dir
	return nil
}

func (f *pipelineFeature) moltarkApplyIsExecuted() error {
	f.apply = testutil.RunCLIInDir(f.dir, "", "apply", "-auto-approve")
	if f.apply.ExitCode != 0 {
		return fmt.Errorf("apply failed with exit code %d: %s", f.apply.ExitCode, f.apply.Stderr)
	}
	return nil
}

func (f *pipelineFeature) moltarkPlanIsExecutedAgain() error {
	f.plan = testutil.RunCLIInDir(f.dir, "", "plan")
	if f.plan.ExitCode != 0 {
		return fmt.Errorf("plan failed with exit code %d: %s", f.plan.ExitCode, f.plan.Stderr)
	}
	return nil
}

func (f *pipelineFeature) thePlanReportsNoPendingChanges() error {
	if !strings.Contains(f.plan.Stdout, "0 to create, 0 to update") {
		return fmt.Errorf("expected no pending changes after re-apply, got: %s", f.plan.Stdout)
	}
	return nil
}

func (f *pipelineFeature) moltarkDoctorIsExecuted() error {
	f.doctor = testutil.RunCLIInDir(f.dir, "", "doctor")
	return nil
}

func (f *pipelineFeature) doctorReportsDriftDetected() error {
	if f.doctor.ExitCode != 1 {
		return fmt.Errorf("expected doctor to exit 1 for drift, got %d", f.doctor.ExitCode)
	}
	if !strings.Contains(f.doctor.Stdout, "drift") {
		return fmt.Errorf("expected drift message in doctor output")
	}
	return nil
}
