package moltark

import (
	"os"
	"testing"
)

func TestBuildPipelineSeparatesCurrentAndNextState(t *testing.T) {
	root, err := prepareFixture("upgraded_template_v1_to_v2")
	if err != nil {
		t.Fatalf("prepare fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(root)
	})

	pipeline, err := NewService().BuildPipeline(root)
	if err != nil {
		t.Fatalf("build pipeline: %v", err)
	}

	if pipeline.Inspect.CurrentState == nil {
		t.Fatal("expected inspection phase to expose current state")
	}
	if pipeline.Persist.NextState == nil {
		t.Fatal("expected persist phase to expose next state")
	}
	if pipeline.stateRaw == pipeline.nextStateRaw {
		t.Fatal("expected current and next persisted state to differ for upgrade fixture")
	}
	if pipeline.Plan.Summary.Update == 0 {
		t.Fatal("expected planning phase to include updates")
	}
}

func TestBuildPipelineHandlesBootstrapWithoutCurrentState(t *testing.T) {
	root, err := prepareFixture("go_lint_vscode_component")
	if err != nil {
		t.Fatalf("prepare fixture: %v", err)
	}
	t.Cleanup(func() {
		_ = os.RemoveAll(root)
	})

	pipeline, err := NewService().BuildPipeline(root)
	if err != nil {
		t.Fatalf("build pipeline: %v", err)
	}

	if pipeline.Inspect.CurrentState != nil {
		t.Fatal("expected bootstrap fixture to have no current state")
	}
	if pipeline.Persist.NextState == nil {
		t.Fatal("expected bootstrap fixture to produce a next state")
	}
	if len(pipeline.Inspect.StructuredFiles) == 0 {
		t.Fatal("expected inspection phase to include structured files")
	}
	if pipeline.Plan.Summary.Create == 0 {
		t.Fatal("expected planning phase to include creates for bootstrap fixture")
	}
}
