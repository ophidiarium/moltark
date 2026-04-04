package testutil

import (
	"errors"
	"testing"
)

func TestPrepareFixtureCleansUpTempDirWhenCopyFails(t *testing.T) {
	expectedErr := errors.New("copy failed")
	removedPath := ""

	_, err := prepareFixture(
		"missing-fixture",
		func(_, _ string) (string, error) {
			return "/tmp/moltark-fixture-test", nil
		},
		func(_, _ string) error {
			return expectedErr
		},
		func(path string) error {
			removedPath = path
			return nil
		},
	)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected copy error, got %v", err)
	}
	if removedPath != "/tmp/moltark-fixture-test" {
		t.Fatalf("expected cleanup for temp dir, got %q", removedPath)
	}
}
