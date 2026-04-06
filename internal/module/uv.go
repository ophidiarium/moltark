package module

import (
	"fmt"

	"github.com/ophidiarium/moltark/internal/model"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type uvWorkspaceDecl struct {
	rootProjectID string
	memberIDs     []string
}

type uvModuleRuntime struct {
	builder *desiredModelBuilder
	decls   []uvWorkspaceDecl
}

func newUVModuleRuntime(builder *desiredModelBuilder) localModule {
	return &uvModuleRuntime{builder: builder}
}

func (m *uvModuleRuntime) Namespace() starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String(model.ModuleSourceUV), starlark.StringDict{
		"uv_workspace": starlark.NewBuiltin("uv_workspace", m.uvWorkspace),
	})
}

func (m *uvModuleRuntime) BuildComponents(desired model.DesiredModel) ([]model.ComponentSpec, error) {
	components := make([]model.ComponentSpec, 0, len(m.decls))
	for _, decl := range m.decls {
		component, err := m.buildUVWorkspaceComponent(desired, decl)
		if err != nil {
			return nil, err
		}
		components = append(components, component)
	}
	return components, nil
}

func (m *uvModuleRuntime) uvWorkspace(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if len(m.decls) > 0 {
		return nil, fmt.Errorf("only one uv_workspace may be declared")
	}

	var rootValue starlark.Value
	var membersValue starlark.Value
	if err := starlark.UnpackArgs("uv_workspace", args, kwargs,
		"root", &rootValue,
		"members", &membersValue,
	); err != nil {
		return nil, err
	}

	rootRef, ok := rootValue.(*projectRef)
	if !ok {
		return nil, fmt.Errorf("uv_workspace root must be a project reference")
	}
	if _, ok := m.builder.projectByID[rootRef.id]; !ok {
		return nil, fmt.Errorf("uv_workspace root %q is not declared", rootRef.id)
	}

	iterable, ok := membersValue.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("uv_workspace members must be a list of project references")
	}

	memberIDs := []string{}
	seen := map[string]struct{}{}
	iter := iterable.Iterate()
	defer iter.Done()
	var value starlark.Value
	for iter.Next(&value) {
		ref, ok := value.(*projectRef)
		if !ok {
			return nil, fmt.Errorf("uv_workspace members must contain only project references")
		}
		if ref.id == rootRef.id {
			return nil, fmt.Errorf("uv_workspace members must not include the root project")
		}
		if _, ok := m.builder.projectByID[ref.id]; !ok {
			return nil, fmt.Errorf("uv_workspace member %q is not declared", ref.id)
		}
		if _, ok := seen[ref.id]; ok {
			return nil, fmt.Errorf("uv_workspace member %q is declared more than once", ref.id)
		}
		seen[ref.id] = struct{}{}
		memberIDs = append(memberIDs, ref.id)
	}

	m.decls = append(m.decls, uvWorkspaceDecl{
		rootProjectID: rootRef.id,
		memberIDs:     memberIDs,
	})
	return starlark.None, nil
}

func (m *uvModuleRuntime) buildUVWorkspaceComponent(desired model.DesiredModel, decl uvWorkspaceDecl) (model.ComponentSpec, error) {
	root := desired.ProjectByID(decl.rootProjectID)
	if root == nil {
		return model.ComponentSpec{}, fmt.Errorf("uv_workspace root %q is not declared", decl.rootProjectID)
	}

	memberPaths := make([]string, 0, len(decl.memberIDs))
	for _, memberID := range decl.memberIDs {
		member := desired.ProjectByID(memberID)
		if member == nil {
			return model.ComponentSpec{}, fmt.Errorf("uv_workspace member %q is not declared", memberID)
		}
		relativePath, err := relativeWorkspaceMemberPath(root.EffectivePath, member.EffectivePath)
		if err != nil {
			return model.ComponentSpec{}, fmt.Errorf("uv_workspace member %q: %w", memberID, err)
		}
		memberPaths = append(memberPaths, relativePath)
	}

	providers := []model.CapabilityProvider{
		{
			Capability:     model.CapabilityPythonWorkspaceManager,
			ScopeProjectID: root.ID,
			Attributes: map[string]string{
				model.ProviderAttrManager:   "uv",
				model.ProviderAttrFilePath:  desired.ProjectPyprojectPath(*root),
				model.ProviderAttrOwnedPath: uvWorkspaceMembersPath,
			},
		},
	}

	for _, projectID := range append([]string{root.ID}, decl.memberIDs...) {
		project := desired.ProjectByID(projectID)
		if project == nil {
			return model.ComponentSpec{}, fmt.Errorf("uv_workspace project %q is not declared", projectID)
		}
		providers = append(providers, model.CapabilityProvider{
			Capability:     model.CapabilityPythonPackageManager,
			ScopeProjectID: project.ID,
			Attributes: map[string]string{
				model.ProviderAttrManager:        "uv",
				model.ProviderAttrEcosystem:      "python",
				model.ProviderAttrFilePath:       desired.ProjectPyprojectPath(*project),
				model.ProviderAttrDependencyPath: "project.dependencies",
			},
			Lists: map[string][]string{
				model.ProviderListArtifactKinds: {model.ArtifactKindPyPI},
			},
		})
	}

	return model.ComponentSpec{
		ID:              m.builder.nextComponentName(),
		Kind:            "uv_workspace",
		Module:          model.ModuleSourceUV,
		Version:         model.UVModuleVersion,
		TargetProjectID: root.ID,
		Providers:       providers,
		RoutedIntents: []model.RoutedIntentSpec{
			{
				Kind:            model.IntentWorkspaceMembersRequest,
				Capability:      model.CapabilityPythonWorkspaceManager,
				TargetProjectID: root.ID,
				Lists: map[string][]string{
					model.IntentListMemberPaths: memberPaths,
				},
			},
		},
	}, nil
}
