package moltark

type DesiredModel struct {
	Projects   []ProjectSpec   `json:"projects"`
	Components []ComponentSpec `json:"components"`
}

type ProjectSpec struct {
	ID            string             `json:"id"`
	Kind          string             `json:"kind"`
	Name          string             `json:"name"`
	Path          string             `json:"path"`
	EffectivePath string             `json:"effective_path"`
	Attributes    map[string]string  `json:"attributes,omitempty"`
	ParentID      string             `json:"parent_id,omitempty"`
	Python        *PythonProjectSpec `json:"python,omitempty"`
}

type PythonProjectSpec struct {
	Version         string      `json:"version"`
	RequiresPython  string      `json:"requires_python"`
	TemplateVersion string      `json:"template_version"`
	BuildSystem     BuildSystem `json:"build_system"`
}

type ComponentSpec struct {
	ID                    string                     `json:"id"`
	Kind                  string                     `json:"kind"`
	Module                string                     `json:"module"`
	Version               string                     `json:"version,omitempty"`
	TargetProjectID       string                     `json:"target_project_id,omitempty"`
	Facts                 []FactProviderSpec         `json:"facts,omitempty"`
	Providers             []CapabilityProvider       `json:"providers,omitempty"`
	Files                 []StructuredFileSpec       `json:"files,omitempty"`
	SynthesisHooks        []SynthesisHookSpec        `json:"synthesis_hooks,omitempty"`
	BootstrapRequirements []BootstrapRequirementSpec `json:"bootstrap_requirements,omitempty"`
	Tasks                 []TaskSpec                 `json:"tasks,omitempty"`
	TaskSurfaces          []TaskSurfaceSpec          `json:"task_surfaces,omitempty"`
	TriggerBindings       []TriggerBindingSpec       `json:"trigger_bindings,omitempty"`
	RoutedIntents         []RoutedIntentSpec         `json:"routed_intents,omitempty"`
}

type CapabilityProvider struct {
	Capability     string              `json:"capability"`
	ScopeProjectID string              `json:"scope_project_id,omitempty"`
	Attributes     map[string]string   `json:"attributes,omitempty"`
	Lists          map[string][]string `json:"lists,omitempty"`
}

type FactProviderSpec struct {
	Name           string         `json:"name"`
	ScopeProjectID string         `json:"scope_project_id,omitempty"`
	Values         map[string]any `json:"values"`
}

type FactValueRef struct {
	TargetProjectID string `json:"target_project_id"`
	Name            string `json:"name"`
	Path            string `json:"path"`
}

type StructuredFileSpec struct {
	Path             string         `json:"path"`
	Format           string         `json:"format"`
	OwnedPaths       []string       `json:"owned_paths"`
	UserManagedPaths []string       `json:"user_managed_paths,omitempty"`
	DesiredValues    map[string]any `json:"desired_values"`
}

type RoutedIntentSpec struct {
	Kind            string              `json:"kind"`
	Capability      string              `json:"capability"`
	TargetProjectID string              `json:"target_project_id"`
	Attributes      map[string]string   `json:"attributes,omitempty"`
	Lists           map[string][]string `json:"lists,omitempty"`
}

type SynthesisHookSpec struct {
	Phase           string `json:"phase"`
	TargetProjectID string `json:"target_project_id"`
	Description     string `json:"description,omitempty"`
}

type BootstrapRequirementSpec struct {
	Tool            string   `json:"tool"`
	TargetProjectID string   `json:"target_project_id"`
	Purpose         string   `json:"purpose,omitempty"`
	Strategies      []string `json:"strategies,omitempty"`
}

type TaskSpec struct {
	Name            string   `json:"name"`
	TargetProjectID string   `json:"target_project_id"`
	Command         []string `json:"command"`
	Runtime         string   `json:"runtime,omitempty"`
	Tags            []string `json:"tags,omitempty"`
}

type TaskSurfaceSpec struct {
	Name            string `json:"name,omitempty"`
	Kind            string `json:"kind"`
	TargetProjectID string `json:"target_project_id"`
}

type TriggerBindingSpec struct {
	Trigger         string   `json:"trigger"`
	TargetProjectID string   `json:"target_project_id"`
	MatchNames      []string `json:"match_names,omitempty"`
	MatchTags       []string `json:"match_tags,omitempty"`
}

type ResolvedModel struct {
	ManagedFiles            []ManagedFileSpec              `json:"managed_files"`
	Facts                   []FactBinding                  `json:"facts,omitempty"`
	Providers               []ProviderBinding              `json:"providers"`
	SynthesisHooks          []ResolvedSynthesisHook        `json:"synthesis_hooks,omitempty"`
	BootstrapRequirements   []ResolvedBootstrapRequirement `json:"bootstrap_requirements,omitempty"`
	Tasks                   []ResolvedTask                 `json:"tasks,omitempty"`
	TaskSurfaces            []ResolvedTaskSurface          `json:"task_surfaces,omitempty"`
	ResolvedTriggerBindings []ResolvedTriggerBinding       `json:"resolved_trigger_bindings,omitempty"`
	ResolvedIntents         []IntentBinding                `json:"resolved_intents"`
}

type ManagedFileSpec struct {
	Path              string            `json:"path"`
	Format            string            `json:"format"`
	OwnedPaths        []string          `json:"owned_paths"`
	OwnedPathOwners   map[string]string `json:"owned_path_owners,omitempty"`
	OwnedPathVersions map[string]string `json:"owned_path_versions,omitempty"`
	UserManagedPaths  []string          `json:"user_managed_paths,omitempty"`
	DesiredValues     map[string]any    `json:"desired_values"`
	SourceComponents  []string          `json:"source_components"`
}

type ProviderBinding struct {
	ComponentID    string              `json:"component_id"`
	Capability     string              `json:"capability"`
	ScopeProjectID string              `json:"scope_project_id,omitempty"`
	Attributes     map[string]string   `json:"attributes,omitempty"`
	Lists          map[string][]string `json:"lists,omitempty"`
}

type FactBinding struct {
	ComponentID    string         `json:"component_id"`
	Name           string         `json:"name"`
	ScopeProjectID string         `json:"scope_project_id,omitempty"`
	Values         map[string]any `json:"values"`
}

type IntentBinding struct {
	ComponentID            string              `json:"component_id"`
	IntentKind             string              `json:"intent_kind"`
	Capability             string              `json:"capability"`
	TargetProjectID        string              `json:"target_project_id"`
	Attributes             map[string]string   `json:"attributes,omitempty"`
	Lists                  map[string][]string `json:"lists,omitempty"`
	ProviderComponentID    string              `json:"provider_component_id"`
	ProviderScopeProjectID string              `json:"provider_scope_project_id,omitempty"`
}

type ResolvedSynthesisHook struct {
	ComponentID       string `json:"component_id"`
	Phase             string `json:"phase"`
	TargetProjectID   string `json:"target_project_id"`
	TargetProjectPath string `json:"target_project_path,omitempty"`
	Description       string `json:"description,omitempty"`
}

type ResolvedBootstrapRequirement struct {
	ComponentID       string   `json:"component_id"`
	Tool              string   `json:"tool"`
	TargetProjectID   string   `json:"target_project_id"`
	TargetProjectPath string   `json:"target_project_path,omitempty"`
	Purpose           string   `json:"purpose,omitempty"`
	Strategies        []string `json:"strategies,omitempty"`
}

type ResolvedTask struct {
	ComponentID       string   `json:"component_id"`
	Name              string   `json:"name"`
	TargetProjectID   string   `json:"target_project_id"`
	TargetProjectPath string   `json:"target_project_path,omitempty"`
	Command           []string `json:"command"`
	Runtime           string   `json:"runtime,omitempty"`
	Tags              []string `json:"tags,omitempty"`
}

type ResolvedTaskSurface struct {
	ComponentID       string `json:"component_id"`
	Name              string `json:"name,omitempty"`
	Kind              string `json:"kind"`
	TargetProjectID   string `json:"target_project_id"`
	TargetProjectPath string `json:"target_project_path,omitempty"`
}

type ResolvedTriggerBinding struct {
	ComponentID           string   `json:"component_id"`
	Trigger               string   `json:"trigger"`
	TargetProjectID       string   `json:"target_project_id"`
	TargetProjectPath     string   `json:"target_project_path,omitempty"`
	MatchNames            []string `json:"match_names,omitempty"`
	MatchTags             []string `json:"match_tags,omitempty"`
	TaskComponentID       string   `json:"task_component_id"`
	TaskName              string   `json:"task_name"`
	TaskTargetProjectID   string   `json:"task_target_project_id"`
	TaskTargetProjectPath string   `json:"task_target_project_path,omitempty"`
}

type BuildSystem struct {
	Requires []string `json:"requires"`
	Backend  string   `json:"backend"`
}

type State struct {
	SchemaVersion    int                `json:"schema_version"`
	ManagedFiles     []ManagedFileState `json:"managed_files"`
	LastAppliedModel ModelSummary       `json:"last_applied_model"`
}

type ManagedFileState struct {
	Path              string            `json:"path"`
	OwnedPaths        []string          `json:"owned_paths"`
	OwnedPathVersions map[string]string `json:"owned_path_versions,omitempty"`
	Fingerprints      map[string]string `json:"fingerprints"`
}

type ModelSummary struct {
	Projects   []ProjectSummary   `json:"projects"`
	Components []ComponentSummary `json:"components"`
}

type ProjectSummary struct {
	ID            string            `json:"id"`
	Kind          string            `json:"kind"`
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	EffectivePath string            `json:"effective_path"`
	Attributes    map[string]string `json:"attributes,omitempty"`
	ParentID      string            `json:"parent_id,omitempty"`
}

type ComponentSummary struct {
	ID              string `json:"id"`
	Kind            string `json:"kind"`
	Module          string `json:"module"`
	Version         string `json:"version,omitempty"`
	TargetProjectID string `json:"target_project_id,omitempty"`
}

type ChangeStatus string

const (
	ChangeCreate   ChangeStatus = "create"
	ChangeUpdate   ChangeStatus = "update"
	ChangeNoOp     ChangeStatus = "no-op"
	ChangeDrift    ChangeStatus = "drift"
	ChangeConflict ChangeStatus = "conflict"
)

type ChangeReason string

const (
	ReasonBootstrap       ChangeReason = "bootstrap"
	ReasonDesiredState    ChangeReason = "desired_state_change"
	ReasonTemplateUpgrade ChangeReason = "template_upgrade"
	ReasonDriftCorrection ChangeReason = "drift_correction"
	ReasonAdoption        ChangeReason = "adoption"
)

type Change struct {
	Status  ChangeStatus `json:"status"`
	File    string       `json:"file"`
	Path    string       `json:"path,omitempty"`
	Reason  ChangeReason `json:"reason"`
	Summary string       `json:"summary"`
	Before  string       `json:"before,omitempty"`
	After   string       `json:"after,omitempty"`
}

type PlanSummary struct {
	Create   int `json:"create"`
	Update   int `json:"update"`
	NoOp     int `json:"no_op"`
	Drift    int `json:"drift"`
	Conflict int `json:"conflict"`
}

type fileDocument struct {
	Raw    string
	Exists bool
	Values map[string]any
}

type Plan struct {
	Desired  DesiredModel  `json:"desired"`
	Resolved ResolvedModel `json:"resolved"`
	State    *State        `json:"state,omitempty"`
	Changes  []Change      `json:"changes"`
	Summary  PlanSummary   `json:"summary"`

	fileDocs            map[string]fileDocument
	stateRaw            string
	gitattributesRaw    string
	gitattributesExists bool
}

func (p Plan) HasConflicts() bool {
	return p.Summary.Conflict > 0
}

func (p Plan) HasActionableChanges() bool {
	return p.Summary.Create > 0 || p.Summary.Update > 0 || p.Summary.Drift > 0
}

type ApplyResult struct {
	Plan  Plan     `json:"plan"`
	Wrote []string `json:"wrote"`
}

type ShowReport struct {
	Desired            DesiredModel                 `json:"desired"`
	Resolved           ResolvedModel                `json:"resolved"`
	State              *State                       `json:"state,omitempty"`
	CurrentOwnedValues map[string]map[string]string `json:"current_owned_values"`
}

type DoctorReport struct {
	Plan      Plan     `json:"plan"`
	HasIssues bool     `json:"has_issues"`
	Messages  []string `json:"messages"`
}

func (m DesiredModel) projectByID(id string) *ProjectSpec {
	for i := range m.Projects {
		if m.Projects[i].ID == id {
			return &m.Projects[i]
		}
	}
	return nil
}

func (m DesiredModel) projectPyprojectPath(project ProjectSpec) string {
	if project.EffectivePath == "." {
		return PyprojectFileName
	}
	return joinProjectPath(project.EffectivePath, PyprojectFileName)
}

func (m ModelSummary) componentVersion(id string) (string, bool) {
	for _, component := range m.Components {
		if component.ID == id {
			return component.Version, true
		}
	}
	return "", false
}
