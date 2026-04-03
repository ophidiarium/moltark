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

Supported ecosystems for now:

- Go
- Python
- Rust

Do not optimize early for broad ecosystem support.

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
