package moltark

import "testing"

func TestCurrentManagedGitattributesBlockUsesEndAfterMatchedBegin(t *testing.T) {
	raw := "# END Moltark managed\n*.go text eol=lf\n# BEGIN Moltark managed\nold block\n# END Moltark managed\n"

	block, ok := currentManagedGitattributesBlock(raw)
	if !ok {
		t.Fatal("expected managed block to be found")
	}

	want := "# BEGIN Moltark managed\nold block\n# END Moltark managed"
	if block != want {
		t.Fatalf("unexpected managed block:\n%s", block)
	}
}
