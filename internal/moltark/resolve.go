package moltark

import (
	"fmt"
	"sort"
)

type providerInstance struct {
	componentID string
	provider    CapabilityProvider
}

func ResolveModel(model DesiredModel) (ResolvedModel, error) {
	managedByPath := map[string]*ManagedFileSpec{}
	providers := make([]providerInstance, 0, len(model.Components))
	publicProviders := []ProviderBinding{}
	synthesisHooks := []ResolvedSynthesisHook{}
	bootstrapRequirements := []ResolvedBootstrapRequirement{}
	tasks := []ResolvedTask{}
	taskSurfaces := []ResolvedTaskSurface{}

	for _, component := range model.Components {
		for _, file := range component.Files {
			if err := mergeManagedFileIntent(managedByPath, component.ID, file); err != nil {
				return ResolvedModel{}, err
			}
		}
		for _, provider := range component.Providers {
			providers = append(providers, providerInstance{
				componentID: component.ID,
				provider:    provider,
			})
			publicProviders = append(publicProviders, ProviderBinding{
				ComponentID:    component.ID,
				Capability:     provider.Capability,
				ScopeProjectID: provider.ScopeProjectID,
				Attributes:     cloneStringMap(provider.Attributes),
				Lists:          cloneStringSliceMap(provider.Lists),
			})
		}
		for _, hook := range component.SynthesisHooks {
			project, err := requireProject(model, hook.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("synthesis hook %q: %w", component.ID, err)
			}
			synthesisHooks = append(synthesisHooks, ResolvedSynthesisHook{
				ComponentID:       component.ID,
				Phase:             hook.Phase,
				TargetProjectID:   hook.TargetProjectID,
				TargetProjectPath: project.EffectivePath,
				Description:       hook.Description,
			})
		}
		for _, requirement := range component.BootstrapRequirements {
			project, err := requireProject(model, requirement.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("bootstrap requirement %q: %w", component.ID, err)
			}
			bootstrapRequirements = append(bootstrapRequirements, ResolvedBootstrapRequirement{
				ComponentID:       component.ID,
				Tool:              requirement.Tool,
				TargetProjectID:   requirement.TargetProjectID,
				TargetProjectPath: project.EffectivePath,
				Purpose:           requirement.Purpose,
				Strategies:        append([]string(nil), requirement.Strategies...),
			})
		}
		for _, task := range component.Tasks {
			project, err := requireProject(model, task.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("task %q: %w", component.ID, err)
			}
			tasks = append(tasks, ResolvedTask{
				ComponentID:       component.ID,
				Name:              task.Name,
				TargetProjectID:   task.TargetProjectID,
				TargetProjectPath: project.EffectivePath,
				Command:           append([]string(nil), task.Command...),
				Runtime:           task.Runtime,
				Tags:              append([]string(nil), task.Tags...),
			})
		}
		for _, surface := range component.TaskSurfaces {
			project, err := requireProject(model, surface.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("task surface %q: %w", component.ID, err)
			}
			taskSurfaces = append(taskSurfaces, ResolvedTaskSurface{
				ComponentID:       component.ID,
				Name:              surface.Name,
				Kind:              surface.Kind,
				TargetProjectID:   surface.TargetProjectID,
				TargetProjectPath: project.EffectivePath,
			})
		}
	}

	intentBindings := []IntentBinding{}
	for _, component := range model.Components {
		for _, intent := range component.RoutedIntents {
			provider, err := resolveCapabilityProvider(model, providers, intent.Capability, intent.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("%s %q: %w", intent.Kind, component.ID, err)
			}

			intentBindings = append(intentBindings, IntentBinding{
				ComponentID:            component.ID,
				IntentKind:             intent.Kind,
				Capability:             intent.Capability,
				TargetProjectID:        intent.TargetProjectID,
				Attributes:             cloneStringMap(intent.Attributes),
				Lists:                  cloneStringSliceMap(intent.Lists),
				ProviderComponentID:    provider.componentID,
				ProviderScopeProjectID: provider.provider.ScopeProjectID,
			})

			switch intent.Kind {
			case IntentWorkspaceMembersRequest:
				if err := applyWorkspaceMembersIntent(managedByPath, component.ID, provider, intent); err != nil {
					return ResolvedModel{}, err
				}
			case IntentPythonDependencyRequest:
				if err := validatePythonDependencyIntent(provider, intent); err != nil {
					return ResolvedModel{}, err
				}
			default:
				return ResolvedModel{}, fmt.Errorf("unsupported routed intent kind %q", intent.Kind)
			}
		}
	}

	triggerBindings := []ResolvedTriggerBinding{}
	for _, component := range model.Components {
		for _, binding := range component.TriggerBindings {
			project, err := requireProject(model, binding.TargetProjectID)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("trigger binding %q: %w", component.ID, err)
			}

			matches, err := resolveTriggerTasks(model, tasks, binding)
			if err != nil {
				return ResolvedModel{}, fmt.Errorf("trigger binding %q: %w", component.ID, err)
			}

			for _, task := range matches {
				triggerBindings = append(triggerBindings, ResolvedTriggerBinding{
					ComponentID:           component.ID,
					Trigger:               binding.Trigger,
					TargetProjectID:       binding.TargetProjectID,
					TargetProjectPath:     project.EffectivePath,
					MatchNames:            append([]string(nil), binding.MatchNames...),
					MatchTags:             append([]string(nil), binding.MatchTags...),
					TaskComponentID:       task.ComponentID,
					TaskName:              task.Name,
					TaskTargetProjectID:   task.TargetProjectID,
					TaskTargetProjectPath: task.TargetProjectPath,
				})
			}
		}
	}

	managedFiles := make([]ManagedFileSpec, 0, len(managedByPath))
	for _, path := range sortedStringKeys(managedByPath) {
		file := managedByPath[path]
		file.OwnedPaths = uniqueStringsInOrder(file.OwnedPaths)
		file.UserManagedPaths = uniqueStringsInOrder(file.UserManagedPaths)
		file.SourceComponents = uniqueStringsInOrder(file.SourceComponents)
		managedFiles = append(managedFiles, *file)
	}

	sort.Slice(publicProviders, func(i, j int) bool {
		if publicProviders[i].Capability != publicProviders[j].Capability {
			return publicProviders[i].Capability < publicProviders[j].Capability
		}
		if publicProviders[i].ScopeProjectID != publicProviders[j].ScopeProjectID {
			return publicProviders[i].ScopeProjectID < publicProviders[j].ScopeProjectID
		}
		return publicProviders[i].ComponentID < publicProviders[j].ComponentID
	})
	sort.Slice(intentBindings, func(i, j int) bool {
		if intentBindings[i].ComponentID != intentBindings[j].ComponentID {
			return intentBindings[i].ComponentID < intentBindings[j].ComponentID
		}
		if intentBindings[i].Capability != intentBindings[j].Capability {
			return intentBindings[i].Capability < intentBindings[j].Capability
		}
		return intentBindings[i].TargetProjectID < intentBindings[j].TargetProjectID
	})
	sort.Slice(synthesisHooks, func(i, j int) bool {
		if synthesisHooks[i].TargetProjectPath != synthesisHooks[j].TargetProjectPath {
			return synthesisHooks[i].TargetProjectPath < synthesisHooks[j].TargetProjectPath
		}
		if synthesisHooks[i].Phase != synthesisHooks[j].Phase {
			return synthesisHooks[i].Phase < synthesisHooks[j].Phase
		}
		return synthesisHooks[i].ComponentID < synthesisHooks[j].ComponentID
	})
	sort.Slice(bootstrapRequirements, func(i, j int) bool {
		if bootstrapRequirements[i].TargetProjectPath != bootstrapRequirements[j].TargetProjectPath {
			return bootstrapRequirements[i].TargetProjectPath < bootstrapRequirements[j].TargetProjectPath
		}
		if bootstrapRequirements[i].Tool != bootstrapRequirements[j].Tool {
			return bootstrapRequirements[i].Tool < bootstrapRequirements[j].Tool
		}
		return bootstrapRequirements[i].ComponentID < bootstrapRequirements[j].ComponentID
	})
	sort.Slice(tasks, func(i, j int) bool {
		if tasks[i].TargetProjectPath != tasks[j].TargetProjectPath {
			return tasks[i].TargetProjectPath < tasks[j].TargetProjectPath
		}
		if tasks[i].Name != tasks[j].Name {
			return tasks[i].Name < tasks[j].Name
		}
		return tasks[i].ComponentID < tasks[j].ComponentID
	})
	sort.Slice(taskSurfaces, func(i, j int) bool {
		if taskSurfaces[i].TargetProjectPath != taskSurfaces[j].TargetProjectPath {
			return taskSurfaces[i].TargetProjectPath < taskSurfaces[j].TargetProjectPath
		}
		if taskSurfaces[i].Kind != taskSurfaces[j].Kind {
			return taskSurfaces[i].Kind < taskSurfaces[j].Kind
		}
		if taskSurfaces[i].Name != taskSurfaces[j].Name {
			return taskSurfaces[i].Name < taskSurfaces[j].Name
		}
		return taskSurfaces[i].ComponentID < taskSurfaces[j].ComponentID
	})
	sort.Slice(triggerBindings, func(i, j int) bool {
		if triggerBindings[i].Trigger != triggerBindings[j].Trigger {
			return triggerBindings[i].Trigger < triggerBindings[j].Trigger
		}
		if triggerBindings[i].TargetProjectPath != triggerBindings[j].TargetProjectPath {
			return triggerBindings[i].TargetProjectPath < triggerBindings[j].TargetProjectPath
		}
		if triggerBindings[i].TaskTargetProjectPath != triggerBindings[j].TaskTargetProjectPath {
			return triggerBindings[i].TaskTargetProjectPath < triggerBindings[j].TaskTargetProjectPath
		}
		if triggerBindings[i].TaskName != triggerBindings[j].TaskName {
			return triggerBindings[i].TaskName < triggerBindings[j].TaskName
		}
		return triggerBindings[i].ComponentID < triggerBindings[j].ComponentID
	})

	return ResolvedModel{
		ManagedFiles:            managedFiles,
		Providers:               publicProviders,
		SynthesisHooks:          synthesisHooks,
		BootstrapRequirements:   bootstrapRequirements,
		Tasks:                   tasks,
		TaskSurfaces:            taskSurfaces,
		ResolvedTriggerBindings: triggerBindings,
		ResolvedIntents:         intentBindings,
	}, nil
}

func mergeManagedFileIntent(managedByPath map[string]*ManagedFileSpec, componentID string, file StructuredFileSpec) error {
	current, ok := managedByPath[file.Path]
	if !ok {
		current = &ManagedFileSpec{
			Path:             file.Path,
			Format:           file.Format,
			OwnedPaths:       append([]string(nil), file.OwnedPaths...),
			UserManagedPaths: append([]string(nil), file.UserManagedPaths...),
			DesiredValues:    cloneNestedMap(file.DesiredValues),
			SourceComponents: []string{componentID},
		}
		managedByPath[file.Path] = current
		return nil
	}

	if current.Format != file.Format {
		return fmt.Errorf("file %q is claimed with conflicting formats %q and %q", file.Path, current.Format, file.Format)
	}

	for _, ownedPath := range file.OwnedPaths {
		value, _ := lookupPath(file.DesiredValues, ownedPath)
		if existingValue, ok := lookupPath(current.DesiredValues, ownedPath); ok && fingerprintValue(existingValue, true) != fingerprintValue(value, true) {
			return fmt.Errorf("file %q has conflicting desired values for %s", file.Path, ownedPath)
		}
		setPathValue(current.DesiredValues, ownedPath, value)
		current.OwnedPaths = append(current.OwnedPaths, ownedPath)
	}
	current.UserManagedPaths = append(current.UserManagedPaths, file.UserManagedPaths...)
	current.SourceComponents = append(current.SourceComponents, componentID)

	return nil
}

func resolveCapabilityProvider(model DesiredModel, providers []providerInstance, capability string, projectID string) (providerInstance, error) {
	scopeChain, err := projectScopeChain(model, projectID)
	if err != nil {
		return providerInstance{}, err
	}

	for _, scopeProjectID := range scopeChain {
		matches := []providerInstance{}
		for _, provider := range providers {
			if provider.provider.Capability == capability && provider.provider.ScopeProjectID == scopeProjectID {
				matches = append(matches, provider)
			}
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
		if len(matches) > 1 {
			return providerInstance{}, fmt.Errorf("capability %q is ambiguous for project %q at scope %q", capability, projectID, scopeProjectID)
		}
	}

	globalMatches := []providerInstance{}
	for _, provider := range providers {
		if provider.provider.Capability == capability && provider.provider.ScopeProjectID == "" {
			globalMatches = append(globalMatches, provider)
		}
	}
	if len(globalMatches) == 1 {
		return globalMatches[0], nil
	}
	if len(globalMatches) > 1 {
		return providerInstance{}, fmt.Errorf("capability %q is ambiguous for project %q at global scope", capability, projectID)
	}

	return providerInstance{}, fmt.Errorf("no provider for capability %q near project %q", capability, projectID)
}

func projectScopeChain(model DesiredModel, projectID string) ([]string, error) {
	project, err := requireProject(model, projectID)
	if err != nil {
		return nil, err
	}

	chain := []string{}
	seen := map[string]struct{}{}
	current := project
	for current != nil {
		if _, ok := seen[current.ID]; ok {
			return nil, fmt.Errorf("project parent cycle detected at %q", current.ID)
		}
		seen[current.ID] = struct{}{}
		chain = append(chain, current.ID)
		if current.ParentID == "" {
			break
		}
		current = model.projectByID(current.ParentID)
		if current == nil {
			return nil, fmt.Errorf("project parent %q is not declared", project.ParentID)
		}
	}
	return chain, nil
}

func applyWorkspaceMembersIntent(managedByPath map[string]*ManagedFileSpec, componentID string, provider providerInstance, intent RoutedIntentSpec) error {
	filePath := provider.provider.Attributes[ProviderAttrFilePath]
	ownedPath := provider.provider.Attributes[ProviderAttrOwnedPath]
	if filePath == "" || ownedPath == "" {
		return fmt.Errorf("workspace manager provider on component %q is missing file_path or owned_path", provider.componentID)
	}

	current, ok := managedByPath[filePath]
	if !ok {
		current = &ManagedFileSpec{
			Path:             filePath,
			Format:           FileFormatTOML,
			DesiredValues:    map[string]any{},
			SourceComponents: []string{},
		}
		managedByPath[filePath] = current
	}

	memberPaths := append([]string(nil), intent.Lists[IntentListMemberPaths]...)
	if existingValue, ok := lookupPath(current.DesiredValues, ownedPath); ok && fingerprintValue(existingValue, true) != fingerprintValue(memberPaths, true) {
		return fmt.Errorf("file %q has conflicting desired values for %s", filePath, ownedPath)
	}

	setPathValue(current.DesiredValues, ownedPath, memberPaths)
	current.OwnedPaths = append(current.OwnedPaths, ownedPath)
	current.SourceComponents = append(current.SourceComponents, componentID, provider.componentID)
	if current.Format == "" {
		current.Format = FileFormatTOML
	}

	return nil
}

func validatePythonDependencyIntent(provider providerInstance, intent RoutedIntentSpec) error {
	requirement := intent.Attributes[IntentAttrRequirement]
	if requirement == "" {
		return fmt.Errorf("python package manager request is missing requirement")
	}

	dependencyPath := provider.provider.Attributes[ProviderAttrDependencyPath]
	if dependencyPath == "" {
		return fmt.Errorf("python package manager provider on component %q is missing dependency_path", provider.componentID)
	}

	artifactKinds := provider.provider.Lists[ProviderListArtifactKinds]
	if len(artifactKinds) == 0 {
		return fmt.Errorf("python package manager provider on component %q is missing artifact kinds", provider.componentID)
	}

	if !stringSliceContains(artifactKinds, ArtifactKindPyPI) {
		return fmt.Errorf("python package manager provider on component %q does not accept %q artifacts", provider.componentID, ArtifactKindPyPI)
	}

	return nil
}

func resolveTriggerTasks(model DesiredModel, tasks []ResolvedTask, binding TriggerBindingSpec) ([]ResolvedTask, error) {
	if len(binding.MatchNames) == 0 && len(binding.MatchTags) == 0 {
		return nil, fmt.Errorf("must specify match_names or match_tags")
	}

	matches := []ResolvedTask{}
	for _, task := range tasks {
		withinScope, err := projectWithinSubtree(model, task.TargetProjectID, binding.TargetProjectID)
		if err != nil {
			return nil, err
		}
		if !withinScope {
			continue
		}
		if !taskMatches(binding, task) {
			continue
		}
		matches = append(matches, task)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no tasks matched trigger %q in project subtree %q", binding.Trigger, binding.TargetProjectID)
	}

	return matches, nil
}

func taskMatches(binding TriggerBindingSpec, task ResolvedTask) bool {
	if len(binding.MatchNames) > 0 && !stringSliceContains(binding.MatchNames, task.Name) {
		return false
	}
	if len(binding.MatchTags) > 0 && !stringSliceContainsAll(task.Tags, binding.MatchTags) {
		return false
	}
	return true
}

func projectWithinSubtree(model DesiredModel, candidateProjectID string, ancestorProjectID string) (bool, error) {
	current, err := requireProject(model, candidateProjectID)
	if err != nil {
		return false, err
	}

	for current != nil {
		if current.ID == ancestorProjectID {
			return true, nil
		}
		if current.ParentID == "" {
			return false, nil
		}
		current = model.projectByID(current.ParentID)
		if current == nil {
			return false, fmt.Errorf("project parent %q is not declared", candidateProjectID)
		}
	}

	return false, nil
}

func requireProject(model DesiredModel, projectID string) (*ProjectSpec, error) {
	project := model.projectByID(projectID)
	if project == nil {
		return nil, fmt.Errorf("project %q is not declared", projectID)
	}
	return project, nil
}

func cloneNestedMap(values map[string]any) map[string]any {
	cloned := map[string]any{}
	for key, value := range values {
		switch typed := value.(type) {
		case map[string]any:
			cloned[key] = cloneNestedMap(typed)
		case []string:
			cloned[key] = append([]string(nil), typed...)
		default:
			cloned[key] = value
		}
	}
	return cloned
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func cloneStringSliceMap(values map[string][]string) map[string][]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(values))
	for key, value := range values {
		cloned[key] = append([]string(nil), value...)
	}
	return cloned
}

func sortedStringKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func uniqueStringsInOrder(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func stringSliceContainsAll(values []string, targets []string) bool {
	for _, target := range targets {
		if !stringSliceContains(values, target) {
			return false
		}
	}
	return true
}
