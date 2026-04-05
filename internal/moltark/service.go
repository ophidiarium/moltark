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

func (s Service) Plan(root string) (Plan, error) {
	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return Plan{}, err
	}
	return pipeline.PlanResult(), nil
}

func (s Service) Apply(root string, plan Plan) (ApplyResult, error) {
	if plan.HasConflicts() {
		return ApplyResult{}, fmt.Errorf("cannot apply a plan with conflicts")
	}

	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return ApplyResult{}, err
	}

	freshPlan := pipeline.PlanResult()
	if !sameApplyIntent(plan, freshPlan) {
		return ApplyResult{}, fmt.Errorf("repository changed since planning; rerun `moltark plan` and review the updated changes")
	}
	plan = freshPlan

	wrote := []string{}
	for _, managedFile := range pipeline.Resolve.Resolved.ManagedFiles {
		doc := pipeline.fileDocs[managedFile.Path]
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

	gitattributesBody := mutateGitattributes(pipeline.gitattributesRaw)
	if gitattributesBody != ensureTrailingNewline(pipeline.gitattributesRaw) {
		if err := writeRepoFile(root, GitattributesFileName, gitattributesBody); err != nil {
			return ApplyResult{}, err
		}
		wrote = append(wrote, GitattributesFileName)
	}

	if pipeline.nextStateRaw != pipeline.stateRaw {
		if err := writeStateFile(root, pipeline.nextStateRaw); err != nil {
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
		reflect.DeepEqual(left.NextState, right.NextState) &&
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

func (s Service) Show(root string) (Pipeline, error) {
	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return Pipeline{}, err
	}
	return pipeline, nil
}

func (s Service) Doctor(root string) (DoctorReport, error) {
	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return DoctorReport{}, err
	}
	plan := pipeline.PlanResult()

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
		values := map[string]any{}
		if err := decodeToml([]byte(raw), &values); err != nil {
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
