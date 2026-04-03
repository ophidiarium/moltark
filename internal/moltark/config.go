package moltark

import (
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
)

func LoadDesiredModel(root string) (DesiredModel, error) {
	path := filepath.Join(root, MoltarkfileName)
	src, err := os.ReadFile(path)
	if err != nil {
		return DesiredModel{}, fmt.Errorf("read %s: %w", MoltarkfileName, err)
	}

	builder := newDesiredModelBuilder()
	globals := starlark.StringDict{
		"use": starlark.NewBuiltin("use", builder.useModule),
	}

	thread := &starlark.Thread{Name: MoltarkfileName}
	if _, err := starlark.ExecFile(thread, MoltarkfileName, src, globals); err != nil {
		return DesiredModel{}, fmt.Errorf("evaluate %s: %w", MoltarkfileName, err)
	}

	return builder.build()
}

func InitRepository(root string) (string, error) {
	path := filepath.Join(root, MoltarkfileName)
	if _, err := os.Stat(path); err == nil {
		return "Moltarkfile already exists. No changes made.", nil
	}

	name := filepath.Base(root)
	version := DefaultProjectVersion
	requiresPython := DefaultRequiresPython

	if raw, err := os.ReadFile(filepath.Join(root, PyprojectFileName)); err == nil {
		values, err := parseTomlValues(raw)
		if err != nil {
			return "", fmt.Errorf("read existing pyproject.toml: %w", err)
		}

		if value, ok := lookupPath(values, "project.name"); ok {
			if nameValue, ok := value.(string); ok && nameValue != "" {
				name = nameValue
			}
		}
		if value, ok := lookupPath(values, "project.version"); ok {
			if versionValue, ok := value.(string); ok && versionValue != "" {
				version = versionValue
			}
		}
		if value, ok := lookupPath(values, "project.requires-python"); ok {
			if requiresValue, ok := value.(string); ok && requiresValue != "" {
				requiresPython = requiresValue
			}
		}
	}

	content := fmt.Sprintf(
		"python = use(%q)\n\nroot = python.python_project(\n    name = %q,\n    path = \".\",\n    version = %q,\n    requires_python = %q,\n)\n",
		ModuleSourcePython,
		name,
		version,
		requiresPython,
	)

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return "", fmt.Errorf("write %s: %w", MoltarkfileName, err)
	}

	return "Created Moltarkfile. Run `moltark plan` to inspect the initial reconciliation.", nil
}

type desiredModelBuilder struct {
	projects         []*ProjectSpec
	projectByID      map[string]*ProjectSpec
	modules          map[string]localModule
	moduleFactories  map[string]localModuleFactory
	moduleBuildOrder []string
	nextProjectID    int
	nextComponentID  int
}

func newDesiredModelBuilder() *desiredModelBuilder {
	builder := &desiredModelBuilder{
		projectByID: map[string]*ProjectSpec{},
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

	_ = version

	module, err := b.loadLocalModule(source)
	if err != nil {
		return nil, err
	}
	return module.Namespace(), nil
}

func (b *desiredModelBuilder) build() (DesiredModel, error) {
	if len(b.projects) == 0 {
		return DesiredModel{}, fmt.Errorf("%s did not declare python_project()", MoltarkfileName)
	}

	model := DesiredModel{
		Projects: make([]ProjectSpec, len(b.projects)),
	}
	for i, project := range b.projects {
		model.Projects[i] = *project
	}

	if err := normalizeProjects(&model); err != nil {
		return DesiredModel{}, err
	}

	components := []ComponentSpec{}
	for _, source := range b.moduleBuildOrder {
		module, ok := b.modules[source]
		if !ok {
			continue
		}
		moduleComponents, err := module.BuildComponents(model)
		if err != nil {
			return DesiredModel{}, err
		}
		components = append(components, moduleComponents...)
	}
	model.Components = components

	return model, nil
}

func (b *desiredModelBuilder) nextProjectName() string {
	b.nextProjectID++
	return fmt.Sprintf("project_%d", b.nextProjectID)
}

func (b *desiredModelBuilder) nextComponentName() string {
	b.nextComponentID++
	return fmt.Sprintf("component_%d", b.nextComponentID)
}

func normalizeProjects(model *DesiredModel) error {
	projectByID := make(map[string]*ProjectSpec, len(model.Projects))
	for i := range model.Projects {
		projectByID[model.Projects[i].ID] = &model.Projects[i]
	}

	resolving := map[string]bool{}
	resolved := map[string]bool{}
	for i := range model.Projects {
		effectivePath, err := resolveProjectPath(&model.Projects[i], projectByID, resolving, resolved)
		if err != nil {
			return err
		}
		model.Projects[i].EffectivePath = effectivePath
	}

	paths := map[string]string{}
	for _, project := range model.Projects {
		if existing, ok := paths[project.EffectivePath]; ok {
			return fmt.Errorf("project %q conflicts with project %q at path %q", project.ID, existing, project.EffectivePath)
		}
		paths[project.EffectivePath] = project.ID
	}

	return nil
}

func resolveProjectPath(project *ProjectSpec, projectByID map[string]*ProjectSpec, resolving map[string]bool, resolved map[string]bool) (string, error) {
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

	project.EffectivePath = joinProjectPath(parentPath, normalizedPath)
	resolved[project.ID] = true
	return project.EffectivePath, nil
}

func normalizeProjectPath(value string) (string, error) {
	if value == "" {
		value = "."
	}
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

func joinProjectPath(base string, child string) string {
	if base == "." {
		return child
	}
	if child == "." {
		return base
	}
	return filepath.ToSlash(filepath.Join(base, child))
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
	return fmt.Sprintf("<python_project %s>", p.id)
}

func (p *projectRef) Type() string {
	return "python_project_ref"
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
