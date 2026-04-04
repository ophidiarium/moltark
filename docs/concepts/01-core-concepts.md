# Core Concepts

Moltark is a reconciler, not a one-shot generator.

The current system now exposes explicit engine phases:

1. `evaluate`: load `Moltarkfile` into a desired in-memory model.
2. `resolve`: turn module-provided files, facts, providers, and routed intents into a concrete reconciled model.
3. `inspect`: read the current repository surfaces Moltark cares about.
4. `persist`: compute the candidate next `.moltark/state.json`.
5. `plan`: classify changes from desired state, current state, and candidate persisted state.

`apply` then executes only the managed file edits that are safe and explainable.

This matters because Moltark is not just a renderer. The phase boundaries are part of the product: they make drift, conflict, adoption, and future migrations easier to reason about and expose through CLI diagnostics.

## Repository Model

Today the desired model has two main layers:

- projects
- components

Projects describe repository-contained units with:

- `id`
- `kind`
- `path`
- `parent`
- `effective_path`
- optional `attributes`

Components are contributed by modules and compile into:

- facts
- managed file intents
- capability providers
- routed intents
- synthesis hooks
- bootstrap requirements
- user tasks
- task surfaces
- trigger bindings

The important separation is:

- projects describe containment and location
- components describe behavior and ownership

Facts are the lightweight project-scoped truth layer between those two.

They let one component say:

- this target supports Go `1.24`
- this Python project requires `>=3.12`
- this repository targets `buildkit`

without forcing every consuming component to know which concrete module or component produced that information.

Projects do not need to be heavyweight archetypes.

They can be lightweight anchors that exist only to give components:

- a target path
- a provider scope
- a parent-relative location
- a stable identity in state and diagnostics

That means a repository can declare a minimal project anchor and then compose only the shared components it actually wants Moltark to manage.

## Ownership

Ownership is explicit and path-scoped.

For the current Python MVP:

- Moltark owns a minimal `pyproject.toml` surface
- Moltark owns a managed block in `.gitattributes`
- Moltark owns `.moltark/state.json`

For `pyproject.toml`, Moltark currently owns only selected paths such as:

- `project.name`
- `project.version`
- `project.requires-python`
- `build-system.*`
- `tool.moltark.*`
- `tool.uv.workspace.members` when an `uv_workspace` is declared

It intentionally does not own:

- `project.dependencies`

That preserves user or tool changes made through `uv add` / `uv remove`.

Generic structured-file primitives are available through `moltark/core`:

- `json_file(target=..., path=..., values=...)`
- `toml_file(target=..., path=..., values=...)`
- `yaml_file(target=..., path=..., values=...)`

That enables component-oriented ownership for files like `.vscode/settings.json`, `typos.toml`, or `.github/labeler.yml` without requiring a large ecosystem archetype.

## Planning Semantics

Plans classify changes as:

- `create`
- `update`
- `no-op`
- `drift`
- `conflict`

The planner compares:

- desired owned values
- current repository values
- last applied state fingerprints

This is what allows Moltark to distinguish:

- desired state changes
- template version changes
- drift in owned surfaces
- conflicts that are unsafe to merge automatically

## State

`.moltark/state.json` exists to make reconciliation explicit.

It currently tracks:

- schema version
- managed files
- owned paths
- per-owned-path template versions when a component provides them
- fingerprints
- a summary of the last applied desired model

State is required for:

- no-op re-apply
- drift detection
- conflict surfacing
- template evolution

Template evolution is no longer modeled as one repo-wide version string.

Instead, Moltark tracks component versions and owned-path versions where a component provides them. That lets a Python project component evolve independently from a generic JSON-only component or a future Go-specific component.

## Current Limitation

The current implementation is still ecosystem-light.

The internal model now supports generic JSON, TOML, and YAML structured-file reconciliation in addition to Python `pyproject.toml` ownership. What remains narrow is the ecosystem layer above those primitives: Python and `uv` are still the only first-party modules that compile richer capability relationships into those managed files. That is an intentional staging choice, not the intended final product shape.
