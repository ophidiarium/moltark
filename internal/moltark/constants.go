package moltark

import "path/filepath"

const (
	MoltarkfileName         = "Moltarkfile"
	PyprojectFileName       = "pyproject.toml"
	GitattributesFileName   = ".gitattributes"
	StateDirName            = ".moltark"
	StateFileName           = "state.json"
	SchemaVersion           = 1
	TemplateVersion         = "python/v2"
	DefaultProjectVersion   = "0.1.0"
	DefaultRequiresPython   = ">=3.12"
	DefaultBuildBackend     = "hatchling.build"
	DefaultBuildRequirement = "hatchling"
	FileFormatTOML          = "toml"

	ModuleSourceCore   = "moltark/core"
	ModuleSourcePython = "moltark/python"
	ModuleSourceUV     = "astral/uv"

	ProjectKindPython = "python_project"

	CapabilityPythonPackageManager   = "moltark.python.package_manager"
	CapabilityPythonWorkspaceManager = "moltark.python.workspace_manager"

	IntentWorkspaceMembersRequest = "workspace_members_request"
	IntentPythonDependencyRequest = "python_dependency_request"

	SynthesisPhasePre  = "pre_synthesize"
	SynthesisPhaseMain = "synthesize"
	SynthesisPhasePost = "post_synthesize"

	TriggerCI        = "ci"
	TriggerPrePush   = "pre-push"
	TriggerPreCommit = "pre-commit"

	ProviderAttrManager        = "manager"
	ProviderAttrEcosystem      = "ecosystem"
	ProviderAttrFilePath       = "file_path"
	ProviderAttrOwnedPath      = "owned_path"
	ProviderAttrDependencyPath = "dependency_path"

	IntentAttrRequirement = "requirement"

	ProviderListArtifactKinds = "artifact_kinds"
	IntentListMemberPaths     = "member_paths"

	ArtifactKindPyPI = "pypi"
)

var baseOwnedPyprojectPaths = []string{
	"project.name",
	"project.version",
	"project.requires-python",
	"build-system.requires",
	"build-system.build-backend",
	"tool.moltark.schema-version",
	"tool.moltark.template-version",
}

const uvWorkspaceMembersPath = "tool.uv.workspace.members"

func statePath(root string) string {
	return filepath.Join(root, StateDirName, StateFileName)
}
