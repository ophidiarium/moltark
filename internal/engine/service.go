package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
)

type Service struct{}

func NewService() Service {
	return Service{}
}

func (s Service) Plan(root string) (model.Plan, error) {
	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return model.Plan{}, err
	}
	return pipeline.PlanResult(), nil
}

func (s Service) Apply(root string, plan model.Plan) (model.ApplyResult, error) {
	if plan.HasConflicts() {
		return model.ApplyResult{}, fmt.Errorf("cannot apply a plan with conflicts")
	}

	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return model.ApplyResult{}, err
	}

	freshPlan := pipeline.PlanResult()
	if !sameApplyIntent(plan, freshPlan) {
		return model.ApplyResult{}, fmt.Errorf("repository changed since planning; rerun `moltark plan` and review the updated changes")
	}
	plan = freshPlan

	wrote := []string{}
	for _, managedFile := range pipeline.Resolve.Resolved.ManagedFiles {
		doc := pipeline.fileDocs[managedFile.Path]
		body, err := mutateStructuredFile(doc.Raw, managedFile)
		if err != nil {
			return model.ApplyResult{}, err
		}
		if body == filefmt.EnsureTrailingNewline(doc.Raw) {
			continue
		}
		if err := writeRepoFile(root, managedFile.Path, body); err != nil {
			return model.ApplyResult{}, err
		}
		wrote = append(wrote, managedFile.Path)
	}

	gitattributesBody := filefmt.MutateGitattributes(pipeline.gitattributesRaw)
	if gitattributesBody != filefmt.EnsureTrailingNewline(pipeline.gitattributesRaw) {
		if err := writeRepoFile(root, model.GitattributesFileName, gitattributesBody); err != nil {
			return model.ApplyResult{}, err
		}
		wrote = append(wrote, model.GitattributesFileName)
	}

	if pipeline.nextStateRaw != pipeline.stateRaw {
		if err := writeStateFile(root, pipeline.nextStateRaw); err != nil {
			return model.ApplyResult{}, err
		}
		wrote = append(wrote, filepath.ToSlash(filepath.Join(model.StateDirName, model.StateFileName)))
	}

	return model.ApplyResult{
		Plan:  plan,
		Wrote: wrote,
	}, nil
}

func sameApplyIntent(left model.Plan, right model.Plan) bool {
	return reflect.DeepEqual(left.Desired, right.Desired) &&
		reflect.DeepEqual(left.Resolved, right.Resolved) &&
		reflect.DeepEqual(left.NextState, right.NextState) &&
		reflect.DeepEqual(actionableChanges(left.Changes), actionableChanges(right.Changes)) &&
		reflect.DeepEqual(actionableSummary(left.Summary), actionableSummary(right.Summary))
}

func actionableChanges(changes []model.Change) []model.Change {
	filtered := make([]model.Change, 0, len(changes))
	for _, change := range changes {
		if change.Status == model.ChangeNoOp {
			continue
		}
		filtered = append(filtered, change)
	}
	return filtered
}

func actionableSummary(summary model.PlanSummary) model.PlanSummary {
	summary.NoOp = 0
	return summary
}

func reasonForStateUpdate(previous *model.State, next *model.State) model.ChangeReason {
	if previous == nil || next == nil {
		return model.ReasonBootstrap
	}
	if modelHasTemplateUpgrade(previous.LastAppliedModel, next.LastAppliedModel) {
		return model.ReasonTemplateUpgrade
	}
	return model.ReasonDesiredState
}

func modelHasTemplateUpgrade(previous model.ModelSummary, next model.ModelSummary) bool {
	for _, component := range next.Components {
		if component.Version == "" {
			continue
		}
		lastVersion, ok := previous.ComponentVersion(component.ID)
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

func (s Service) Doctor(root string) (model.DoctorReport, error) {
	pipeline, err := s.BuildPipeline(root)
	if err != nil {
		return model.DoctorReport{}, err
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

	return model.DoctorReport{
		Plan:      plan,
		HasIssues: hasIssues,
		Messages:  messages,
	}, nil
}

func loadStructuredDocument(path string, format string) (model.FileDocument, error) {
	raw, exists, err := readOptionalFile(path)
	if err != nil {
		return model.FileDocument{}, err
	}

	doc := model.FileDocument{
		Raw:    raw,
		Exists: exists,
		Values: map[string]any{},
	}
	if !exists {
		return doc, nil
	}

	switch format {
	case model.FileFormatTOML:
		values := map[string]any{}
		if err := filefmt.DecodeToml([]byte(raw), &values); err != nil {
			return model.FileDocument{}, err
		}
		doc.Values = values
	case model.FileFormatJSON:
		values, err := filefmt.ParseJSONValues([]byte(raw))
		if err != nil {
			return model.FileDocument{}, err
		}
		doc.Values = values
	case model.FileFormatYAML:
		values, err := filefmt.ParseYAMLValues([]byte(raw))
		if err != nil {
			return model.FileDocument{}, err
		}
		doc.Values = values
	default:
		return model.FileDocument{}, fmt.Errorf("unsupported file format %q", format)
	}
	return doc, nil
}

func mutateStructuredFile(raw string, file model.ManagedFileSpec) (string, error) {
	switch file.Format {
	case model.FileFormatTOML:
		return filefmt.MutateTOMLFile(raw, file.DesiredValues, file.OwnedPaths)
	case model.FileFormatJSON:
		return filefmt.MutateJSONFile(raw, file.DesiredValues, file.OwnedPaths)
	case model.FileFormatYAML:
		return filefmt.MutateYAMLFile(raw, file.DesiredValues, file.OwnedPaths)
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
