package moltark

import (
	"fmt"
	"path/filepath"
	"strings"
)

func classifyPath(format string, file string, path string, ownerComponentID string, desiredVersion string, desiredValue any, actualValue any, actualPresent bool, stateFile *ManagedFileState, state *State) (Change, error) {
	desiredFingerprint, err := fingerprintValue(desiredValue, true)
	if err != nil {
		return Change{}, fmt.Errorf("fingerprint desired value for %s %s: %w", file, path, err)
	}
	actualFingerprint, err := fingerprintValue(actualValue, actualPresent)
	if err != nil {
		return Change{}, fmt.Errorf("fingerprint actual value for %s %s: %w", file, path, err)
	}
	displayDesired := renderDisplayValue(format, desiredValue, true)
	displayActual := renderDisplayValue(format, actualValue, actualPresent)

	if state == nil {
		if !actualPresent {
			return Change{
				Status:  ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  ReasonBootstrap,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}

		if actualFingerprint == desiredFingerprint {
			return Change{
				Status:  ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  ReasonAdoption,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}

		return Change{
			Status:  ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  ReasonAdoption,
			Summary: conflictSummary(path, displayActual, displayDesired, true),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if stateFile == nil {
		reason := reasonForNewOwnedPath(state, ownerComponentID, desiredVersion)
		if !actualPresent {
			return Change{
				Status:  ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}

		if actualFingerprint == desiredFingerprint {
			return Change{
				Status:  ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}

		return Change{
			Status:  ChangeConflict,
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
			return Change{
				Status:  ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}, nil
		}
		if actualFingerprint == desiredFingerprint {
			return Change{
				Status:  ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  reason,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}, nil
		}
		return Change{
			Status:  ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  reason,
			Summary: conflictSummary(path, displayActual, displayDesired, false),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if actualFingerprint == desiredFingerprint {
		return Change{
			Status:  ChangeNoOp,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
			Summary: noOpSummary(path),
			After:   displayDesired,
		}, nil
	}

	if actualFingerprint == lastFingerprint {
		return Change{
			Status:  ChangeUpdate,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
			Summary: updateSummary(path, displayActual, displayDesired),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	if desiredFingerprint == lastFingerprint {
		return Change{
			Status:  ChangeDrift,
			File:    file,
			Path:    path,
			Reason:  ReasonDriftCorrection,
			Summary: driftSummary(path, displayActual, displayDesired),
			Before:  displayActual,
			After:   displayDesired,
		}, nil
	}

	return Change{
		Status:  ChangeConflict,
		File:    file,
		Path:    path,
		Reason:  reasonForDesiredChange(state, ownerComponentID, desiredVersion, lastFingerprint, desiredFingerprint),
		Summary: conflictSummary(path, displayActual, displayDesired, false),
		Before:  displayActual,
		After:   displayDesired,
	}, nil
}

func reasonForDesiredChange(state *State, ownerComponentID string, desiredVersion string, lastFingerprint string, desiredFingerprint string) ChangeReason {
	if state != nil && componentVersionChanged(state, ownerComponentID, desiredVersion) && lastFingerprint != desiredFingerprint {
		return ReasonTemplateUpgrade
	}
	return ReasonDesiredState
}

func reasonForNewOwnedPath(state *State, ownerComponentID string, desiredVersion string) ChangeReason {
	if state != nil && componentVersionChanged(state, ownerComponentID, desiredVersion) {
		return ReasonTemplateUpgrade
	}
	return ReasonDesiredState
}

func componentVersionChanged(state *State, ownerComponentID string, desiredVersion string) bool {
	if state == nil || ownerComponentID == "" || desiredVersion == "" {
		return false
	}
	lastVersion, ok := state.LastAppliedModel.componentVersion(ownerComponentID)
	return ok && lastVersion != "" && lastVersion != desiredVersion
}

func stateManagedFile(state *State, path string) *ManagedFileState {
	if state == nil {
		return nil
	}

	return state.managedFile(path)
}

func summarizeChanges(changes []Change) PlanSummary {
	var summary PlanSummary
	for _, change := range changes {
		switch change.Status {
		case ChangeCreate:
			summary.Create++
		case ChangeUpdate:
			summary.Update++
		case ChangeNoOp:
			summary.NoOp++
		case ChangeDrift:
			summary.Drift++
		case ChangeConflict:
			summary.Conflict++
		}
	}
	return summary
}

func hasConflict(changes []Change) bool {
	for _, change := range changes {
		if change.Status == ChangeConflict {
			return true
		}
	}
	return false
}

func compactChanges(changes []Change) []Change {
	compacted := make([]Change, 0, len(changes))
	for _, change := range changes {
		if change.Status == ChangeNoOp && change.Path == "" {
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
	case FileFormatJSON, FileFormatYAML:
		return renderJSONValue(value)
	default:
		rendered, err := renderTomlValue(value)
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

func RenderPlan(plan Plan) string {
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

	return ensureTrailingNewline(stringsJoin(lines, "\n"))
}

func RenderApply(result ApplyResult) string {
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

func RenderDoctor(report DoctorReport) string {
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
