# Modules And Providers

Moltark is moving toward a module-oriented architecture.

Today, `use("...")` loads first-party local modules through a local registry. That registry lives in the Go implementation, not in remote module fetching yet.

Current first-party module sources are:

- `moltark/core`
- `moltark/python`
- `astral/uv`

`moltark/core` is intentionally becoming the place for low-level, reusable primitives such as:

- lightweight project anchors
- facts and fact references
- structured-file resources
- bootstrap requirements
- user tasks
- trigger bindings

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

## Facts

Facts are different from providers.

- providers answer: who can do something?
- facts answer: what is true about this target or scope?

Examples:

- a Python project can publish its supported Python range
- a Go project can publish its target Go version
- a container-oriented component can publish whether the repo targets BuildKit, Podman, or Containerd

Facts are resolved by nearest project scope, just like providers, but they are consumed as values rather than invoked as handlers.

Current core surface:

- `core.fact(name=..., target=..., values=...)`
- `core.fact_value(name=..., target=..., path=...)`

The important constraint in the current implementation is that fact references are resolved only inside structured-file values. That is enough for config-driven components like linters, editor settings, and labelers while keeping the initial semantics narrow.

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

`moltark/core` also now supports component-oriented authoring for non-Python cases through:

- `core.project(...)`
- `core.json_file(...)`
- `core.toml_file(...)`
- `core.yaml_file(...)`
- `core.task(...)`
- `core.bootstrap_requirement(...)`
- `core.trigger_binding(...)`

That means a repository can be anchored with a small generic project and then managed only through the shared components it actually opts into.

Those structured-file primitives are intentionally first-class in core because many shareable components are mostly about ownership-aware repo config, not full ecosystem archetypes. A spellchecker, editor-policy component, or labeler component may only need to write TOML, JSON, or YAML.

## Important Current Nuance

`python_dependency_request` is intentionally resolved but not yet applied to `project.dependencies`.

That is not an accidental omission. It preserves the current MVP rule that dependency lists remain user- or package-manager-owned.

So the current provider concept is demonstrated in `show` and in resolved intent bindings before dependency mutation is introduced.

## Current Limitation

Modules are still local-first-party.

Today there is no:

- GitHub-backed module fetching
- version lock file
- remote module cache

That is a future path, not current behavior.

## Minimal Component-Oriented Example

A small repo that only wants Go linting ergonomics in VS Code does not need a large generated Go project template.

It can use a lightweight project anchor and just the components it wants:

```python
core = use("moltark/core")

app = core.project(
    id = "app",
    kind = "go_project",
    path = ".",
    attributes = {
        "go_version": "1.24",
    },
)

core.bootstrap_requirement(
    tool = "golangci-lint",
    target = app,
)

core.task(
    name = "lint",
    target = app,
    command = ["golangci-lint", "run", "./..."],
    runtime = "go",
    tags = ["lint", "go"],
)

core.fact(
    name = "moltark.language.go",
    target = app,
    values = {
        "version": "1.24",
    },
)

core.trigger_binding(
    trigger = "pre-commit",
    target = app,
    match_names = ["lint"],
)

core.json_file(
    target = app,
    path = ".vscode/settings.json",
    values = {
        "go.lintTool": "golangci-lint",
        "go.lintOnSave": "package",
    },
)

core.toml_file(
    target = app,
    path = ".golangci.toml",
    values = {
        "run": {
            "go": core.fact_value(
                target = app,
                name = "moltark.language.go",
                path = "version",
            ),
        },
    },
)
```

That is the intended direction: keep the internal tree for scope and resolution, but let the user-facing surface stay component-first.

An even smaller component-only case can stay completely generic:

```python
core = use("moltark/core")

repo = core.project(
    id = "repo",
    kind = "workspace",
    path = ".",
)

core.toml_file(
    target = repo,
    path = "typos.toml",
    values = {
        "files": {
            "extend-exclude": ["vendor/**"],
        },
    },
)

core.yaml_file(
    target = repo,
    path = ".github/labeler.yml",
    values = {
        "docs": {
            "changed-files": {
                "any-glob-to-any-file": ["docs/**"],
            },
        },
    },
)
```

That repo does not need a language-specific project model at all. The anchor exists only to scope ownership and future provider resolution.

## Opaque Projects

Not every project anchor needs language-specific metadata.

For a polyglot repository root that is mostly a container for subprojects, a generic opaque root is enough:

```python
core = use("moltark/core")
python = use("moltark/python")

repo = core.project(
    id = "repo",
    kind = "workspace",
    path = ".",
)

python.python_project(
    id = "backend",
    parent = repo,
    name = "backend",
    path = "backend",
    version = "0.1.0",
    requires_python = ">=3.12",
)
```

The root anchor exists for:

- containment
- provider scope
- path resolution
- diagnostics and state

It does not need to own files itself unless some component targets it.
