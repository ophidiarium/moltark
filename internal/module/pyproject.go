package module

import "github.com/ophidiarium/moltark/internal/model"

var baseOwnedPyprojectPaths = []string{
	"project.name",
	"project.version",
	"project.requires-python",
	"build-system.requires",
	"build-system.build-backend",
	"tool.moltark.schema-version",
	"tool.moltark.template-version",
}

const uvWorkspaceMembersPath = "tool.uv.workspace.members"

func pythonProjectFileValues(project model.ProjectSpec) map[string]any {
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
				"schema-version":   model.SchemaVersion,
				"template-version": project.Python.TemplateVersion,
			},
		},
	}
}
