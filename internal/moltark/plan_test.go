package moltark

import "testing"

func TestClassifyPathUsesDesiredStateReasonForNewOwnedPathInCurrentTemplate(t *testing.T) {
	change := classifyPath(
		"pyproject.toml",
		"tool.uv.workspace.members",
		[]string{"packages/api"},
		nil,
		false,
		&ManagedFileState{Fingerprints: map[string]string{}},
		&State{TemplateVersion: TemplateVersion},
	)

	if change.Reason != ReasonDesiredState {
		t.Fatalf("expected desired-state reason, got %q", change.Reason)
	}
}

func TestClassifyPathUsesTemplateUpgradeReasonForNewOwnedPathInOlderTemplate(t *testing.T) {
	change := classifyPath(
		"pyproject.toml",
		"tool.uv.workspace.members",
		[]string{"packages/api"},
		nil,
		false,
		&ManagedFileState{Fingerprints: map[string]string{}},
		&State{TemplateVersion: "python/v1"},
	)

	if change.Reason != ReasonTemplateUpgrade {
		t.Fatalf("expected template-upgrade reason, got %q", change.Reason)
	}
}
