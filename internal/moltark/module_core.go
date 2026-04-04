package moltark

import (
	"fmt"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type synthesisHookDecl struct {
	phase           string
	targetProjectID string
	description     string
}

type bootstrapRequirementDecl struct {
	tool            string
	targetProjectID string
	purpose         string
	strategies      []string
}

type factDecl struct {
	name            string
	targetProjectID string
	values          map[string]any
}

type structuredFileDecl struct {
	format          string
	targetProjectID string
	path            string
	values          map[string]any
}

type taskDecl struct {
	name            string
	targetProjectID string
	command         []string
	runtime         string
	tags            []string
}

type taskSurfaceDecl struct {
	name            string
	kind            string
	targetProjectID string
}

type triggerBindingDecl struct {
	trigger         string
	targetProjectID string
	matchNames      []string
	matchTags       []string
}

type pythonDependencyDecl struct {
	targetProjectID string
	requirement     string
}

type coreModuleRuntime struct {
	builder                   *desiredModelBuilder
	factDecls                 []factDecl
	structuredFileDecls       []structuredFileDecl
	synthesisHookDecls        []synthesisHookDecl
	bootstrapRequirementDecls []bootstrapRequirementDecl
	taskDecls                 []taskDecl
	taskSurfaceDecls          []taskSurfaceDecl
	triggerBindingDecls       []triggerBindingDecl
	pythonDependencyDecls     []pythonDependencyDecl
}

func newCoreModuleRuntime(builder *desiredModelBuilder) localModule {
	return &coreModuleRuntime{builder: builder}
}

func (m *coreModuleRuntime) Namespace() starlark.Value {
	return starlarkstruct.FromStringDict(starlark.String(ModuleSourceCore), starlark.StringDict{
		"project":               starlark.NewBuiltin("project", m.project),
		"fact":                  starlark.NewBuiltin("fact", m.fact),
		"fact_value":            starlark.NewBuiltin("fact_value", m.factValue),
		"json_file":             starlark.NewBuiltin("json_file", m.jsonFile),
		"toml_file":             starlark.NewBuiltin("toml_file", m.tomlFile),
		"yaml_file":             starlark.NewBuiltin("yaml_file", m.yamlFile),
		"synthesis_hook":        starlark.NewBuiltin("synthesis_hook", m.synthesisHook),
		"bootstrap_requirement": starlark.NewBuiltin("bootstrap_requirement", m.bootstrapRequirement),
		"python_dependency":     starlark.NewBuiltin("python_dependency", m.pythonDependency),
		"task":                  starlark.NewBuiltin("task", m.task),
		"task_surface":          starlark.NewBuiltin("task_surface", m.taskSurface),
		"trigger_binding":       starlark.NewBuiltin("trigger_binding", m.triggerBinding),
	})
}

func (m *coreModuleRuntime) BuildComponents(_ DesiredModel) ([]ComponentSpec, error) {
	components := make([]ComponentSpec, 0, len(m.factDecls)+len(m.structuredFileDecls)+len(m.synthesisHookDecls)+len(m.bootstrapRequirementDecls)+len(m.pythonDependencyDecls)+len(m.taskDecls)+len(m.taskSurfaceDecls)+len(m.triggerBindingDecls))

	for _, decl := range m.factDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "fact",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			Facts: []FactProviderSpec{
				{
					Name:           decl.name,
					ScopeProjectID: decl.targetProjectID,
					Values:         cloneNestedMap(decl.values),
				},
			},
		})
	}

	for _, decl := range m.structuredFileDecls {
		project := m.builder.projectByID[decl.targetProjectID]
		if project == nil {
			return nil, fmt.Errorf("%s_file target %q is not declared", decl.format, decl.targetProjectID)
		}

		ownedPaths, err := inferOwnedPaths(decl.format, decl.values)
		if err != nil {
			return nil, fmt.Errorf("%s_file %q: %w", decl.format, decl.path, err)
		}

		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            decl.format + "_file",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			Files: []StructuredFileSpec{
				{
					Path:          joinProjectPath(project.EffectivePath, decl.path),
					Format:        decl.format,
					OwnedPaths:    ownedPaths,
					DesiredValues: cloneNestedMap(decl.values),
				},
			},
		})
	}

	for _, decl := range m.synthesisHookDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "synthesis_hook",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			SynthesisHooks: []SynthesisHookSpec{
				{
					Phase:           decl.phase,
					TargetProjectID: decl.targetProjectID,
					Description:     decl.description,
				},
			},
		})
	}

	for _, decl := range m.bootstrapRequirementDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "bootstrap_requirement",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			BootstrapRequirements: []BootstrapRequirementSpec{
				{
					Tool:            decl.tool,
					TargetProjectID: decl.targetProjectID,
					Purpose:         decl.purpose,
					Strategies:      append([]string(nil), decl.strategies...),
				},
			},
		})
	}

	for _, decl := range m.pythonDependencyDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "python_dependency",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			RoutedIntents: []RoutedIntentSpec{
				{
					Kind:            IntentPythonDependencyRequest,
					Capability:      CapabilityPythonPackageManager,
					TargetProjectID: decl.targetProjectID,
					Attributes: map[string]string{
						IntentAttrRequirement: decl.requirement,
					},
				},
			},
		})
	}

	for _, decl := range m.taskDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "task",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			Tasks: []TaskSpec{
				{
					Name:            decl.name,
					TargetProjectID: decl.targetProjectID,
					Command:         append([]string(nil), decl.command...),
					Runtime:         decl.runtime,
					Tags:            append([]string(nil), decl.tags...),
				},
			},
		})
	}

	for _, decl := range m.taskSurfaceDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "task_surface",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			TaskSurfaces: []TaskSurfaceSpec{
				{
					Name:            decl.name,
					Kind:            decl.kind,
					TargetProjectID: decl.targetProjectID,
				},
			},
		})
	}

	for _, decl := range m.triggerBindingDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "trigger_binding",
			Module:          ModuleSourceCore,
			Version:         CoreModuleVersion,
			TargetProjectID: decl.targetProjectID,
			TriggerBindings: []TriggerBindingSpec{
				{
					Trigger:         decl.trigger,
					TargetProjectID: decl.targetProjectID,
					MatchNames:      append([]string(nil), decl.matchNames...),
					MatchTags:       append([]string(nil), decl.matchTags...),
				},
			},
		})
	}

	return components, nil
}

func (m *coreModuleRuntime) project(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var kind string = ProjectKindGeneric
	var name string
	var path string = "."
	var id string
	var parentValue starlark.Value = starlark.None
	var attributesValue starlark.Value = starlark.None

	if err := starlark.UnpackArgs("project", args, kwargs,
		"name?", &name,
		"path?", &path,
		"kind?", &kind,
		"id?", &id,
		"parent?", &parentValue,
		"attributes?", &attributesValue,
	); err != nil {
		return nil, err
	}

	parentID := ""
	if parentValue != starlark.None {
		ref, ok := parentValue.(*projectRef)
		if !ok {
			return nil, fmt.Errorf("project parent must be a project reference")
		}
		parentID = ref.id
		if _, ok := m.builder.projectByID[parentID]; !ok {
			return nil, fmt.Errorf("project parent %q is not declared", parentID)
		}
	}

	if id == "" {
		id = m.builder.nextProjectName()
	}
	if _, exists := m.builder.projectByID[id]; exists {
		return nil, fmt.Errorf("project id %q is already declared", id)
	}

	attributes, err := starlarkStringMap(attributesValue, "project attributes")
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = defaultProjectName(path, id)
	}

	project := &ProjectSpec{
		ID:         id,
		Kind:       kind,
		Name:       name,
		Path:       path,
		Attributes: attributes,
		ParentID:   parentID,
	}

	m.builder.projects = append(m.builder.projects, project)
	m.builder.projectByID[id] = project
	return &projectRef{id: id}, nil
}

func (m *coreModuleRuntime) fact(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var targetValue starlark.Value
	var valuesValue starlark.Value
	if err := starlark.UnpackArgs("fact", args, kwargs,
		"name", &name,
		"target", &targetValue,
		"values", &valuesValue,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "fact target")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("fact name must not be empty")
	}

	converted, err := starlarkValueToGo(valuesValue)
	if err != nil {
		return nil, fmt.Errorf("fact values: %w", err)
	}
	values, ok := converted.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fact values must be an object")
	}
	if err := ensureNoFactRefs(values); err != nil {
		return nil, fmt.Errorf("fact values: %w", err)
	}

	m.factDecls = append(m.factDecls, factDecl{
		name:            name,
		targetProjectID: targetProjectID,
		values:          values,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) factValue(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var path string
	var targetValue starlark.Value
	if err := starlark.UnpackArgs("fact_value", args, kwargs,
		"name", &name,
		"target", &targetValue,
		"path", &path,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "fact_value target")
	if err != nil {
		return nil, err
	}
	if name == "" {
		return nil, fmt.Errorf("fact_value name must not be empty")
	}
	if path == "" {
		return nil, fmt.Errorf("fact_value path must not be empty")
	}

	return &factValueRefValue{
		targetProjectID: targetProjectID,
		name:            name,
		path:            path,
	}, nil
}

func (m *coreModuleRuntime) jsonFile(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return m.declareStructuredFile("json_file", FileFormatJSON, args, kwargs)
}

func (m *coreModuleRuntime) tomlFile(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return m.declareStructuredFile("toml_file", FileFormatTOML, args, kwargs)
}

func (m *coreModuleRuntime) yamlFile(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return m.declareStructuredFile("yaml_file", FileFormatYAML, args, kwargs)
}

func (m *coreModuleRuntime) declareStructuredFile(name string, format string, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var targetValue starlark.Value
	var path string
	var valuesValue starlark.Value
	if err := starlark.UnpackArgs(name, args, kwargs,
		"target", &targetValue,
		"path", &path,
		"values", &valuesValue,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, name+" target")
	if err != nil {
		return nil, err
	}

	normalizedPath, err := normalizeProjectPath(path)
	if err != nil {
		return nil, fmt.Errorf("%s path %q: %w", name, path, err)
	}

	converted, err := starlarkValueToGo(valuesValue)
	if err != nil {
		return nil, fmt.Errorf("%s values: %w", name, err)
	}
	values, ok := converted.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s values must be an object", name)
	}

	m.structuredFileDecls = append(m.structuredFileDecls, structuredFileDecl{
		format:          format,
		targetProjectID: targetProjectID,
		path:            normalizedPath,
		values:          values,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) synthesisHook(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var phase string
	var description string
	var targetValue starlark.Value
	if err := starlark.UnpackArgs("synthesis_hook", args, kwargs,
		"phase", &phase,
		"target", &targetValue,
		"description?", &description,
	); err != nil {
		return nil, err
	}

	if !isKnownSynthesisPhase(phase) {
		return nil, fmt.Errorf("unsupported synthesis phase %q", phase)
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "synthesis_hook target")
	if err != nil {
		return nil, err
	}

	m.synthesisHookDecls = append(m.synthesisHookDecls, synthesisHookDecl{
		phase:           phase,
		targetProjectID: targetProjectID,
		description:     description,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) bootstrapRequirement(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var tool string
	var purpose string
	var targetValue starlark.Value
	var strategiesValue starlark.Value = starlark.NewList(nil)
	if err := starlark.UnpackArgs("bootstrap_requirement", args, kwargs,
		"tool", &tool,
		"target", &targetValue,
		"purpose?", &purpose,
		"strategies?", &strategiesValue,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "bootstrap_requirement target")
	if err != nil {
		return nil, err
	}
	strategies, err := starlarkStringList(strategiesValue, "bootstrap_requirement strategies")
	if err != nil {
		return nil, err
	}

	m.bootstrapRequirementDecls = append(m.bootstrapRequirementDecls, bootstrapRequirementDecl{
		tool:            tool,
		targetProjectID: targetProjectID,
		purpose:         purpose,
		strategies:      strategies,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) pythonDependency(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var targetValue starlark.Value
	var requirement string
	if err := starlark.UnpackArgs("python_dependency", args, kwargs,
		"target", &targetValue,
		"requirement", &requirement,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "python_dependency target")
	if err != nil {
		return nil, err
	}
	if requirement == "" {
		return nil, fmt.Errorf("python_dependency requirement must not be empty")
	}

	m.pythonDependencyDecls = append(m.pythonDependencyDecls, pythonDependencyDecl{
		targetProjectID: targetProjectID,
		requirement:     requirement,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) task(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var name string
	var runtime string
	var targetValue starlark.Value
	var commandValue starlark.Value
	var tagsValue starlark.Value = starlark.NewList(nil)
	if err := starlark.UnpackArgs("task", args, kwargs,
		"name", &name,
		"target", &targetValue,
		"command", &commandValue,
		"runtime?", &runtime,
		"tags?", &tagsValue,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "task target")
	if err != nil {
		return nil, err
	}
	command, err := starlarkStringList(commandValue, "task command")
	if err != nil {
		return nil, err
	}
	if len(command) == 0 {
		return nil, fmt.Errorf("task command must not be empty")
	}
	tags, err := starlarkStringList(tagsValue, "task tags")
	if err != nil {
		return nil, err
	}

	m.taskDecls = append(m.taskDecls, taskDecl{
		name:            name,
		targetProjectID: targetProjectID,
		command:         command,
		runtime:         runtime,
		tags:            tags,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) taskSurface(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var kind string
	var name string
	var targetValue starlark.Value
	if err := starlark.UnpackArgs("task_surface", args, kwargs,
		"kind", &kind,
		"target", &targetValue,
		"name?", &name,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "task_surface target")
	if err != nil {
		return nil, err
	}
	if name == "" {
		name = kind
	}

	m.taskSurfaceDecls = append(m.taskSurfaceDecls, taskSurfaceDecl{
		name:            name,
		kind:            kind,
		targetProjectID: targetProjectID,
	})
	return starlark.None, nil
}

func (m *coreModuleRuntime) triggerBinding(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var trigger string
	var targetValue starlark.Value
	var matchNamesValue starlark.Value = starlark.NewList(nil)
	var matchTagsValue starlark.Value = starlark.NewList(nil)
	if err := starlark.UnpackArgs("trigger_binding", args, kwargs,
		"trigger", &trigger,
		"target", &targetValue,
		"match_names?", &matchNamesValue,
		"match_tags?", &matchTagsValue,
	); err != nil {
		return nil, err
	}

	targetProjectID, err := m.builder.requireProjectRef(targetValue, "trigger_binding target")
	if err != nil {
		return nil, err
	}
	matchNames, err := starlarkStringList(matchNamesValue, "trigger_binding match_names")
	if err != nil {
		return nil, err
	}
	matchTags, err := starlarkStringList(matchTagsValue, "trigger_binding match_tags")
	if err != nil {
		return nil, err
	}
	if len(matchNames) == 0 && len(matchTags) == 0 {
		return nil, fmt.Errorf("trigger_binding must specify match_names or match_tags")
	}

	m.triggerBindingDecls = append(m.triggerBindingDecls, triggerBindingDecl{
		trigger:         trigger,
		targetProjectID: targetProjectID,
		matchNames:      matchNames,
		matchTags:       matchTags,
	})
	return starlark.None, nil
}

func starlarkStringList(value starlark.Value, label string) ([]string, error) {
	if value == nil || value == starlark.None {
		return nil, nil
	}

	switch value.(type) {
	case *starlark.List, starlark.Tuple:
	default:
		return nil, fmt.Errorf("%s must be a list of strings", label)
	}

	values := []string{}
	iter := value.(starlark.Iterable).Iterate()
	defer iter.Done()
	var item starlark.Value
	for iter.Next(&item) {
		str, ok := starlark.AsString(item)
		if !ok {
			return nil, fmt.Errorf("%s must contain only strings", label)
		}
		values = append(values, str)
	}
	return values, nil
}

func starlarkStringMap(value starlark.Value, label string) (map[string]string, error) {
	if value == nil || value == starlark.None {
		return nil, nil
	}

	dict, ok := value.(*starlark.Dict)
	if !ok {
		return nil, fmt.Errorf("%s must be a dict of strings", label)
	}

	values := make(map[string]string, dict.Len())
	for _, item := range dict.Items() {
		key, ok := starlark.AsString(item[0])
		if !ok {
			return nil, fmt.Errorf("%s keys must be strings", label)
		}
		itemValue, ok := starlark.AsString(item[1])
		if !ok {
			return nil, fmt.Errorf("%s values must be strings", label)
		}
		values[key] = itemValue
	}
	return values, nil
}

func defaultProjectName(path string, id string) string {
	if path == "" || path == "." {
		return id
	}
	parts := strings.Split(path, "/")
	name := parts[len(parts)-1]
	if name == "" || name == "." {
		return id
	}
	return name
}

func isKnownSynthesisPhase(value string) bool {
	switch value {
	case SynthesisPhasePre, SynthesisPhaseMain, SynthesisPhasePost:
		return true
	default:
		return false
	}
}

func ensureNoFactRefs(value any) error {
	switch typed := value.(type) {
	case FactValueRef:
		return fmt.Errorf("fact values must not contain fact_value references")
	case map[string]any:
		for _, nested := range typed {
			if err := ensureNoFactRefs(nested); err != nil {
				return err
			}
		}
	case []any:
		for _, nested := range typed {
			if err := ensureNoFactRefs(nested); err != nil {
				return err
			}
		}
	}
	return nil
}
