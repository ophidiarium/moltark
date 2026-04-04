# Moltark Docs

This directory documents the current Moltark architecture and the intended direction of travel.

The docs are arranged for incremental discovery:

1. Start with [Core Concepts](./concepts/01-core-concepts.md).
2. Then read [Modules And Providers](./concepts/02-modules-and-providers.md).
3. Then read [Execution Model](./concepts/03-execution-model.md).
4. If you are changing tests or adding behavior, read [Testing](./testing.md).
5. If you are evaluating architecture decisions, read [Future Paths](./future-paths.md).
6. If you need prior-art and external comparisons, read [Research Notes](./research/README.md).

## Current Scope

The current MVP implementation is intentionally narrow:

- one root `Moltarkfile` per repository
- Python projects as the first concrete project type
- `astral/uv` as the first concrete provider module
- reconciliation of `pyproject.toml`, `.gitattributes`, and `.moltark/state.json`
- Bazel/Gazelle as the developer build and test loop for this repository
- local Bazel rules for first-class Gherkin feature inputs used by `godog`

The architecture is being shaped for broader module and capability composition, but the implementation is still local-first-party and Python-first today.

## Reading By Task

If you need to understand:

- the reconciler model: [Core Concepts](./concepts/01-core-concepts.md)
- how `use("...")` works today: [Modules And Providers](./concepts/02-modules-and-providers.md)
- the difference between synth hooks, bootstrap requirements, user tasks, and triggers: [Execution Model](./concepts/03-execution-model.md)
- how snapshots and Gherkin tests are organized: [Testing](./testing.md)
- how the repo is built and regenerated locally: see [README.md](../README.md) and [Testing](./testing.md)
- what is intentionally not implemented yet: [Future Paths](./future-paths.md)
- what we learned from adjacent tools and internal notes: [Research Notes](./research/README.md)
