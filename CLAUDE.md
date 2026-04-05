# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Moltark

Moltark is a project templater powered by Starlark that supports long-term template evolution, not one-shot scaffolding. It bootstraps projects from reusable templates and keeps them aligned as templates evolve. The core model: projects define scope, components define behavior, facts expose truth, providers expose capabilities, routed intents bind consumers to providers, and managed files are the reconciliation surface.

## Build and Test Commands

Use `bazelisk` (not bare `bazel`) so the version pinned in `.bazelversion` is used.

```bash
# Build
go build ./cmd/moltark              # Go build
bazelisk build //:moltark           # Bazel build

# Test
go test ./...                        # Full Go suite
go test -count=1 ./...               # Full suite (no cache)
bazelisk test //...                  # Full Bazel suite
go test -count=1 ./internal/moltark/...   # Core package tests
go test -count=1 ./tests/integration/...  # Integration tests only
go test -count=1 ./tests/features/...     # Gherkin feature tests only

# Refresh integration snapshots
UPDATE_SNAPS=true go test -count=1 ./tests/integration/...

# Regenerate BUILD.bazel files
bazelisk run //:gazelle

# Lint (CI checks)
gofmt -l .                           # Format check
go vet ./...                         # Vet
```

Use `-count=1` when changing snapshots or CLI behavior to avoid stale test-cache results.

## Architecture

**CLI layer**: `cmd/moltark/main.go` -> `internal/cliapp/app.go` -> `internal/command/` (one file per subcommand: init, plan, apply, show, doctor, version). Uses `github.com/mitchellh/cli`.

**Engine pipeline** (`internal/moltark/pipeline.go`): Five sequential phases:
1. **Evaluate** - load `Moltarkfile` via Starlark (`config.go`) into `DesiredModel` (projects + components)
2. **Resolve** - resolve facts, providers, routed intents (`resolve.go`) into `ResolvedModel` with managed files
3. **Inspect** - read current repo state: structured files (TOML/JSON/YAML), `.moltark/state.json`, `.gitattributes`
4. **Persist** - build next state from desired+resolved model
5. **Plan** - classify each owned path as create/update/no-op/drift/conflict (`plan.go`)

**Service** (`internal/moltark/service.go`): Orchestrates `Plan`, `Apply`, `Show`, `Doctor` operations. Apply re-runs the full pipeline and verifies intent hasn't changed before writing.

**First-party modules** (`internal/moltark/module_*.go`): `moltark/core`, `moltark/python`, `astral/uv`. Go and Rust are target ecosystems but do not yet have first-party module depth.

**Structured file mutation**: Format-specific mutators for TOML (`pyproject.go`), JSON (`jsonfile.go`), YAML (`yamlfile.go`) that write only owned paths.

**Types** (`internal/moltark/types.go`): All IR types -- `DesiredModel`, `ResolvedModel`, `Pipeline`, `Plan`, `Change`, `State`, etc.

## Testing Structure

- **Package tests**: `internal/moltark/*_test.go` -- planner, resolver, mutator, state logic
- **Integration snapshots**: `tests/integration/` -- copies fixture repos to temp dirs, runs CLI commands, snapshots output via `go-snaps`
- **Gherkin features**: `tests/features/` -- behavioral scenarios via `godog`
- **Fixtures**: `tests/fixtures/` -- real repository structures (Moltarkfile + pyproject.toml + state.json)
- **Test helpers**: `internal/testutil/` (general), `internal/testrepo/` (Bazel-aware path helpers)

## Bazel / Gazelle

The build is Gazelle-first with bzlmod (`MODULE.bazel`). Gazelle generates package-level `BUILD.bazel` files; hand-maintain only what Gazelle can't infer (fixtures, snapshots, `.feature` file data attributes, repo-level directives). Local Gherkin rules in `tools/bazel/gherkin_defs.bzl` model `.feature` files as first-class Bazel inputs through `gherkin_library` and `godog_feature_test` macros.

## Key Design Constraints

- Moltark is a reconciler. Every feature must answer: how does update work, how is drift detected, how are conflicts surfaced, which user edits are preserved?
- The `apply` command re-runs the pipeline and verifies the plan matches before writing -- never bypass this staleness check.
- First-party modules should use core primitives (facts, providers, routed intents) rather than hardcoding behavior into the engine.
- Generated output must be readable in normal code review.
- Favor update safety over bootstrap convenience.
