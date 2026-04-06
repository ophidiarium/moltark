package module

import "testing"

func TestResolveModelResolvesFactReferencesInStructuredFiles(t *testing.T) {
	// This test exercises fact resolution which lives in the engine package.
	// It was originally in moltark/facts_test.go. If it tests module-layer
	// config loading, it belongs here; otherwise it may need to move to engine.
	// For now it is kept as a placeholder to document the migration.
	t.Skip("fact resolution tests migrated to engine/resolve_test.go")
}
