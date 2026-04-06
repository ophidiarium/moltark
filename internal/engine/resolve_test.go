package engine

import (
	"strings"
	"testing"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
)

func TestResolveModelResolvesFactReferencesInStructuredFiles(t *testing.T) {
	desired := model.DesiredModel{
		Projects: []model.ProjectSpec{
			{
				ID:            "app",
				Kind:          model.ProjectKindGeneric,
				Name:          "app",
				Path:          ".",
				EffectivePath: ".",
			},
		},
		Components: []model.ComponentSpec{
			{
				ID:              "component_fact",
				Kind:            "fact",
				Module:          model.ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []model.FactProviderSpec{
					{
						Name:           model.FactLanguageGo,
						ScopeProjectID: "app",
						Values: map[string]any{
							"version": "1.24",
						},
					},
				},
			},
			{
				ID:              "component_file",
				Kind:            "toml_file",
				Module:          model.ModuleSourceCore,
				TargetProjectID: "app",
				Files: []model.StructuredFileSpec{
					{
						Path:       ".golangci.toml",
						Format:     model.FileFormatTOML,
						OwnedPaths: []string{"run.go"},
						DesiredValues: map[string]any{
							"run": map[string]any{
								"go": model.FactValueRef{
									TargetProjectID: "app",
									Name:            model.FactLanguageGo,
									Path:            "version",
								},
							},
						},
					},
				},
			},
		},
	}

	resolved, err := ResolveModel(desired)
	if err != nil {
		t.Fatalf("ResolveModel: %v", err)
	}

	if len(resolved.ManagedFiles) != 1 {
		t.Fatalf("expected 1 managed file, got %d", len(resolved.ManagedFiles))
	}

	got, ok := filefmt.LookupStructuredValue(resolved.ManagedFiles[0].DesiredValues, model.FileFormatTOML, "run.go")
	if !ok {
		t.Fatalf("expected run.go to be resolved")
	}
	if got != "1.24" {
		t.Fatalf("unexpected fact value: got %#v want %q", got, "1.24")
	}
}

func TestResolveModelRejectsAmbiguousFactsAtSameScope(t *testing.T) {
	desired := model.DesiredModel{
		Projects: []model.ProjectSpec{
			{
				ID:            "app",
				Kind:          model.ProjectKindGeneric,
				Name:          "app",
				Path:          ".",
				EffectivePath: ".",
			},
		},
		Components: []model.ComponentSpec{
			{
				ID:              "component_fact_1",
				Kind:            "fact",
				Module:          model.ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []model.FactProviderSpec{
					{
						Name:           model.FactLanguageGo,
						ScopeProjectID: "app",
						Values:         map[string]any{"version": "1.24"},
					},
				},
			},
			{
				ID:              "component_fact_2",
				Kind:            "fact",
				Module:          model.ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []model.FactProviderSpec{
					{
						Name:           model.FactLanguageGo,
						ScopeProjectID: "app",
						Values:         map[string]any{"version": "1.25"},
					},
				},
			},
			{
				ID:              "component_file",
				Kind:            "toml_file",
				Module:          model.ModuleSourceCore,
				TargetProjectID: "app",
				Files: []model.StructuredFileSpec{
					{
						Path:       ".golangci.toml",
						Format:     model.FileFormatTOML,
						OwnedPaths: []string{"run.go"},
						DesiredValues: map[string]any{
							"run": map[string]any{
								"go": model.FactValueRef{
									TargetProjectID: "app",
									Name:            model.FactLanguageGo,
									Path:            "version",
								},
							},
						},
					},
				},
			},
		},
	}

	if _, err := ResolveModel(desired); err == nil {
		t.Fatal("expected ambiguous fact resolution to fail")
	}
}

func TestProjectScopeChainReportsActualMissingAncestor(t *testing.T) {
	desired := model.DesiredModel{
		Projects: []model.ProjectSpec{
			{ID: "leaf", ParentID: "mid"},
			{ID: "mid", ParentID: "missing"},
		},
	}

	_, err := projectScopeChain(desired, "leaf")
	if err == nil {
		t.Fatal("expected missing ancestor to fail")
	}
	if !strings.Contains(err.Error(), `project parent "missing" is not declared`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProjectWithinSubtreeReportsActualMissingAncestor(t *testing.T) {
	desired := model.DesiredModel{
		Projects: []model.ProjectSpec{
			{ID: "leaf", ParentID: "mid"},
			{ID: "mid", ParentID: "missing"},
		},
	}

	_, err := projectWithinSubtree(desired, "leaf", "root")
	if err == nil {
		t.Fatal("expected missing ancestor to fail")
	}
	if !strings.Contains(err.Error(), `project parent "missing" is not declared`) {
		t.Fatalf("unexpected error: %v", err)
	}
}
