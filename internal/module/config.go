package module

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
	"go.starlark.net/starlark"
)

func LoadDesiredModel(root string) (model.DesiredModel, error) {
	path := filepath.Join(root, model.ProjectSpecFileName)
	src, err := os.ReadFile(path)
	if err != nil {
		return model.DesiredModel{}, fmt.Errorf("read %s: %w", model.ProjectSpecFileName, err)
	}

	builder := newDesiredModelBuilder()
	globals := starlark.StringDict{
		"use": starlark.NewBuiltin("use", builder.useModule),
	}

	thread := &starlark.Thread{Name: model.ProjectSpecFileName}
	if _, err := starlark.ExecFile(thread, model.ProjectSpecFileName, src, globals); err != nil {
		return model.DesiredModel{}, fmt.Errorf("evaluate %s: %w", model.ProjectSpecFileName, err)
	}

	return builder.build()
}

func InitRepository(root string) (string, error) {
	path := filepath.Join(root, model.ProjectSpecFileName)
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("%s already exists. No changes made.", model.ProjectSpecFileName), nil
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat %s: %w", model.ProjectSpecFileName, err)
	}

	name := filepath.Base(root)
	version := model.DefaultProjectVersion
	requiresPython := model.DefaultRequiresPython

	if raw, err := os.ReadFile(filepath.Join(root, model.PyprojectFileName)); err == nil {
		var info struct {
			Project struct {
				Name           string `toml:"name"`
				Version        string `toml:"version"`
				RequiresPython string `toml:"requires-python"`
			} `toml:"project"`
		}
		if err := filefmt.DecodeToml(raw, &info); err != nil {
			return "", fmt.Errorf("read existing pyproject.toml: %w", err)
		}
		if info.Project.Name != "" {
			name = info.Project.Name
		}
		if info.Project.Version != "" {
			version = info.Project.Version
		}
		if info.Project.RequiresPython != "" {
			requiresPython = info.Project.RequiresPython
		}
	}

	content := fmt.Sprintf(
		"python = use(%q)\n\nroot = python.python_project(\n    name = %q,\n    path = \".\",\n    version = %q,\n    requires_python = %q,\n)\n",
		model.ModuleSourcePython,
		name,
		version,
		requiresPython,
	)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", model.ProjectSpecFileName, err)
	}

	return fmt.Sprintf("Created %s. Run `moltark plan` to inspect the initial reconciliation.", model.ProjectSpecFileName), nil
}

type desiredModelBuilder struct {
	projects         []*model.ProjectSpec
	projectByID      map[string]*model.ProjectSpec
	modules          map[string]localModule
	moduleFactories  map[string]localModuleFactory
	moduleBuildOrder []string
	nextProjectID    int
	nextComponentID  int
}

func newDesiredModelBuilder() *desiredModelBuilder {
	builder := &desiredModelBuilder{
		projectByID: map[string]*model.ProjectSpec{},
		modules:     map[string]localModule{},
	}
	builder.registerLocalModules()
	return builder
}

func (b *desiredModelBuilder) useModule(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var source string
	var version string
	if err := starlark.UnpackArgs("use", args, kwargs,
		"source", &source,
		"version?", &version,
	); err != nil {
		return nil, err
	}

	if version != "" {
		return nil, fmt.Errorf("module version selection is not implemented yet for %q", source)
	}

	module, err := b.loadLocalModule(source)
	if err != nil {
		return nil, err
	}
	return module.Namespace(), nil
}

func (b *desiredModelBuilder) build() (model.DesiredModel, error) {
	if len(b.projects) == 0 {
		return model.DesiredModel{}, fmt.Errorf("%s did not declare any projects", model.ProjectSpecFileName)
	}

	desired := model.DesiredModel{
		Projects: make([]model.ProjectSpec, len(b.projects)),
	}
	for i, project := range b.projects {
		desired.Projects[i] = *project
	}

	if err := normalizeProjects(&desired); err != nil {
		return model.DesiredModel{}, err
	}

	components := []model.ComponentSpec{}
	for _, source := range b.moduleBuildOrder {
		module, ok := b.modules[source]
		if !ok {
			continue
		}
		moduleComponents, err := module.BuildComponents(desired)
		if err != nil {
			return model.DesiredModel{}, err
		}
		components = append(components, moduleComponents...)
	}
	desired.Components = components

	return desired, nil
}

func (b *desiredModelBuilder) nextProjectName() string {
	b.nextProjectID++
	return fmt.Sprintf("project_%d", b.nextProjectID)
}

func (b *desiredModelBuilder) nextComponentName() string {
	b.nextComponentID++
	return fmt.Sprintf("component_%d", b.nextComponentID)
}

func normalizeProjects(desired *model.DesiredModel) error {
	projectByID := make(map[string]*model.ProjectSpec, len(desired.Projects))
	for i := range desired.Projects {
		projectByID[desired.Projects[i].ID] = &desired.Projects[i]
	}

	resolving := map[string]bool{}
	resolved := map[string]bool{}
	for i := range desired.Projects {
		effectivePath, err := resolveProjectPath(&desired.Projects[i], projectByID, resolving, resolved)
		if err != nil {
			return err
		}
		desired.Projects[i].EffectivePath = effectivePath
	}

	paths := map[string]string{}
	for _, project := range desired.Projects {
		if existing, ok := paths[project.EffectivePath]; ok {
			return fmt.Errorf("project %q conflicts with project %q at path %q", project.ID, existing, project.EffectivePath)
		}
		paths[project.EffectivePath] = project.ID
	}

	return nil
}

func resolveProjectPath(project *model.ProjectSpec, projectByID map[string]*model.ProjectSpec, resolving map[string]bool, resolved map[string]bool) (string, error) {
	if resolved[project.ID] {
		return project.EffectivePath, nil
	}
	if resolving[project.ID] {
		return "", fmt.Errorf("project parent cycle detected at %q", project.ID)
	}

	normalizedPath, err := normalizeProjectPath(project.Path)
	if err != nil {
		return "", fmt.Errorf("project %q path %q: %w", project.ID, project.Path, err)
	}

	resolving[project.ID] = true
	defer delete(resolving, project.ID)

	if project.ParentID == "" {
		project.EffectivePath = normalizedPath
		resolved[project.ID] = true
		return project.EffectivePath, nil
	}

	parent := projectByID[project.ParentID]
	if parent == nil {
		return "", fmt.Errorf("project %q parent %q is not declared", project.ID, project.ParentID)
	}

	parentPath, err := resolveProjectPath(parent, projectByID, resolving, resolved)
	if err != nil {
		return "", err
	}

	project.EffectivePath = model.JoinProjectPath(parentPath, normalizedPath)
	resolved[project.ID] = true
	return project.EffectivePath, nil
}

func normalizeProjectPath(value string) (string, error) {
	if value == "" {
		value = "."
	}
	// Canonicalize to forward slashes before filepath.Clean so that ".."
	// segments separated by backslashes are resolved on every platform.
	// On Unix filepath.Clean treats '\' as a literal filename character,
	// so "pkg\..\..\outside" would pass the escape check untouched.
	value = strings.ReplaceAll(value, "\\", "/")
	if filepath.IsAbs(value) {
		return "", fmt.Errorf("must be relative")
	}

	cleaned := filepath.ToSlash(filepath.Clean(value))
	if cleaned == "." {
		return ".", nil
	}
	if cleaned == ".." || strings.HasPrefix(cleaned, "../") {
		return "", fmt.Errorf("must not escape the repository")
	}
	return cleaned, nil
}

func (b *desiredModelBuilder) requireProjectRef(value starlark.Value, label string) (string, error) {
	ref, ok := value.(*projectRef)
	if !ok {
		return "", fmt.Errorf("%s must be a project reference", label)
	}
	if _, ok := b.projectByID[ref.id]; !ok {
		return "", fmt.Errorf("%s %q is not declared", label, ref.id)
	}
	return ref.id, nil
}

func relativeWorkspaceMemberPath(rootPath string, memberPath string) (string, error) {
	relativePath, err := filepath.Rel(rootPath, memberPath)
	if err != nil {
		return "", err
	}
	relativePath = filepath.ToSlash(relativePath)
	if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, "../") {
		return "", fmt.Errorf("path %q is not contained inside root %q", memberPath, rootPath)
	}
	return relativePath, nil
}

type projectRef struct {
	id string
}

func (p *projectRef) String() string {
	return fmt.Sprintf("<project %s>", p.id)
}

func (p *projectRef) Type() string {
	return "project_ref"
}

func (p *projectRef) Freeze() {}

func (p *projectRef) Truth() starlark.Bool {
	return starlark.True
}

func (p *projectRef) Hash() (uint32, error) {
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(p.id))
	return hash.Sum32(), nil
}
