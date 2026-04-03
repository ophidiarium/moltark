package moltark

import (
	"fmt"

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
		"synthesis_hook":        starlark.NewBuiltin("synthesis_hook", m.synthesisHook),
		"bootstrap_requirement": starlark.NewBuiltin("bootstrap_requirement", m.bootstrapRequirement),
		"python_dependency":     starlark.NewBuiltin("python_dependency", m.pythonDependency),
		"task":                  starlark.NewBuiltin("task", m.task),
		"task_surface":          starlark.NewBuiltin("task_surface", m.taskSurface),
		"trigger_binding":       starlark.NewBuiltin("trigger_binding", m.triggerBinding),
	})
}

func (m *coreModuleRuntime) BuildComponents(_ DesiredModel) ([]ComponentSpec, error) {
	components := make([]ComponentSpec, 0, len(m.synthesisHookDecls)+len(m.bootstrapRequirementDecls)+len(m.pythonDependencyDecls)+len(m.taskDecls)+len(m.taskSurfaceDecls)+len(m.triggerBindingDecls))

	for _, decl := range m.synthesisHookDecls {
		components = append(components, ComponentSpec{
			ID:              m.builder.nextComponentName(),
			Kind:            "synthesis_hook",
			Module:          ModuleSourceCore,
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

	iterable, ok := value.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("%s must be a list of strings", label)
	}

	values := []string{}
	iter := iterable.Iterate()
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

func isKnownSynthesisPhase(value string) bool {
	switch value {
	case SynthesisPhasePre, SynthesisPhaseMain, SynthesisPhasePost:
		return true
	default:
		return false
	}
}
