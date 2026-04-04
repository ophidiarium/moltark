package moltark

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
)

type Service struct{}

func NewService() Service {
	return Service{}
}

func (Service) Plan(root string) (Plan, error) {
	desired, err := LoadDesiredModel(root)
	if err != nil {
		return Plan{}, err
	}

	resolved, err := ResolveModel(desired)
	if err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Desired:  desired,
		Resolved: resolved,
		fileDocs: map[string]fileDocument{},
	}
	plan.gitattributesRaw, plan.gitattributesExists, err = readOptionalFile(filepath.Join(root, GitattributesFileName))
	if err != nil {
		return Plan{}, err
	}

	stateDoc, err := loadState(root)
	if err != nil {
		return Plan{}, err
	}
	plan.stateRaw = stateDoc.Raw
	plan.State = stateDoc.State

	changes := make([]Change, 0, len(resolved.ManagedFiles)*8+4)
	for _, managedFile := range resolved.ManagedFiles {
		doc, err := loadStructuredDocument(filepath.Join(root, managedFile.Path), managedFile.Format)
		if err != nil {
			return Plan{}, fmt.Errorf("load %s: %w", managedFile.Path, err)
		}
		plan.fileDocs[managedFile.Path] = doc

		if !doc.Exists {
			changes = append(changes, Change{
				Status:  ChangeCreate,
				File:    managedFile.Path,
				Reason:  ReasonBootstrap,
				Summary: fmt.Sprintf("create %s", managedFile.Path),
			})
		}

		stateFile := stateManagedFile(plan.State, managedFile.Path)
		for _, ownedPath := range managedFile.OwnedPaths {
			desiredValue, _ := lookupStructuredValue(managedFile.DesiredValues, managedFile.Format, ownedPath)
			actualValue, actualPresent := lookupStructuredValue(doc.Values, managedFile.Format, ownedPath)
			ownerComponentID := managedFile.OwnedPathOwners[ownedPath]
			desiredVersion := managedFile.OwnedPathVersions[ownedPath]
			change, err := classifyPath(
				managedFile.Format,
				managedFile.Path,
				ownedPath,
				ownerComponentID,
				desiredVersion,
				desiredValue,
				actualValue,
				actualPresent,
				stateFile,
				plan.State,
			)
			if err != nil {
				return Plan{}, err
			}
			changes = append(changes, change)
		}

		for _, userManagedPath := range managedFile.UserManagedPaths {
			if value, ok := lookupStructuredValue(doc.Values, managedFile.Format, userManagedPath); ok && value != nil {
				changes = append(changes, Change{
					Status:  ChangeNoOp,
					File:    managedFile.Path,
					Path:    userManagedPath,
					Reason:  ReasonAdoption,
					Summary: noOpSummary(userManagedPath),
				})
			}
		}
	}

	block, blockExists := currentManagedGitattributesBlock(plan.gitattributesRaw)
	gitattributesState := stateManagedFile(plan.State, GitattributesFileName)
	if !plan.gitattributesExists {
		changes = append(changes, Change{
			Status:  ChangeCreate,
			File:    GitattributesFileName,
			Reason:  ReasonBootstrap,
			Summary: "create .gitattributes",
		})
	}
	change, err := classifyPath(
		FileFormatTOML,
		GitattributesFileName,
		"moltark.block",
		"",
		"",
		managedGitattributesBlock(),
		block,
		blockExists,
		gitattributesState,
		plan.State,
	)
	if err != nil {
		return Plan{}, err
	}
	changes = append(changes, change)

	if !hasConflict(changes) {
		state, err := buildState(plan.Desired, plan.Resolved)
		if err != nil {
			return Plan{}, err
		}
		stateRaw, err := renderState(state)
		if err != nil {
			return Plan{}, err
		}
		plan.State = &state

		if !stateDoc.Exists {
			changes = append(changes, Change{
				Status:  ChangeCreate,
				File:    filepath.ToSlash(filepath.Join(StateDirName, StateFileName)),
				Reason:  ReasonBootstrap,
				Summary: "initialize .moltark/state.json",
			})
		} else if stateRaw != stateDoc.Raw {
			reason := reasonForStateUpdate(stateDoc.State, &state)
			changes = append(changes, Change{
				Status:  ChangeUpdate,
				File:    filepath.ToSlash(filepath.Join(StateDirName, StateFileName)),
				Reason:  reason,
				Summary: "update .moltark/state.json",
			})
		} else {
			changes = append(changes, Change{
				Status:  ChangeNoOp,
				File:    filepath.ToSlash(filepath.Join(StateDirName, StateFileName)),
				Reason:  ReasonAdoption,
				Summary: "no change to .moltark/state.json",
			})
		}
	}

	plan.Changes = compactChanges(changes)
	plan.Summary = summarizeChanges(plan.Changes)
	return plan, nil
}

func (Service) Apply(root string, plan Plan) (ApplyResult, error) {
	if plan.HasConflicts() {
		return ApplyResult{}, fmt.Errorf("cannot apply a plan with conflicts")
	}

	freshPlan, err := (Service{}).Plan(root)
	if err != nil {
		return ApplyResult{}, err
	}
	if !sameApplyIntent(plan, freshPlan) {
		return ApplyResult{}, fmt.Errorf("repository changed since planning; rerun `moltark plan` and review the updated changes")
	}
	plan = freshPlan

	wrote := []string{}
	for _, managedFile := range plan.Resolved.ManagedFiles {
		doc := plan.fileDocs[managedFile.Path]
		body, err := mutateStructuredFile(doc.Raw, managedFile)
		if err != nil {
			return ApplyResult{}, err
		}
		if body == ensureTrailingNewline(doc.Raw) {
			continue
		}
		if err := writeRepoFile(root, managedFile.Path, body); err != nil {
			return ApplyResult{}, err
		}
		wrote = append(wrote, managedFile.Path)
	}

	gitattributesBody := mutateGitattributes(plan.gitattributesRaw)
	if gitattributesBody != ensureTrailingNewline(plan.gitattributesRaw) {
		if err := writeRepoFile(root, GitattributesFileName, gitattributesBody); err != nil {
			return ApplyResult{}, err
		}
		wrote = append(wrote, GitattributesFileName)
	}

	state, err := buildState(plan.Desired, plan.Resolved)
	if err != nil {
		return ApplyResult{}, err
	}
	stateBody, err := renderState(state)
	if err != nil {
		return ApplyResult{}, err
	}
	if stateBody != plan.stateRaw {
		if err := writeStateFile(root, stateBody); err != nil {
			return ApplyResult{}, err
		}
		wrote = append(wrote, filepath.ToSlash(filepath.Join(StateDirName, StateFileName)))
	}

	return ApplyResult{
		Plan:  plan,
		Wrote: wrote,
	}, nil
}

func sameApplyIntent(left Plan, right Plan) bool {
	return reflect.DeepEqual(left.Desired, right.Desired) &&
		reflect.DeepEqual(left.Resolved, right.Resolved) &&
		reflect.DeepEqual(left.State, right.State) &&
		reflect.DeepEqual(actionableChanges(left.Changes), actionableChanges(right.Changes)) &&
		reflect.DeepEqual(actionableSummary(left.Summary), actionableSummary(right.Summary))
}

func actionableChanges(changes []Change) []Change {
	filtered := make([]Change, 0, len(changes))
	for _, change := range changes {
		if change.Status == ChangeNoOp {
			continue
		}
		filtered = append(filtered, change)
	}
	return filtered
}

func actionableSummary(summary PlanSummary) PlanSummary {
	summary.NoOp = 0
	return summary
}

func reasonForStateUpdate(previous *State, next *State) ChangeReason {
	if previous == nil || next == nil {
		return ReasonBootstrap
	}
	if modelHasTemplateUpgrade(previous.LastAppliedModel, next.LastAppliedModel) {
		return ReasonTemplateUpgrade
	}
	return ReasonDesiredState
}

func modelHasTemplateUpgrade(previous ModelSummary, next ModelSummary) bool {
	for _, component := range next.Components {
		if component.Version == "" {
			continue
		}
		lastVersion, ok := previous.componentVersion(component.ID)
		if ok && lastVersion != "" && lastVersion != component.Version {
			return true
		}
	}
	return false
}

func (s Service) Show(root string) (ShowReport, error) {
	plan, err := s.Plan(root)
	if err != nil {
		return ShowReport{}, err
	}

	currentValues := map[string]map[string]string{}
	for _, managedFile := range plan.Resolved.ManagedFiles {
		doc := plan.fileDocs[managedFile.Path]
		currentValues[managedFile.Path] = map[string]string{}
		for _, ownedPath := range managedFile.OwnedPaths {
			value, ok := lookupStructuredValue(doc.Values, managedFile.Format, ownedPath)
			currentValues[managedFile.Path][ownedPath] = renderDisplayValue(managedFile.Format, value, ok)
		}
	}

	block, ok := currentManagedGitattributesBlock(plan.gitattributesRaw)
	currentValues[GitattributesFileName] = map[string]string{
		"moltark.block": renderDisplayValue(FileFormatTOML, block, ok),
	}

	return ShowReport{
		Desired:            plan.Desired,
		Resolved:           plan.Resolved,
		State:              plan.State,
		CurrentOwnedValues: currentValues,
	}, nil
}

func (s Service) Doctor(root string) (DoctorReport, error) {
	plan, err := s.Plan(root)
	if err != nil {
		return DoctorReport{}, err
	}

	messages := []string{}
	hasIssues := false
	if plan.HasConflicts() {
		hasIssues = true
		messages = append(messages, "conflicts require manual resolution before apply")
	}
	if plan.Summary.Drift > 0 {
		hasIssues = true
		messages = append(messages, "drift detected in Moltark-owned surfaces")
	}
	if !hasIssues {
		messages = append(messages, "repository is healthy")
	}

	return DoctorReport{
		Plan:      plan,
		HasIssues: hasIssues,
		Messages:  messages,
	}, nil
}

func loadStructuredDocument(path string, format string) (fileDocument, error) {
	raw, exists, err := readOptionalFile(path)
	if err != nil {
		return fileDocument{}, err
	}

	doc := fileDocument{
		Raw:    raw,
		Exists: exists,
		Values: map[string]any{},
	}
	if !exists {
		return doc, nil
	}

	switch format {
	case FileFormatTOML:
		values, err := parseTomlValues([]byte(raw))
		if err != nil {
			return fileDocument{}, err
		}
		doc.Values = values
	case FileFormatJSON:
		values, err := parseJSONValues([]byte(raw))
		if err != nil {
			return fileDocument{}, err
		}
		doc.Values = values
	case FileFormatYAML:
		values, err := parseYAMLValues([]byte(raw))
		if err != nil {
			return fileDocument{}, err
		}
		doc.Values = values
	default:
		return fileDocument{}, fmt.Errorf("unsupported file format %q", format)
	}
	return doc, nil
}

func mutateStructuredFile(raw string, file ManagedFileSpec) (string, error) {
	switch file.Format {
	case FileFormatTOML:
		return mutateTOMLFile(raw, file.DesiredValues, file.OwnedPaths)
	case FileFormatJSON:
		return mutateJSONFile(raw, file.DesiredValues, file.OwnedPaths)
	case FileFormatYAML:
		return mutateYAMLFile(raw, file.DesiredValues, file.OwnedPaths)
	default:
		return "", fmt.Errorf("unsupported file format %q", file.Format)
	}
}

func writeRepoFile(root string, relativePath string, body string) error {
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(body), 0o644)
}

func readOptionalFile(path string) (string, bool, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, err
	}
	return string(raw), true, nil
}
