package engine

import (
	"testing"

	"github.com/ophidiarium/moltark/internal/model"
)

func TestClassifyPathUsesDesiredStateReasonForNewOwnedPathInCurrentTemplate(t *testing.T) {
	change, err := classifyPath(
		model.FileFormatTOML,
		"pyproject.toml",
		"tool.uv.workspace.members",
		"component_1",
		model.UVModuleVersion,
		[]string{"packages/api"},
		nil,
		false,
		&model.ManagedFileState{Fingerprints: map[string]string{}},
		&model.State{
			LastAppliedModel: model.ModelSummary{
				Components: []model.ComponentSummary{
					{ID: "component_1", Version: model.UVModuleVersion},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("classifyPath: %v", err)
	}

	if change.Reason != model.ReasonDesiredState {
		t.Fatalf("expected desired-state reason, got %q", change.Reason)
	}
}

func TestClassifyPathUsesTemplateUpgradeReasonForNewOwnedPathInOlderTemplate(t *testing.T) {
	change, err := classifyPath(
		model.FileFormatTOML,
		"pyproject.toml",
		"tool.uv.workspace.members",
		"component_1",
		model.UVModuleVersion,
		[]string{"packages/api"},
		nil,
		false,
		&model.ManagedFileState{Fingerprints: map[string]string{}},
		&model.State{
			LastAppliedModel: model.ModelSummary{
				Components: []model.ComponentSummary{
					{ID: "component_1", Version: "astral/uv/v0"},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("classifyPath: %v", err)
	}

	if change.Reason != model.ReasonTemplateUpgrade {
		t.Fatalf("expected template-upgrade reason, got %q", change.Reason)
	}
}
