package moltark

import (
	"fmt"
	"path/filepath"
)

func (s Service) BuildPipeline(root string) (Pipeline, error) {
	evaluate, err := evaluatePhase(root)
	if err != nil {
		return Pipeline{}, err
	}

	resolve, err := resolvePhase(evaluate)
	if err != nil {
		return Pipeline{}, err
	}

	inspect, docs, stateRaw, gitattributesRaw, err := inspectPhase(root, resolve)
	if err != nil {
		return Pipeline{}, err
	}

	persist, nextStateRaw, err := persistPhase(evaluate, resolve)
	if err != nil {
		return Pipeline{}, err
	}

	plan, err := planningPhase(resolve, inspect, persist, docs, stateRaw, nextStateRaw, gitattributesRaw)
	if err != nil {
		return Pipeline{}, err
	}

	return Pipeline{
		Root:             root,
		Evaluate:         evaluate,
		Resolve:          resolve,
		Inspect:          inspect,
		Persist:          persist,
		Plan:             plan,
		fileDocs:         docs,
		stateRaw:         stateRaw,
		nextStateRaw:     nextStateRaw,
		gitattributesRaw: gitattributesRaw,
	}, nil
}

func (p Pipeline) PlanResult() Plan {
	return Plan{
		Desired:   p.Evaluate.Desired,
		Resolved:  p.Resolve.Resolved,
		NextState: p.Persist.NextState,
		Changes:   append([]Change(nil), p.Plan.Changes...),
		Summary:   p.Plan.Summary,
	}
}

func evaluatePhase(root string) (EvaluationPhase, error) {
	desired, err := LoadDesiredModel(root)
	if err != nil {
		return EvaluationPhase{}, err
	}
	return EvaluationPhase{Desired: desired}, nil
}

func resolvePhase(evaluate EvaluationPhase) (ResolutionPhase, error) {
	resolved, err := ResolveModel(evaluate.Desired)
	if err != nil {
		return ResolutionPhase{}, err
	}
	return ResolutionPhase{Resolved: resolved}, nil
}

func inspectPhase(root string, resolve ResolutionPhase) (InspectionPhase, map[string]fileDocument, string, string, error) {
	fileDocs := make(map[string]fileDocument, len(resolve.Resolved.ManagedFiles))
	phase := InspectionPhase{
		StateFile:       filepath.ToSlash(filepath.Join(StateDirName, StateFileName)),
		StructuredFiles: make([]InspectedStructuredFile, 0, len(resolve.Resolved.ManagedFiles)),
	}

	gitattributesRaw, gitattributesExists, err := readOptionalFile(filepath.Join(root, GitattributesFileName))
	if err != nil {
		return InspectionPhase{}, nil, "", "", err
	}

	stateDoc, err := loadState(root)
	if err != nil {
		return InspectionPhase{}, nil, "", "", err
	}
	phase.CurrentState = stateDoc.State

	for _, managedFile := range resolve.Resolved.ManagedFiles {
		doc, err := loadStructuredDocument(filepath.Join(root, managedFile.Path), managedFile.Format)
		if err != nil {
			return InspectionPhase{}, nil, "", "", fmt.Errorf("load %s: %w", managedFile.Path, err)
		}
		fileDocs[managedFile.Path] = doc

		inspected := InspectedStructuredFile{
			Path:        managedFile.Path,
			Format:      managedFile.Format,
			Exists:      doc.Exists,
			OwnedValues: map[string]string{},
		}
		for _, ownedPath := range managedFile.OwnedPaths {
			value, ok := lookupStructuredValue(doc.Values, managedFile.Format, ownedPath)
			inspected.OwnedValues[ownedPath] = renderDisplayValue(managedFile.Format, value, ok)
		}
		if len(inspected.OwnedValues) == 0 {
			inspected.OwnedValues = nil
		}

		if len(managedFile.UserManagedPaths) > 0 {
			inspected.UserManagedValues = map[string]string{}
			for _, userManagedPath := range managedFile.UserManagedPaths {
				value, ok := lookupStructuredValue(doc.Values, managedFile.Format, userManagedPath)
				if ok && value != nil {
					inspected.UserManagedValues[userManagedPath] = renderDisplayValue(managedFile.Format, value, true)
				}
			}
			if len(inspected.UserManagedValues) == 0 {
				inspected.UserManagedValues = nil
			}
		}

		phase.StructuredFiles = append(phase.StructuredFiles, inspected)
	}

	block, ok := currentManagedGitattributesBlock(gitattributesRaw)
	phase.Gitattributes = GitattributesInspection{
		Exists:       gitattributesExists,
		ManagedBlock: renderDisplayValue(FileFormatTOML, block, ok),
	}

	return phase, fileDocs, stateDoc.Raw, gitattributesRaw, nil
}

func persistPhase(evaluate EvaluationPhase, resolve ResolutionPhase) (PersistPhase, string, error) {
	nextState, err := buildState(evaluate.Desired, resolve.Resolved)
	if err != nil {
		return PersistPhase{}, "", err
	}
	nextStateRaw, err := renderState(nextState)
	if err != nil {
		return PersistPhase{}, "", err
	}
	return PersistPhase{
		StateFile: filepath.ToSlash(filepath.Join(StateDirName, StateFileName)),
		NextState: &nextState,
	}, nextStateRaw, nil
}

func planningPhase(
	resolve ResolutionPhase,
	inspect InspectionPhase,
	persist PersistPhase,
	docs map[string]fileDocument,
	stateRaw string,
	nextStateRaw string,
	gitattributesRaw string,
) (PlanningPhase, error) {
	changes := make([]Change, 0, len(resolve.Resolved.ManagedFiles)*8+4)
	for _, managedFile := range resolve.Resolved.ManagedFiles {
		doc, ok := docs[managedFile.Path]
		if !ok {
			return PlanningPhase{}, fmt.Errorf("inspection missing structured file %s", managedFile.Path)
		}

		if !doc.Exists {
			changes = append(changes, Change{
				Status:  ChangeCreate,
				File:    managedFile.Path,
				Reason:  ReasonBootstrap,
				Summary: fmt.Sprintf("create %s", managedFile.Path),
			})
		}

		stateFile := stateManagedFile(inspect.CurrentState, managedFile.Path)
		for _, ownedPath := range managedFile.OwnedPaths {
			desiredValue, err := requireStructuredValue(managedFile.DesiredValues, managedFile.Format, ownedPath)
			if err != nil {
				return PlanningPhase{}, fmt.Errorf("plan %s %s: %w", managedFile.Path, ownedPath, err)
			}
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
				inspect.CurrentState,
			)
			if err != nil {
				return PlanningPhase{}, err
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

	block, blockExists := currentManagedGitattributesBlock(gitattributesRaw)
	gitattributesState := stateManagedFile(inspect.CurrentState, GitattributesFileName)
	if !inspect.Gitattributes.Exists {
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
		inspect.CurrentState,
	)
	if err != nil {
		return PlanningPhase{}, err
	}
	changes = append(changes, change)

	if !hasConflict(changes) && persist.NextState != nil {
		if inspect.CurrentState == nil {
			changes = append(changes, Change{
				Status:  ChangeCreate,
				File:    persist.StateFile,
				Reason:  ReasonBootstrap,
				Summary: "initialize .moltark/state.json",
			})
		} else {
			if nextStateRaw != stateRaw {
				reason := reasonForStateUpdate(inspect.CurrentState, persist.NextState)
				changes = append(changes, Change{
					Status:  ChangeUpdate,
					File:    persist.StateFile,
					Reason:  reason,
					Summary: "update .moltark/state.json",
				})
			} else {
				changes = append(changes, Change{
					Status:  ChangeNoOp,
					File:    persist.StateFile,
					Reason:  ReasonAdoption,
					Summary: "no change to .moltark/state.json",
				})
			}
		}
	}

	compacted := compactChanges(changes)
	return PlanningPhase{
		Changes: compacted,
		Summary: summarizeChanges(compacted),
	}, nil
}
