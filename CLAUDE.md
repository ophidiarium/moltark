# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What is Moltark

Moltark is a project templater powered by Starlark that supports long-term template evolution, not one-shot scaffolding. It bootstraps projects from reusable templates and keeps them aligned as templates evolve. The core model: projects define scope, components define behavior, facts expose truth, providers expose capabilities, routed intents bind consumers to providers, and managed files are the reconciliation surface.

## Build and Test Commands

Use `bazelisk` (not bare `bazel`) as the primary build and test runner so the version pinned in `.bazelversion` is used. Always verify changes with `bazelisk` before pushing — CI runs Bazel, not bare `go test`.

```bash
# Build (prefer bazelisk)
bazelisk build //:moltark           # Bazel build (CI-authoritative)
go build ./cmd/moltark              # Go build (quick local check)

# Test (prefer bazelisk)
bazelisk test //...                  # Full Bazel suite (CI-authoritative)
go test ./...                        # Full Go suite
go test -count=1 ./...               # Full suite (no cache)
go test -count=1 ./internal/engine/...    # Engine tests (pipeline, resolve, plan, state)
go test -count=1 ./internal/filefmt/...   # File format handler tests
go test -count=1 ./internal/module/...    # Starlark module tests
go test -count=1 ./tests/integration/...  # Integration tests only
go test -count=1 ./tests/features/...     # Gherkin feature tests only

# Refresh integration snapshots
UPDATE_SNAPS=true go test -count=1 ./tests/integration/...

# Dependency management
# After changing go.mod (adding/removing/switching deps):
bazelisk run //:gazelle              # Regenerate BUILD.bazel files
bazelisk mod tidy                    # Update MODULE.bazel use_repo directives

# Lint (CI checks)
gofmt -l .                           # Format check
go vet ./...                         # Vet
```

Use `-count=1` when changing snapshots or CLI behavior to avoid stale test-cache results.

## Architecture

**CLI layer**: `cmd/moltark/main.go` -> `internal/cliapp/app.go` -> `internal/command/` (one file per subcommand: init, plan, apply, show, doctor, version). Uses `github.com/mitchellh/cli`.

**Package layout** (acyclic dependency order: model <- filefmt <- module <- engine):

- **`internal/model`** — Shared domain types and constants. All IR types (`DesiredModel`, `ResolvedModel`, `Plan`, `Change`, `State`, etc.), change status/reason enums, file format constants, module source identifiers, and clone helpers. Zero internal dependencies.

- **`internal/filefmt`** — Structured file format handlers. Path resolution for TOML (dot notation) and JSON/YAML (JSON Pointer), format-specific parsers/mutators (TOML, JSON, YAML), `.gitattributes` managed-block logic. Depends on model only.

- **`internal/module`** — Starlark DSL and module system. Config loading (`LoadDesiredModel`, `InitRepository`), module registry, first-party modules (`moltark/core`, `moltark/python`, `astral/uv`), Starlark value conversion, and fact-ref value type. Depends on model + filefmt.

- **`internal/engine`** — Reconciliation engine. Five-phase pipeline (evaluate -> resolve -> inspect -> persist -> plan), service layer (`Plan`, `Apply`, `Show`, `Doctor`), change classification, state management, and plan rendering. Depends on model + filefmt + module.

## Testing Structure

- **Package tests**: `internal/engine/*_test.go` (pipeline, resolve, plan, state), `internal/filefmt/*_test.go` (format handlers, paths), `internal/module/*_test.go` (config loading, core module)
- **Integration snapshots**: `tests/integration/` -- copies fixture repos to temp dirs, runs CLI commands, snapshots output via `go-snaps`
- **Gherkin features**: `tests/features/` -- behavioral scenarios via `godog`
- **Fixtures**: `tests/fixtures/` -- real repository structures (molt.star + pyproject.toml + state.json)
- **Test helpers**: `internal/testutil/` (general), `internal/testrepo/` (Bazel-aware path helpers)

## Bazel / Gazelle

The build is Gazelle-first with bzlmod (`MODULE.bazel`). Gazelle generates package-level `BUILD.bazel` files; hand-maintain only what Gazelle can't infer (fixtures, snapshots, `.feature` file data attributes, repo-level directives). Local Gherkin rules in `tools/bazel/gherkin_defs.bzl` model `.feature` files as first-class Bazel inputs through `gherkin_library` and `godog_feature_test` macros.

## Key Design Constraints

- Moltark is a reconciler. Every feature must answer: how does update work, how is drift detected, how are conflicts surfaced, which user edits are preserved?
- The `apply` command re-runs the pipeline and verifies the plan matches before writing -- never bypass this staleness check.
- First-party modules should use core primitives (facts, providers, routed intents) rather than hardcoding behavior into the engine.
- Generated output must be readable in normal code review.
- Favor update safety over bootstrap convenience.
