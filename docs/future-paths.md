# Future Paths

This page records important architectural direction that has already influenced the current implementation, but is not fully implemented yet.

## 1. Remote Modules

Today `use("...")` loads first-party local modules through a Go registry.

The intended future direction is:

- GitHub-backed module sources
- version pinning
- local module cache
- explicit lock state

The current local module runtime is a staging step toward that, not the final module story.

## 2. First-Party Modules Built On More Generic Primitives

Today the first-party modules still compile their semantics in Go.

The intended direction is to keep pushing module semantics toward generic primitives such as:

- structured files
- owned paths
- providers
- routed intents
- tasks
- trigger bindings

That is how `astral/uv` can eventually stop being a special first-party Go implementation and become “just another module” on top of shared Moltark building blocks.

## 3. Polyglot Monorepos

Moltark should support one root repository workspace per `Moltarkfile`, not many unrelated workspaces in one file.

Within that root, it should support:

- multiple projects
- parent-relative project paths
- repeated subprojects
- ecosystem-specific workspace modules

The current parent-relative project model is already shaped with this in mind.

## 4. Provider-Based Composition

The provider model exists because reusable components need discovery without hardcoded knowledge of the concrete tool chosen by a project.

Examples that motivated the design:

- a linter wanting to install via `uv`, `npm`, or `bundler`
- a hook component wanting to bind tasks into git hooks
- a CI component wanting to bind tasks into workflows

The current `uv` package-manager and workspace-manager providers are the first concrete demonstration of that direction.

## 5. Execution Beyond Reconciliation

The system now distinguishes:

- synthesis hooks
- bootstrap requirements
- user tasks
- task surfaces
- trigger bindings

The next major step is to decide which of those become executable by Moltark, and which remain declarative artifacts for external providers to realize.

That step should be taken carefully. Execution should not be added in a way that blurs:

- reconciliation-time behavior
- environment bootstrapping
- user-invoked work
- external triggers

## 6. Dependency Ownership

The current MVP intentionally does not mutate `project.dependencies`.

That protects `uv add` / `uv remove` workflows and preserves a clear user-owned surface.

The future path is not “take over dependency sections blindly.” It is to introduce dependency mutation only when Moltark has a clear ownership and mediation model that keeps package-manager workflows intact.
