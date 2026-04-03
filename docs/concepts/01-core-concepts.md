# Core Concepts

Moltark is a reconciler, not a one-shot generator.

The current system follows this pipeline:

1. Evaluate `Moltarkfile`.
2. Build a desired in-memory project model.
3. Resolve module-provided capabilities and routed intents into a concrete reconciled model.
4. Inspect the repository state.
5. Produce a plan.
6. Apply only the managed file edits that are safe and explainable.

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

Components are contributed by modules and compile into:

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
- template version
- managed files
- owned paths
- fingerprints
- a summary of the last applied desired model

State is required for:

- no-op re-apply
- drift detection
- conflict surfacing
- template evolution

## Current Limitation

The current implementation is still Python-first.

The internal model is more general than the file mutators, but only the Python `pyproject.toml` path is actually reconciled today. That is an intentional staging choice, not the intended final product shape.
