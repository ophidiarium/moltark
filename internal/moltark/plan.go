package moltark

import (
	"fmt"
	"path/filepath"
)

func classifyPath(file string, path string, desiredValue any, actualValue any, actualPresent bool, stateFile *ManagedFileState, state *State) Change {
	desiredFingerprint := fingerprintValue(desiredValue, true)
	actualFingerprint := fingerprintValue(actualValue, actualPresent)
	displayDesired := renderDisplayValue(desiredValue, true)
	displayActual := renderDisplayValue(actualValue, actualPresent)

	if state == nil || stateFile == nil {
		if !actualPresent {
			return Change{
				Status:  ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  ReasonBootstrap,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}
		}

		if actualFingerprint == desiredFingerprint {
			return Change{
				Status:  ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  ReasonAdoption,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}
		}

		return Change{
			Status:  ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  ReasonAdoption,
			Summary: conflictSummary(path, displayActual, displayDesired, true),
			Before:  displayActual,
			After:   displayDesired,
		}
	}

	lastFingerprint, tracked := stateFile.Fingerprints[path]
	if !tracked {
		if !actualPresent {
			return Change{
				Status:  ChangeCreate,
				File:    file,
				Path:    path,
				Reason:  ReasonTemplateUpgrade,
				Summary: createSummary(path, displayDesired),
				After:   displayDesired,
			}
		}
		if actualFingerprint == desiredFingerprint {
			return Change{
				Status:  ChangeNoOp,
				File:    file,
				Path:    path,
				Reason:  ReasonTemplateUpgrade,
				Summary: noOpSummary(path),
				After:   displayDesired,
			}
		}
		return Change{
			Status:  ChangeConflict,
			File:    file,
			Path:    path,
			Reason:  ReasonTemplateUpgrade,
			Summary: conflictSummary(path, displayActual, displayDesired, false),
			Before:  displayActual,
			After:   displayDesired,
		}
	}

	if actualFingerprint == desiredFingerprint {
		return Change{
			Status:  ChangeNoOp,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, lastFingerprint, desiredFingerprint),
			Summary: noOpSummary(path),
			After:   displayDesired,
		}
	}

	if actualFingerprint == lastFingerprint {
		return Change{
			Status:  ChangeUpdate,
			File:    file,
			Path:    path,
			Reason:  reasonForDesiredChange(state, lastFingerprint, desiredFingerprint),
			Summary: updateSummary(path, displayActual, displayDesired),
			Before:  displayActual,
			After:   displayDesired,
		}
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
		}
	}

	return Change{
		Status:  ChangeConflict,
		File:    file,
		Path:    path,
		Reason:  reasonForDesiredChange(state, lastFingerprint, desiredFingerprint),
		Summary: conflictSummary(path, displayActual, displayDesired, false),
		Before:  displayActual,
		After:   displayDesired,
	}
}

func reasonForDesiredChange(state *State, lastFingerprint string, desiredFingerprint string) ChangeReason {
	if state != nil && state.TemplateVersion != TemplateVersion && lastFingerprint != desiredFingerprint {
		return ReasonTemplateUpgrade
	}
	return ReasonDesiredState
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

func renderDisplayValue(value any, present bool) string {
	if !present {
		return "<absent>"
	}

	return renderTomlValue(value)
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
	return fmt.Sprintf("%s\n", joinStringSlice(lines, sep))
}

func joinStringSlice(lines []string, sep string) string {
	if len(lines) == 0 {
		return ""
	}
	out := lines[0]
	for _, line := range lines[1:] {
		out += sep + line
	}
	return out
}
