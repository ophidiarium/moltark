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

func fingerprintValue(value any, present bool) string {
	payload := map[string]any{
		"present": present,
		"value":   value,
	}

	encoded, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func buildState(model DesiredModel, resolved ResolvedModel) State {
	managedFiles := make([]ManagedFileState, 0, len(resolved.ManagedFiles)+1)
	for _, file := range resolved.ManagedFiles {
		fingerprints := make(map[string]string, len(file.OwnedPaths))
		for _, ownedPath := range file.OwnedPaths {
			value, _ := lookupPath(file.DesiredValues, ownedPath)
			fingerprints[ownedPath] = fingerprintValue(value, true)
		}

		managedFiles = append(managedFiles, ManagedFileState{
			Path:         file.Path,
			OwnedPaths:   append([]string(nil), file.OwnedPaths...),
			Fingerprints: fingerprints,
		})
	}

	gitattributesBlock := managedGitattributesBlock()
	managedFiles = append(managedFiles, ManagedFileState{
		Path:       GitattributesFileName,
		OwnedPaths: []string{"moltark.block"},
		Fingerprints: map[string]string{
			"moltark.block": fingerprintValue(gitattributesBlock, true),
		},
	})

	return State{
		SchemaVersion:    SchemaVersion,
		TemplateVersion:  TemplateVersion,
		ManagedFiles:     managedFiles,
		LastAppliedModel: summarizeModel(model),
	}
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
			ParentID:      project.ParentID,
		})
	}
	for _, component := range model.Components {
		summary.Components = append(summary.Components, ComponentSummary{
			ID:              component.ID,
			Kind:            component.Kind,
			Module:          component.Module,
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
