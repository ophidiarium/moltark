package moltark

import "testing"

func TestResolveModelResolvesFactReferencesInStructuredFiles(t *testing.T) {
	model := DesiredModel{
		Projects: []ProjectSpec{
			{
				ID:            "app",
				Kind:          ProjectKindGeneric,
				Name:          "app",
				Path:          ".",
				EffectivePath: ".",
			},
		},
		Components: []ComponentSpec{
			{
				ID:              "component_fact",
				Kind:            "fact",
				Module:          ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []FactProviderSpec{
					{
						Name:           FactLanguageGo,
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
				Module:          ModuleSourceCore,
				TargetProjectID: "app",
				Files: []StructuredFileSpec{
					{
						Path:       ".golangci.toml",
						Format:     FileFormatTOML,
						OwnedPaths: []string{"run.go"},
						DesiredValues: map[string]any{
							"run": map[string]any{
								"go": FactValueRef{
									TargetProjectID: "app",
									Name:            FactLanguageGo,
									Path:            "version",
								},
							},
						},
					},
				},
			},
		},
	}

	resolved, err := ResolveModel(model)
	if err != nil {
		t.Fatalf("ResolveModel: %v", err)
	}

	if len(resolved.ManagedFiles) != 1 {
		t.Fatalf("expected 1 managed file, got %d", len(resolved.ManagedFiles))
	}

	got, ok := lookupStructuredValue(resolved.ManagedFiles[0].DesiredValues, FileFormatTOML, "run.go")
	if !ok {
		t.Fatalf("expected run.go to be resolved")
	}
	if got != "1.24" {
		t.Fatalf("unexpected fact value: got %#v want %q", got, "1.24")
	}
}

func TestResolveModelRejectsAmbiguousFactsAtSameScope(t *testing.T) {
	model := DesiredModel{
		Projects: []ProjectSpec{
			{
				ID:            "app",
				Kind:          ProjectKindGeneric,
				Name:          "app",
				Path:          ".",
				EffectivePath: ".",
			},
		},
		Components: []ComponentSpec{
			{
				ID:              "component_fact_1",
				Kind:            "fact",
				Module:          ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []FactProviderSpec{
					{
						Name:           FactLanguageGo,
						ScopeProjectID: "app",
						Values:         map[string]any{"version": "1.24"},
					},
				},
			},
			{
				ID:              "component_fact_2",
				Kind:            "fact",
				Module:          ModuleSourceCore,
				TargetProjectID: "app",
				Facts: []FactProviderSpec{
					{
						Name:           FactLanguageGo,
						ScopeProjectID: "app",
						Values:         map[string]any{"version": "1.25"},
					},
				},
			},
			{
				ID:              "component_file",
				Kind:            "toml_file",
				Module:          ModuleSourceCore,
				TargetProjectID: "app",
				Files: []StructuredFileSpec{
					{
						Path:       ".golangci.toml",
						Format:     FileFormatTOML,
						OwnedPaths: []string{"run.go"},
						DesiredValues: map[string]any{
							"run": map[string]any{
								"go": FactValueRef{
									TargetProjectID: "app",
									Name:            FactLanguageGo,
									Path:            "version",
								},
							},
						},
					},
				},
			},
		},
	}

	if _, err := ResolveModel(model); err == nil {
		t.Fatal("expected ambiguous fact resolution to fail")
	}
}
