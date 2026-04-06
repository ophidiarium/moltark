package engine

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ophidiarium/moltark/internal/filefmt"
	"github.com/ophidiarium/moltark/internal/model"
)

type stateDocument struct {
	Exists bool
	Raw    string
	State  *model.State
}

func statePath(root string) string {
	return filepath.Join(root, model.StateDirName, model.StateFileName)
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

	var state model.State
	if err := json.Unmarshal(raw, &state); err != nil {
		return stateDocument{}, fmt.Errorf("parse %s: %w", path, err)
	}
	if state.SchemaVersion != model.SchemaVersion {
		return stateDocument{}, fmt.Errorf(
			"parse %s: unsupported schema_version %d (expected %d)",
			path,
			state.SchemaVersion,
			model.SchemaVersion,
		)
	}

	return stateDocument{
		Exists: true,
		Raw:    string(raw),
		State:  &state,
	}, nil
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

func buildState(desired model.DesiredModel, resolved model.ResolvedModel) (model.State, error) {
	managedFiles := make([]model.ManagedFileState, 0, len(resolved.ManagedFiles)+1)
	for _, file := range resolved.ManagedFiles {
		fingerprints := make(map[string]string, len(file.OwnedPaths))
		versions := make(map[string]string, len(file.OwnedPaths))
		for _, ownedPath := range file.OwnedPaths {
			value, err := filefmt.RequireStructuredValue(file.DesiredValues, file.Format, ownedPath)
			if err != nil {
				return model.State{}, fmt.Errorf("build state for %s %s: %w", file.Path, ownedPath, err)
			}
			fingerprint, err := fingerprintValue(value, true)
			if err != nil {
				return model.State{}, fmt.Errorf("build state for %s %s: %w", file.Path, ownedPath, err)
			}
			fingerprints[ownedPath] = fingerprint
			if version := file.OwnedPathVersions[ownedPath]; version != "" {
				versions[ownedPath] = version
			}
		}

		managedFiles = append(managedFiles, model.ManagedFileState{
			Path:              file.Path,
			OwnedPaths:        append([]string(nil), file.OwnedPaths...),
			OwnedPathVersions: versions,
			Fingerprints:      fingerprints,
		})
	}

	gitattributesBlock := filefmt.ManagedGitattributesBlock()
	gitattributesFingerprint, err := fingerprintValue(gitattributesBlock, true)
	if err != nil {
		return model.State{}, fmt.Errorf("build state for %s moltark.block: %w", model.GitattributesFileName, err)
	}
	managedFiles = append(managedFiles, model.ManagedFileState{
		Path:       model.GitattributesFileName,
		OwnedPaths: []string{"moltark.block"},
		Fingerprints: map[string]string{
			"moltark.block": gitattributesFingerprint,
		},
	})

	return model.State{
		SchemaVersion:    model.SchemaVersion,
		ManagedFiles:     managedFiles,
		LastAppliedModel: summarizeModel(desired),
	}, nil
}

func summarizeModel(desired model.DesiredModel) model.ModelSummary {
	summary := model.ModelSummary{
		Projects:   make([]model.ProjectSummary, 0, len(desired.Projects)),
		Components: make([]model.ComponentSummary, 0, len(desired.Components)),
	}
	for _, project := range desired.Projects {
		summary.Projects = append(summary.Projects, model.ProjectSummary{
			ID:            project.ID,
			Kind:          project.Kind,
			Name:          project.Name,
			Path:          project.Path,
			EffectivePath: project.EffectivePath,
			Attributes:    model.CloneStringMap(project.Attributes),
			ParentID:      project.ParentID,
		})
	}
	for _, component := range desired.Components {
		summary.Components = append(summary.Components, model.ComponentSummary{
			ID:              component.ID,
			Kind:            component.Kind,
			Module:          component.Module,
			Version:         component.Version,
			TargetProjectID: component.TargetProjectID,
		})
	}
	return summary
}

func renderState(state model.State) (string, error) {
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
	dir := filepath.Join(root, model.StateDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, model.StateFileName), []byte(body), 0o644)
}
