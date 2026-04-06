package module

import (
	"fmt"

	"github.com/ophidiarium/moltark/internal/model"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type pythonModuleRuntime struct {
	builder *desiredModelBuilder
}

func newPythonModuleRuntime(builder *desiredModelBuilder) localModule {
	return &pythonModuleRuntime{builder: builder}
}

func (m *pythonModuleRuntime) Namespace() starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String(model.ModuleSourcePython), starlark.StringDict{
		"python_project": starlark.NewBuiltin("python_project", m.pythonProject),
	})
}

func (m *pythonModuleRuntime) BuildComponents(desired model.DesiredModel) ([]model.ComponentSpec, error) {
	components := make([]model.ComponentSpec, 0, len(desired.Projects))
	for _, project := range desired.Projects {
		if project.Kind != model.ProjectKindPython {
			continue
		}
		components = append(components, model.ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            model.ProjectKindPython,
			Module:          model.ModuleSourcePython,
			Version:         project.Python.TemplateVersion,
			TargetProjectID: project.ID,
			Facts: []model.FactProviderSpec{
				{
					Name:           model.FactLanguagePython,
					ScopeProjectID: project.ID,
					Values: map[string]any{
						"requires_python": project.Python.RequiresPython,
						"package_version": project.Python.Version,
					},
				},
			},
			Files: []model.StructuredFileSpec{
				{
					Path:             desired.ProjectPyprojectPath(project),
					Format:           model.FileFormatTOML,
					OwnedPaths:       append([]string(nil), baseOwnedPyprojectPaths...),
					UserManagedPaths: []string{"project.dependencies"},
					DesiredValues:    pythonProjectFileValues(project),
				},
			},
		})
	}
	return components, nil
}

func (m *pythonModuleRuntime) pythonProject(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var path string = "."
	var version string
	var requiresPython string
	var id string
	var parentValue starlark.Value = starlark.None

	if err := starlark.UnpackArgs("python_project", args, kwargs,
		"name", &name,
		"path?", &path,
		"version", &version,
		"requires_python", &requiresPython,
		"id?", &id,
		"parent?", &parentValue,
	); err != nil {
		return nil, err
	}

	parentID := ""
	if parentValue != starlark.None {
		ref, ok := parentValue.(*projectRef)
		if !ok {
			return nil, fmt.Errorf("python_project parent must be a project reference")
		}
		parentID = ref.id
		if _, ok := m.builder.projectByID[parentID]; !ok {
			return nil, fmt.Errorf("python_project parent %q is not declared", parentID)
		}
	}

	if id == "" {
		id = m.builder.nextProjectName()
	}
	if _, exists := m.builder.projectByID[id]; exists {
		return nil, fmt.Errorf("python_project id %q is already declared", id)
	}

	project := &model.ProjectSpec{
		ID:       id,
		Kind:     model.ProjectKindPython,
		Name:     name,
		Path:     path,
		ParentID: parentID,
		Python: &model.PythonProjectSpec{
			Version:         version,
			RequiresPython:  requiresPython,
			TemplateVersion: model.PythonTemplateVersion,
			BuildSystem: model.BuildSystem{
				Requires: []string{model.DefaultBuildRequirement},
				Backend:  model.DefaultBuildBackend,
			},
		},
	}

	m.builder.projects = append(m.builder.projects, project)
	m.builder.projectByID[id] = project
	return &projectRef{id: id}, nil
}
