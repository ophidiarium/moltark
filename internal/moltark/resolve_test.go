package moltark

import (
	"strings"
	"testing"
)

func TestProjectScopeChainReportsActualMissingAncestor(t *testing.T) {
	model := DesiredModel{
		Projects: []ProjectSpec{
			{ID: "leaf", ParentID: "mid"},
			{ID: "mid", ParentID: "missing"},
		},
	}

	_, err := projectScopeChain(model, "leaf")
	if err == nil {
		t.Fatal("expected missing ancestor to fail")
	}
	if !strings.Contains(err.Error(), `project parent "missing" is not declared`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectWithinSubtreeReportsActualMissingAncestor(t *testing.T) {
	model := DesiredModel{
		Projects: []ProjectSpec{
			{ID: "leaf", ParentID: "mid"},
			{ID: "mid", ParentID: "missing"},
		},
	}

	_, err := projectWithinSubtree(model, "leaf", "root")
	if err == nil {
		t.Fatal("expected missing ancestor to fail")
	}
	if !strings.Contains(err.Error(), `project parent "missing" is not declared`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
