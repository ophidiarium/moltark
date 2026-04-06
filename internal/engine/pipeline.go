package engine

import (
	"fmt"
	"path/filepath"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
	"github.com/ophidiarium/moltark/internal/module"
)

type Pipeline struct {
	Root     string                `json:"-"`
	Evaluate model.EvaluationPhase `json:"evaluate"`
	Resolve  model.ResolutionPhase `json:"resolve"`
	Inspect  model.InspectionPhase `json:"inspect"`
	Persist  model.PersistPhase    `json:"persist"`
	Plan     model.PlanningPhase   `json:"plan"`

	fileDocs         map[string]model.FileDocument
	stateRaw         string
	nextStateRaw     string
	gitattributesRaw string
}

func (s Service) BuildPipeline(root string) (Pipeline, error) {
	evaluate, err := evaluatePhase(root)
	if err != nil {
		return Pipeline{}, err
	}

	resolve, err := resolvePhase(evaluate)
	if err != nil {
		return Pipeline{}, err
	}

	inspection, err := inspectPhase(root, resolve)
	if err != nil {
		return Pipeline{}, err
	}

	persist, nextStateRaw, err := persistPhase(evaluate, resolve)
	if err != nil {
		return Pipeline{}, err
	}

	plan, err := planningPhase(resolve, inspection.Phase, persist, inspection.FileDocs, inspection.StateRaw, nextStateRaw, inspection.GitattributesRaw)
	if err != nil {
		return Pipeline{}, err
	}

	return Pipeline{
		Root:             root,
		Evaluate:         evaluate,
		Resolve:          resolve,
		Inspect:          inspection.Phase,
		Persist:          persist,
		Plan:             plan,
		fileDocs:         inspection.FileDocs,
		stateRaw:         inspection.StateRaw,
		nextStateRaw:     nextStateRaw,
		gitattributesRaw: inspection.GitattributesRaw,
	}, nil
}

func (p Pipeline) PlanResult() model.Plan {
	return model.Plan{
		Desired:   p.Evaluate.Desired,
		Resolved:  p.Resolve.Resolved,
		NextState: p.Persist.NextState,
		Changes:   append([]model.Change(nil), p.Plan.Changes...),
		Summary:   p.Plan.Summary,
	}
}

func evaluatePhase(root string) (model.EvaluationPhase, error) {
	desired, err := module.LoadDesiredModel(root)
	if err != nil {
		return model.EvaluationPhase{}, err
	}
	return model.EvaluationPhase{Desired: desired}, nil
}

func resolvePhase(evaluate model.EvaluationPhase) (model.ResolutionPhase, error) {
	resolved, err := ResolveModel(evaluate.Desired)
	if err != nil {
		return model.ResolutionPhase{}, err
	}
	return model.ResolutionPhase{Resolved: resolved}, nil
}

type inspectionResult struct {
	Phase            model.InspectionPhase
	FileDocs         map[string]model.FileDocument
	StateRaw         string
	GitattributesRaw string
}

func inspectPhase(root string, resolve model.ResolutionPhase) (inspectionResult, error) {
	fileDocs := make(map[string]model.FileDocument, len(resolve.Resolved.ManagedFiles))
	phase := model.InspectionPhase{
		StateFile:       filepath.ToSlash(filepath.Join(model.StateDirName, model.StateFileName)),
		StructuredFiles: make([]model.InspectedStructuredFile, 0, len(resolve.Resolved.ManagedFiles)),
	}

	gitattributesRaw, gitattributesExists, err := readOptionalFile(filepath.Join(root, model.GitattributesFileName))
	if err != nil {
		return inspectionResult{}, err
	}

	stateDoc, err := loadState(root)
	if err != nil {
		return inspectionResult{}, err
	}
	phase.CurrentState = stateDoc.State

	for _, managedFile := range resolve.Resolved.ManagedFiles {
		doc, err := loadStructuredDocument(filepath.Join(root, managedFile.Path), managedFile.Format)
		if err != nil {
			return inspectionResult{}, fmt.Errorf("load %s: %w", managedFile.Path, err)
		}
		fileDocs[managedFile.Path] = doc

		inspected := model.InspectedStructuredFile{
			Path:        managedFile.Path,
			Format:      managedFile.Format,
			Exists:      doc.Exists,
			OwnedValues: map[string]any{},
		}
		for _, ownedPath := range managedFile.OwnedPaths {
			value, ok := filefmt.LookupStructuredValue(doc.Values, managedFile.Format, ownedPath)
			if ok {
				inspected.OwnedValues[ownedPath] = value
			}
		}
		if len(inspected.OwnedValues) == 0 {
			inspected.OwnedValues = nil
		}

		if len(managedFile.UserManagedPaths) > 0 {
			inspected.UserManagedValues = map[string]any{}
			for _, userManagedPath := range managedFile.UserManagedPaths {
				value, ok := filefmt.LookupStructuredValue(doc.Values, managedFile.Format, userManagedPath)
				if ok && value != nil {
					inspected.UserManagedValues[userManagedPath] = value
				}
			}
			if len(inspected.UserManagedValues) == 0 {
				inspected.UserManagedValues = nil
			}
		}

		phase.StructuredFiles = append(phase.StructuredFiles, inspected)
	}

	block, ok := filefmt.CurrentManagedGitattributesBlock(gitattributesRaw)
	phase.Gitattributes = model.GitattributesInspection{
		Exists:       gitattributesExists,
		ManagedBlock: renderDisplayValue(model.FileFormatTOML, block, ok),
	}

	return inspectionResult{
		Phase:            phase,
		FileDocs:         fileDocs,
		StateRaw:         stateDoc.Raw,
		GitattributesRaw: gitattributesRaw,
	}, nil
}

func persistPhase(evaluate model.EvaluationPhase, resolve model.ResolutionPhase) (model.PersistPhase, string, error) {
	nextState, err := buildState(evaluate.Desired, resolve.Resolved)
	if err != nil {
		return model.PersistPhase{}, "", err
	}
	nextStateRaw, err := renderState(nextState)
	if err != nil {
		return model.PersistPhase{}, "", err
	}
	return model.PersistPhase{
		StateFile: filepath.ToSlash(filepath.Join(model.StateDirName, model.StateFileName)),
		NextState: &nextState,
	}, nextStateRaw, nil
}

func planningPhase(
	resolve model.ResolutionPhase,
	inspect model.InspectionPhase,
	persist model.PersistPhase,
	docs map[string]model.FileDocument,
	stateRaw string,
	nextStateRaw string,
	gitattributesRaw string,
) (model.PlanningPhase, error) {
	changes := make([]model.Change, 0, len(resolve.Resolved.ManagedFiles)*8+4)
	for _, managedFile := range resolve.Resolved.ManagedFiles {
		doc, ok := docs[managedFile.Path]
		if !ok {
			return model.PlanningPhase{}, fmt.Errorf("inspection missing structured file %s", managedFile.Path)
		}

		if !doc.Exists {
			changes = append(changes, model.Change{
				Status:  model.ChangeCreate,
				File:    managedFile.Path,
				Reason:  model.ReasonBootstrap,
				Summary: fmt.Sprintf("create %s", managedFile.Path),
			})
		}

		stateFile := stateManagedFile(inspect.CurrentState, managedFile.Path)
		for _, ownedPath := range managedFile.OwnedPaths {
			desiredValue, err := filefmt.RequireStructuredValue(managedFile.DesiredValues, managedFile.Format, ownedPath)
			if err != nil {
				return model.PlanningPhase{}, fmt.Errorf("plan %s %s: %w", managedFile.Path, ownedPath, err)
			}
			actualValue, actualPresent := filefmt.LookupStructuredValue(doc.Values, managedFile.Format, ownedPath)
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
				return model.PlanningPhase{}, err
			}
			changes = append(changes, change)
		}

		for _, userManagedPath := range managedFile.UserManagedPaths {
			if value, ok := filefmt.LookupStructuredValue(doc.Values, managedFile.Format, userManagedPath); ok && value != nil {
				changes = append(changes, model.Change{
					Status:  model.ChangeNoOp,
					File:    managedFile.Path,
					Path:    userManagedPath,
					Reason:  model.ReasonAdoption,
					Summary: noOpSummary(userManagedPath),
				})
			}
		}
	}

	block, blockExists := filefmt.CurrentManagedGitattributesBlock(gitattributesRaw)
	gitattributesState := stateManagedFile(inspect.CurrentState, model.GitattributesFileName)
	if !inspect.Gitattributes.Exists {
		changes = append(changes, model.Change{
			Status:  model.ChangeCreate,
			File:    model.GitattributesFileName,
			Reason:  model.ReasonBootstrap,
			Summary: "create .gitattributes",
		})
	}
	change, err := classifyPath(
		model.FileFormatTOML,
		model.GitattributesFileName,
		"moltark.block",
		"",
		"",
		filefmt.ManagedGitattributesBlock(),
		block,
		blockExists,
		gitattributesState,
		inspect.CurrentState,
	)
	if err != nil {
		return model.PlanningPhase{}, err
	}
	changes = append(changes, change)

	if !hasConflict(changes) && persist.NextState != nil {
		if inspect.CurrentState == nil {
			changes = append(changes, model.Change{
				Status:  model.ChangeCreate,
				File:    persist.StateFile,
				Reason:  model.ReasonBootstrap,
				Summary: "initialize .moltark/state.json",
			})
		} else {
			if nextStateRaw != stateRaw {
				reason := reasonForStateUpdate(inspect.CurrentState, persist.NextState)
				changes = append(changes, model.Change{
					Status:  model.ChangeUpdate,
					File:    persist.StateFile,
					Reason:  reason,
					Summary: "update .moltark/state.json",
				})
			} else {
				changes = append(changes, model.Change{
					Status:  model.ChangeNoOp,
					File:    persist.StateFile,
					Reason:  model.ReasonAdoption,
					Summary: "no change to .moltark/state.json",
				})
			}
		}
	}

	compacted := compactChanges(changes)
	return model.PlanningPhase{
		Changes: compacted,
		Summary: summarizeChanges(compacted),
	}, nil
}
