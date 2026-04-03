# AGENTS.md

## Purpose

Moltark is a modern software project templater with first-class support for updates and template evolution.

The goal is not just to scaffold a repository once, but to let projects continue adopting improvements from shared templates over time.

Agents working in this repository should optimize for that long-term goal.

## Product direction

Moltark should evolve toward these outcomes:

- bootstrap new projects from reusable templates
- compose templates from smaller building blocks
- update existing projects to newer template versions
- keep generated output understandable, reviewable, and project-owned
- make project setup reusable at team and organization scale

Moltark is not a one-shot generator. Its core value is maintaining the relationship between a project and its evolving template source.

## Initial scope

Near-term implementation should focus on:

- Go CLI
- Starlark module system
- bootstrap workflow
- update and reconciliation engine

Initial ecosystem support is intentionally narrow:

- Go
- Python
- Rust

This narrow scope is deliberate. Prefer depth and solid architecture over broad early ecosystem coverage.

## Planned evolution

The next architectural expansion areas are:

- Bazel, where Starlark-native and code-driven configuration are a natural fit
- support for ecosystems with code-based configuration, starting with TypeScript and Ruby
- extensible config stubs that can evolve from generated defaults into maintainable project-owned code

When making design choices, avoid baking in assumptions that only work for static file templating.

## Core principles

### 1. Updates are first-class

Do not treat updates as an afterthought.

When designing APIs, state models, template metadata, file rendering, or CLI flows, always ask:

- how will an existing project adopt a newer template version?
- how will drift be detected?
- how will conflicts be surfaced?
- how will user-owned modifications be preserved?

If a design is good for initial generation but weak for later reconciliation, it is incomplete.

### 2. Generated output must remain project-owned

Moltark should help maintain projects, not trap them behind opaque machinery.

Favor designs where:

- generated files stay readable
- users can understand what was written and why
- changes are reviewable in normal code review
- ownership stays with the repository, not hidden runtime state

Avoid designs that require users to trust unexplained magic.

### 3. Composition matters more than monoliths

Templates should be composable from smaller reusable building blocks.

Favor:

- modular template primitives
- explicit composition boundaries
- small reusable units over giant all-in-one templates

Do not let the system collapse into a collection of large, rigid project archetypes.

### 4. Governance is a real use case

Moltark is not just a convenience scaffold tool. It can become a governance mechanism for teams and organizations.

Features should make it possible to standardize and evolve things like:

- linting and formatting
- spell checking
- license validation
- duplication detection
- complexity monitoring
- CI/CD setup
- git hooks
- release automation
- repository metadata
- security and compliance defaults

Design toward reusable organizational defaults, not only individual-project ergonomics.

### 5. AI-era workflows need stronger guardrails

Moltark should be useful in projects increasingly changed by coding agents.

That means the repository setup it produces should support:

- strict, automated feedback loops
- maintainability checks in CI/CD
- clear machine-readable structure
- conventions that reduce drift
- evolving quality controls over time

Do not optimize only for human manual setup. Optimize for long-lived repositories that will be modified by both humans and agents.

### 6. Determinism and reviewability matter

Prefer deterministic behavior and explicit planning over hidden side effects.

When possible, structure features so users can inspect:

- what Moltark believes the desired state is
- what changed since last application
- what update is about to happen
- where conflicts or ambiguity exist

A good update engine should make change visible before it makes change automatic.

## Design guidance

### Template model

The template system should eventually support more than raw file rendering.

Design toward a model that can express:

- file generation
- file updates
- merge/reconciliation behavior
- ownership boundaries
- metadata needed for future updates

Avoid reducing the system to “copy files with variables.”

### Starlark usage

Starlark is important, but it is not the product by itself.

Use Starlark where it improves:

- composability
- determinism
- portability
- understandable configuration logic

Do not introduce unnecessary complexity just to make the system feel more dynamic.

### Code-based config support

Future support for Bazel, TypeScript, and Ruby means Moltark will need to handle code-based configuration, not just static formats.

Keep this in mind when designing abstractions around:

- config stubs
- file ownership
- safe updates
- structural edits
- reconciliation strategies

Do not assume every target file can be safely rewritten from scratch.

### CLI UX

The CLI should feel predictable and professional.

Favor commands and output that help users answer:

- what will be generated?
- what will be updated?
- what changed?
- what requires manual intervention?
- how is this project related to its template modules?

Avoid vague success messages and hidden mutations.

## Implementation guidance

### Prefer architecture over surface area

Early on, resist adding many ecosystems, many commands, or many flags.

A smaller system with a sound template/update model is more valuable than a feature-rich scaffold tool with no coherent reconciliation story.

### Keep internal models explicit

Prefer explicit internal representations for:

- project state
- template state
- desired outputs
- update plans
- drift/conflict detection

Avoid spreading critical behavior across ad hoc string manipulation.

### Preserve room for structural editing

Some future targets will require updating structured or code-based config files.

Where possible, keep room for AST/CST-based or structure-aware editing rather than assuming line-based patching is sufficient.

## Testing approach

Testing should be **fixture-driven**.

Prefer tests that describe a project state, apply Moltark behavior, and assert the resulting outcome through highly observable artifacts. The goal is to make tests easy to read, resilient during large refactorings, and strongly aligned with real product behavior.

Favor a model like:

- **input fixture** → project/template/module state before execution
- **operation** → generate, update, reconcile, detect drift, etc.
- **observable outcome** → resulting files, plans, diagnostics, conflicts, or snapshots

### Testing principles

- Prefer **fixture-based integration tests** over narrow implementation-detail unit tests for core behavior
- Keep tests **human-readable** and easy to inspect in code review
- Optimize for **observability**: when a test fails, it should be obvious what changed and why
- Preserve **relevance across refactorings**: tests should validate behavior and outputs, not incidental internal structure
- Make update/reconciliation scenarios first-class, not just initial generation

### Preferred tools and styles

Steer the test suite toward:

- **`go-snaps`** for snapshot-based assertions where the output is naturally textual or structured
- **Gherkin feature scenarios** where behavior benefits from explicit Given/When/Then framing

These should be used to make behavior legible, not to hide complexity.

### What to test

Core scenarios should include:

- bootstrap from fixtures
- re-apply without changes
- template evolution across versions
- drift detection
- conflict surfacing
- preservation of intended user edits
- update planning and execution
- diagnostics and user-visible output

### Test design guidance

- Prefer a small number of **high-signal fixtures** over many shallow tests
- Keep fixture names descriptive and scenario-oriented
- Snapshot the most useful observable artifacts: generated files, update plans, diffs, diagnostics
- Avoid brittle assertions on incidental formatting unless formatting is the behavior being tested
- Ensure failures are easy to understand without stepping through implementation details

A Moltark test should read like an example of real repository evolution, not like a microscopic probe into private internals.

## Non-goals for now

Do not prematurely optimize for:

- supporting every programming language
- becoming a general-purpose build tool
- reproducing Projen exactly
- implementing a full IDE/editor integration story
- solving every config-language mutation problem in v1

The immediate goal is to establish a strong foundation for bootstrap + update in a focused set of ecosystems.

## When uncertain

When facing design ambiguity, prefer the option that better supports:

1. long-term template evolution
2. understandable generated output
3. deterministic and reviewable updates
4. composable organizational defaults
5. future support for code-based configuration
