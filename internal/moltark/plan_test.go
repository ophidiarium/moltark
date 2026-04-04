package moltark

import "testing"

func TestClassifyPathUsesDesiredStateReasonForNewOwnedPathInCurrentTemplate(t *testing.T) {
	change, err := classifyPath(
		FileFormatTOML,
		"pyproject.toml",
		"tool.uv.workspace.members",
		"component_1",
		UVModuleVersion,
		[]string{"packages/api"},
		nil,
		false,
		&ManagedFileState{Fingerprints: map[string]string{}},
		&State{
			LastAppliedModel: ModelSummary{
				Components: []ComponentSummary{
					{ID: "component_1", Version: UVModuleVersion},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("classifyPath: %v", err)
	}

	if change.Reason != ReasonDesiredState {
		t.Fatalf("expected desired-state reason, got %q", change.Reason)
	}
}

func TestClassifyPathUsesTemplateUpgradeReasonForNewOwnedPathInOlderTemplate(t *testing.T) {
	change, err := classifyPath(
		FileFormatTOML,
		"pyproject.toml",
		"tool.uv.workspace.members",
		"component_1",
		UVModuleVersion,
		[]string{"packages/api"},
		nil,
		false,
		&ManagedFileState{Fingerprints: map[string]string{}},
		&State{
			LastAppliedModel: ModelSummary{
				Components: []ComponentSummary{
					{ID: "component_1", Version: "astral/uv/v0"},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("classifyPath: %v", err)
	}

	if change.Reason != ReasonTemplateUpgrade {
		t.Fatalf("expected template-upgrade reason, got %q", change.Reason)
	}
}
