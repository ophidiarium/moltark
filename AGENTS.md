# AGENTS.md

## Mission

Moltark is a project templater for long-term template evolution, not one-shot scaffolding.

Optimize decisions for:

- bootstrapping projects from reusable templates
- composing templates from smaller building blocks
- updating existing projects as templates evolve
- keeping generated output readable, reviewable, and project-owned
- enabling team and organization defaults to evolve over time

If a choice helps initial generation but weakens future updates, it is the wrong default.

## Current focus

Build depth before breadth.

Prioritize:

- Go CLI
- Starlark module system
- bootstrap workflow
- update and reconciliation engine

Target ecosystems for near-term design work:

- Go
- Python
- Rust

Do not optimize early for broad ecosystem support.

Current first-party implementation reality:

- one root `Moltarkfile` per repository
- first-class core primitives live in `moltark/core`
- current first-party ecosystem modules are `moltark/python` and `astral/uv`
- generic structured-file reconciliation exists for JSON, TOML, and YAML
- Go and Rust are important target ecosystems, but they are not yet first-party module families in the repo today

Do not describe the current codebase as if Go and Rust already have the same first-party module depth as Python.

## Start here

If you are new to the codebase, read in this order:

1. [`docs/README.md`](./docs/README.md)
2. [`docs/concepts/01-core-concepts.md`](./docs/concepts/01-core-concepts.md)
3. [`docs/concepts/02-modules-and-providers.md`](./docs/concepts/02-modules-and-providers.md)
4. [`docs/concepts/03-execution-model.md`](./docs/concepts/03-execution-model.md)
5. [`docs/testing.md`](./docs/testing.md)

Then orient in code with:

- [`internal/moltark/types.go`](./internal/moltark/types.go) for the core IR
- [`internal/moltark/config.go`](./internal/moltark/config.go) for `Moltarkfile` evaluation
- [`internal/moltark/modules.go`](./internal/moltark/modules.go) for first-party module loading
- [`internal/moltark/resolve.go`](./internal/moltark/resolve.go) for provider, fact, and intent resolution
- [`internal/moltark/plan.go`](./internal/moltark/plan.go) for change classification
- [`internal/moltark/service.go`](./internal/moltark/service.go) for `plan` / `apply` orchestration

The main mental model is:

- projects define containment and scope
- components define behavior and ownership
- facts expose project-scoped truth
- providers expose capabilities
- routed intents bind consumers to providers
- managed files are the concrete reconciliation surface

## Default stance

Prefer:

- update-safe models over one-shot generation shortcuts
- explicit state and plans over hidden side effects
- composable building blocks over large rigid archetypes
- deterministic behavior over convenience magic
- readable repo-owned output over opaque internal machinery
- designs that leave room for structural editing, not just file rewrites

Avoid:

- treating updates as an afterthought
- ad hoc string manipulation for core state transitions
- assumptions that every target file can be regenerated from scratch
- large archetype templates when smaller modules would do
- vague CLI behavior or hidden mutations

## Required design questions

For any feature touching templates, rendering, metadata, state, reconciliation, or CLI UX, answer:

1. How does an existing project adopt a newer template version?
2. How is drift detected?
3. How are conflicts surfaced?
4. Which user-owned edits are preserved?
5. What can the user inspect before changes are applied?

If these answers are weak, the design is incomplete.

## Design rules

### Template model

Design for more than file copying. The model should leave room for:

- file generation
- file updates and reconciliation
- ownership boundaries
- metadata needed for future upgrades
- structured edits where rewrite-from-scratch is unsafe

Do not reduce Moltark to "copy files with variables."

### Starlark

Use Starlark where it improves:

- composition
- determinism
- portability
- understandable configuration logic

Do not add complexity just to make the system feel dynamic.

### Code-based configuration

Future support for Bazel, TypeScript, and Ruby means code-based config is a real target.

Keep room for:

- config stubs that can evolve into project-owned code
- structural edits
- ownership-aware updates
- safe reconciliation of partially user-edited files

Do not assume line-based patching or full rewrites will always be sufficient.

### CLI UX

CLI output should help users answer:

- what will be generated
- what will be updated
- what changed
- what requires manual intervention
- how the project relates to its template modules

Favor explicit plans and diagnostics over vague success messages.

## Implementation rules

- Favor architecture over surface area. A smaller coherent system is better than more commands, flags, or ecosystems.
- Keep internal models explicit: project state, template state, desired outputs, update plans, drift, and conflicts.
- Generated output must stay understandable in normal code review.
- Preserve room for AST, CST, or structure-aware editing where needed.
- When choosing between bootstrap convenience and update safety, choose update safety.

## Governance and AI workflows

Moltark should support repositories maintained by both humans and coding agents.

Prefer designs that improve:

- strong CI feedback loops
- maintainability and policy checks
- machine-readable structure
- reusable team and organization defaults
- reduced drift over time

This is not just a scaffold tool. It should be able to carry evolving standards across many repositories.

## Testing

Testing should be fixture-driven and behavior-oriented.

Primary docs:

- [`docs/testing.md`](./docs/testing.md)
- [`docs/concepts/01-core-concepts.md`](./docs/concepts/01-core-concepts.md)
- [`docs/concepts/02-modules-and-providers.md`](./docs/concepts/02-modules-and-providers.md)
- [`docs/concepts/03-execution-model.md`](./docs/concepts/03-execution-model.md)
- [`docs/future-paths.md`](./docs/future-paths.md)

Actual test layers in this repo:

- package-level tests under `internal/moltark/` for planner, resolver, structured-file mutation, and state logic
- integration snapshot tests under `tests/integration/`
- Gherkin feature tests under `tests/features/`

Prefer:

- input fixture -> operation -> observable outcome
- integration-style tests for core behavior
- `go-snaps` when outputs are naturally textual or structured
- Gherkin scenarios when Given/When/Then improves readability

Core scenarios should cover:

- bootstrap from fixtures
- re-apply with no changes
- template version upgrades
- drift detection
- conflict surfacing
- preservation of intended user edits
- update planning and execution
- user-visible diagnostics

A good test should read like a real repository evolution scenario and fail with obvious artifacts.

Current test framework details:

- integration snapshots use `github.com/gkampitakis/go-snaps`
- Gherkin scenarios use `github.com/cucumber/godog`
- package tests use the standard Go `testing` package
- fixtures live under `tests/fixtures/`
- integration tests live under `tests/integration/`
- snapshot files live under `tests/integration/__snapshots__/`
- feature tests live under `tests/features/`

Current commands:

- full suite: `go test -count=1 ./...`
- core package tests: `go test -count=1 ./internal/moltark/...`
- integration only: `go test -count=1 ./tests/integration/...`
- feature only: `go test -count=1 ./tests/features/...`
- refresh integration snapshots: `UPDATE_SNAPS=true go test -count=1 ./tests/integration/...`

Use `-count=1` when changing snapshots or CLI behavior to avoid stale test-cache confusion.

When adding behavior:

- prefer package tests for narrow planner / resolver / mutator logic
- prefer integration snapshots for user-visible CLI and repo-state behavior
- prefer Gherkin only when the repository evolution story reads better as a scenario than as raw snapshots
- if the feature changes ownership, provider resolution, facts, or plan/apply semantics, add at least one repository-level scenario

## Agent checklist

Before finishing meaningful work, check:

- does this improve or at least preserve the update and reconciliation story
- are state transitions explicit instead of hidden in ad hoc behavior
- would the generated or updated output make sense in normal code review
- if behavior changed, is there fixture-driven coverage for the relevant scenario
- did the change stay within current scope instead of expanding surface area by default

## Non-goals for now

Do not prematurely optimize for:

- every language or ecosystem
- becoming a general-purpose build tool
- reproducing Projen
- full IDE or editor integration
- solving every config-language mutation problem in v1

## When uncertain

Choose the option that best supports, in order:

1. long-term template evolution
2. understandable project-owned output
3. deterministic and reviewable updates
4. composable organizational defaults
5. future code-based configuration support
