package engine

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
)

func classifyPath(format string, file string, path string, ownerComponentID string, desiredVersion string, desiredValue any, actualValue any, actualPresent bool, stateFile *model.ManagedFileState, state *model.State) (model.Change, error) {
	desiredFingerprint, err := fingerprintValue(desiredValue, true)
	if err != nil {
		return model.Change{}, fmt.Errorf("fingerprint desired value for %s %s: %w", file, path, err)
	}
	actualFingerprint, err := fingerprintValue(actualValue, actualPresent)
	if err != nil {
		return model.Change{}, fmt.Errorf("fingerprint actual value for %s %s: %w", file, path, err)
	}
	displayDesired := renderDisplayValue(format, desiredValue, true)
	displayActual := renderDisplayValue(format, actualValue, actualPresent)

	if state == nil {
		if !actualPresent {
			return model.Change{
				Status:  model.ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  model.ReasonBootstrap,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}

		if actualFingerprint == desiredFingerprint {
			return model.Change{
				Status:  model.ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  model.ReasonAdoption,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}

		return model.Change{
			Status:  model.ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  model.ReasonAdoption,
			Summary: conflictSummary(path, displayActual, displayDesired, true),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if stateFile == nil {
		reason := reasonForNewOwnedPath(state, ownerComponentID, desiredVersion)
		if !actualPresent {
			return model.Change{
				Status:  model.ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}

		if actualFingerprint == desiredFingerprint {
			return model.Change{
				Status:  model.ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}

		return model.Change{
			Status:  model.ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  reason,
			Summary: conflictSummary(path, displayActual, displayDesired, false),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	lastFingerprint, tracked := stateFile.Fingerprints[path]
	if !tracked {
		reason := reasonForNewOwnedPath(state, ownerComponentID, desiredVersion)
		if !actualPresent {
			return model.Change{
				Status:  model.ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}
		if actualFingerprint == desiredFingerprint {
			return model.Change{
				Status:  model.ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}
		return model.Change{
			Status:  model.ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  reason,
			Summary: conflictSummary(path, displayActual, displayDesired, false),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if actualFingerprint == desiredFingerprint {
		return model.Change{
			Status:  model.ChangeNoOp,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
			Summary: noOpSummary(path),
			After:   displayDesired,
		}, nil
	}

	if actualFingerprint == lastFingerprint {
		return model.Change{
			Status:  model.ChangeUpdate,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
			Summary: updateSummary(path, displayActual, displayDesired),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if desiredFingerprint == lastFingerprint {
		return model.Change{
			Status:  model.ChangeDrift,
			File:    file,
			Path:    path,
			Reason:  model.ReasonDriftCorrection,
			Summary: driftSummary(path, displayActual, displayDesired),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	return model.Change{
		Status:  model.ChangeConflict,
		File:    file,
		Path:    path,
		Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
		Summary: conflictSummary(path, displayActual, displayDesired, false),
		Before:  displayActual,
		After:   displayDesired,
	}, nil
}

func reasonForDesiredChange(state *model.State, ownerComponentID string, desiredVersion string, lastFingerprint string, desiredFingerprint string) model.ChangeReason {
	if state != nil && componentVersionChanged(state, ownerComponentID, desiredVersion) && lastFingerprint != desiredFingerprint {
		return model.ReasonTemplateUpgrade
	}
	return model.ReasonDesiredState
}

func reasonForNewOwnedPath(state *model.State, ownerComponentID string, desiredVersion string) model.ChangeReason {
	if state != nil && componentVersionChanged(state, ownerComponentID, desiredVersion) {
		return model.ReasonTemplateUpgrade
	}
	return model.ReasonDesiredState
}

func componentVersionChanged(state *model.State, ownerComponentID string, desiredVersion string) bool {
	if state == nil || ownerComponentID == "" || desiredVersion == "" {
		return false
	}
	lastVersion, ok := state.LastAppliedModel.ComponentVersion(ownerComponentID)
	return ok && lastVersion != "" && lastVersion != desiredVersion
}

func stateManagedFile(state *model.State, path string) *model.ManagedFileState {
	if state == nil {
		return nil
	}

	return state.ManagedFile(path)
}

func summarizeChanges(changes []model.Change) model.PlanSummary {
	var summary model.PlanSummary
	for _, change := range changes {
		switch change.Status {
		case model.ChangeCreate:
			summary.Create++
		case model.ChangeUpdate:
			summary.Update++
		case model.ChangeNoOp:
			summary.NoOp++
		case model.ChangeDrift:
			summary.Drift++
		case model.ChangeConflict:
			summary.Conflict++
		}
	}
	return summary
}

func hasConflict(changes []model.Change) bool {
	for _, change := range changes {
		if change.Status == model.ChangeConflict {
			return true
		}
	}
	return false
}

func compactChanges(changes []model.Change) []model.Change {
	compacted := make([]model.Change, 0, len(changes))
	for _, change := range changes {
		if change.Status == model.ChangeNoOp && change.Path == "" {
			continue
		}
		compacted = append(compacted, change)
	}
	return compacted
}

func renderDisplayValue(format string, value any, present bool) string {
	if !present {
		return "<absent>"
	}

	switch format {
	case model.FileFormatJSON, model.FileFormatYAML:
		return filefmt.RenderJSONValue(value)
	default:
		rendered, err := filefmt.RenderTomlValue(value)
		if err != nil {
			return fmt.Sprintf("<render error: %v>", err)
		}
		return rendered
	}
}

func createSummary(path string, after string) string {
	if path == "moltark.block" {
		return "ensure Moltark managed block in .gitattributes"
	}
	return fmt.Sprintf("set %s = %s", path, after)
}

func noOpSummary(path string) string {
	if path == "moltark.block" {
		return "no change to Moltark managed block"
	}
	if path == "project.dependencies" {
		return "no change to dependencies (user-managed)"
	}
	return fmt.Sprintf("no change to %s", path)
}

func updateSummary(path string, before string, after string) string {
	if path == "moltark.block" {
		return "update Moltark managed block in .gitattributes"
	}
	return fmt.Sprintf("update %s: %s -> %s", path, before, after)
}

func driftSummary(path string, before string, after string) string {
	if path == "moltark.block" {
		return "drift detected in Moltark managed block"
	}
	return fmt.Sprintf("drift detected in %s: %s -> %s", path, before, after)
}

func conflictSummary(path string, before string, after string, adoption bool) string {
	if path == "moltark.block" {
		if adoption {
			return "conflict in Moltark managed block during initial adoption"
		}
		return "conflict in Moltark managed block"
	}
	if adoption {
		return fmt.Sprintf("conflict in %s during initial adoption: actual %s, desired %s", path, before, after)
	}
	return fmt.Sprintf("conflict in %s: actual %s, desired %s", path, before, after)
}

func RenderPlan(plan model.Plan) string {
	lines := []string{
		fmt.Sprintf(
			"Plan: %d to create, %d to update, %d drift, %d conflict, %d no-op",
			plan.Summary.Create,
			plan.Summary.Update,
			plan.Summary.Drift,
			plan.Summary.Conflict,
			plan.Summary.NoOp,
		),
		"",
	}

	for _, change := range plan.Changes {
		lines = append(lines, change.Summary)
	}

	return filefmt.EnsureTrailingNewline(stringsJoin(lines, "\n"))
}

func RenderApply(result model.ApplyResult) string {
	lines := []string{"Apply complete."}
	if len(result.Wrote) == 0 {
		lines = append(lines, "No files changed.")
		return stringsJoin(lines, "\n")
	}

	for _, path := range result.Wrote {
		lines = append(lines, fmt.Sprintf("wrote %s", filepath.ToSlash(path)))
	}

	return stringsJoin(lines, "\n")
}

func RenderDoctor(report model.DoctorReport) string {
	lines := []string{
		"Doctor report:",
	}
	lines = append(lines, report.Messages...)
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf(
		"%d create, %d update, %d drift, %d conflict, %d no-op",
		report.Plan.Summary.Create,
		report.Plan.Summary.Update,
		report.Plan.Summary.Drift,
		report.Plan.Summary.Conflict,
		report.Plan.Summary.NoOp,
	))
	return stringsJoin(lines, "\n")
}

func stringsJoin(lines []string, sep string) string {
	return strings.Join(lines, sep) + "\n"
}
