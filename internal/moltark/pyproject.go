package moltark

func pythonProjectFileValues(project ProjectSpec) map[string]any {
	if project.Python == nil {
		return map[string]any{}
	}

	return map[string]any{
		"build-system": map[string]any{
			"requires":      append([]string(nil), project.Python.BuildSystem.Requires...),
			"build-backend": project.Python.BuildSystem.Backend,
		},
		"project": map[string]any{
			"name":            project.Name,
			"version":         project.Python.Version,
			"requires-python": project.Python.RequiresPython,
		},
		"tool": map[string]any{
			"moltark": map[string]any{
				"schema-version":   SchemaVersion,
				"template-version": project.Python.TemplateVersion,
			},
		},
	}
}
