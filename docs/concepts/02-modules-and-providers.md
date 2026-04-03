# Modules And Providers

Moltark is moving toward a module-oriented architecture.

Today, `use("...")` loads first-party local modules through a local registry. That registry lives in the Go implementation, not in remote module fetching yet.

Current first-party module sources are:

- `moltark/core`
- `moltark/python`
- `astral/uv`

## Why Modules Matter

The core evaluator should not permanently own ecosystem concepts like:

- `uv_workspace`
- `bun_project`
- `cargo_workspace`

Those belong to modules.

The current code already reflects that boundary better than the earlier shape:

- the central evaluator loads a module namespace
- each module records declarations
- each module later builds components into the shared IR

That keeps the evaluator smaller and makes the module seam explicit even before remote fetching exists.

## Providers

A provider is a capability exposed by a component.

Current capability examples:

- `moltark.python.workspace_manager`
- `moltark.python.package_manager`

Providers are scoped to projects. Resolution walks the project-parent chain outward and chooses the nearest matching provider.

That lets one component ask for a capability without hardcoding the concrete tool that satisfies it.

## Current `uv` Example

`astral/uv` currently demonstrates two provider roles:

1. workspace manager
2. Python package manager

That means `uv` can satisfy:

- workspace membership requests
- Python dependency requests

The important detail is that these are resolved as capability relationships, not as ad hoc branching in the planner.

## Routed Intents

Some component requests are modeled as routed intents.

Current routed intent kinds include:

- `workspace_members_request`
- `python_dependency_request`

The resolution model is:

1. a component emits a routed intent
2. the resolver finds a provider for the requested capability
3. the resolver either lowers that relationship into managed file changes or validates the request against provider metadata

Current examples:

- `uv_workspace` emits a workspace-members request
- `core.python_dependency(...)` emits a Python dependency request

## Important Current Nuance

`python_dependency_request` is intentionally resolved but not yet applied to `project.dependencies`.

That is not an omission by accident. It preserves the current MVP rule that dependency lists remain user or package-manager owned.

So the current provider concept is demonstrated in `show` and in resolved intent bindings before dependency mutation is introduced.

## Current Limitation

Modules are still local-first-party.

Today there is no:

- GitHub-backed module fetching
- version lock file
- remote module cache

That is a future path, not current behavior.
