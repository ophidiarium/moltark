package moltark

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type stateDocument struct {
	Exists bool
	Raw    string
	State  *State
}

func loadState(root string) (stateDocument, error) {
	path := statePath(root)
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return stateDocument{}, nil
		}
		return stateDocument{}, fmt.Errorf("read %s: %w", path, err)
	}

	var state State
	if err := json.Unmarshal(raw, &state); err != nil {
		return stateDocument{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if state.SchemaVersion != SchemaVersion {
		return stateDocument{}, fmt.Errorf(
			"parse %s: unsupported schema_version %d (expected %d)",
			path,
			state.SchemaVersion,
			SchemaVersion,
		)
	}

	return stateDocument{
		Exists: true,
		Raw:    string(raw),
		State:  &state,
	}, nil
}

func (s *State) managedFile(path string) *ManagedFileState {
	if s == nil {
		return nil
	}

	for i := range s.ManagedFiles {
		if s.ManagedFiles[i].Path == path {
			return &s.ManagedFiles[i]
		}
	}

	return nil
}

func fingerprintValue(value any, present bool) (string, error) {
	payload := map[string]any{
		"present": present,
		"value":   value,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("fingerprint marshal: %w", err)
	}

	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:]), nil
}

func buildState(model DesiredModel, resolved ResolvedModel) (State, error) {
	managedFiles := make([]ManagedFileState, 0, len(resolved.ManagedFiles)+1)
	for _, file := range resolved.ManagedFiles {
		fingerprints := make(map[string]string, len(file.OwnedPaths))
		versions := make(map[string]string, len(file.OwnedPaths))
		for _, ownedPath := range file.OwnedPaths {
			value, err := requireStructuredValue(file.DesiredValues, file.Format, ownedPath)
			if err != nil {
				return State{}, fmt.Errorf("build state for %s %s: %w", file.Path, ownedPath, err)
			}
			fingerprint, err := fingerprintValue(value, true)
			if err != nil {
				return State{}, fmt.Errorf("build state for %s %s: %w", file.Path, ownedPath, err)
			}
			fingerprints[ownedPath] = fingerprint
			if version := file.OwnedPathVersions[ownedPath]; version != "" {
				versions[ownedPath] = version
			}
		}

		managedFiles = append(managedFiles, ManagedFileState{
			Path:              file.Path,
			OwnedPaths:        append([]string(nil), file.OwnedPaths...),
			OwnedPathVersions: versions,
			Fingerprints:      fingerprints,
		})
	}

	gitattributesBlock := managedGitattributesBlock()
	gitattributesFingerprint, err := fingerprintValue(gitattributesBlock, true)
	if err != nil {
		return State{}, fmt.Errorf("build state for %s moltark.block: %w", GitattributesFileName, err)
	}
	managedFiles = append(managedFiles, ManagedFileState{
		Path:       GitattributesFileName,
		OwnedPaths: []string{"moltark.block"},
		Fingerprints: map[string]string{
			"moltark.block": gitattributesFingerprint,
		},
	})

	return State{
		SchemaVersion:    SchemaVersion,
		ManagedFiles:     managedFiles,
		LastAppliedModel: summarizeModel(model),
	}, nil
}

func summarizeModel(model DesiredModel) ModelSummary {
	summary := ModelSummary{
		Projects:   make([]ProjectSummary, 0, len(model.Projects)),
		Components: make([]ComponentSummary, 0, len(model.Components)),
	}
	for _, project := range model.Projects {
		summary.Projects = append(summary.Projects, ProjectSummary{
			ID:            project.ID,
			Kind:          project.Kind,
			Name:          project.Name,
			Path:          project.Path,
			EffectivePath: project.EffectivePath,
			Attributes:    cloneStringMap(project.Attributes),
			ParentID:      project.ParentID,
		})
	}
	for _, component := range model.Components {
		summary.Components = append(summary.Components, ComponentSummary{
			ID:              component.ID,
			Kind:            component.Kind,
			Module:          component.Module,
			Version:         component.Version,
			TargetProjectID: component.TargetProjectID,
		})
	}
	return summary
}

func renderState(state State) (string, error) {
	var buffer bytes.Buffer
	encoder := json.NewEncoder(&buffer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(state); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func writeStateFile(root string, body string) error {
	dir := filepath.Join(root, StateDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, StateFileName), []byte(body), 0o644)
}
