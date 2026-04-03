# Execution Model

One source of confusion in systems like Moltark is overloading the word `task`.

Moltark now separates several categories that should not be collapsed together.

## 1. Synthesis Hooks

These are Moltark-driven lifecycle hooks around project synthesis.

They are conceptually similar to:

- CDK aspects or synth-time structure inspection
- Projen `preSynthesize` / `synth` / `postSynthesize`

These belong to Moltark's own reconciliation lifecycle.

Current phases are:

- `pre_synthesize`
- `synthesize`
- `post_synthesize`

## 2. Bootstrap Requirements

These describe runtime or machine prerequisites, for example:

- install `uv`
- install `bun`
- install `cargo-nextest`
- ensure a package-manager lifecycle step exists

These are not ordinary user tasks. They are environment or workspace bootstrapping concerns.

## 3. User Tasks

These are user-invoked units of work such as:

- `test`
- `lint`
- `typecheck`
- `build`

These should be the normal thing a developer thinks of as a task.

## 4. Task Surfaces

A task surface is how tasks become visible or invokable through some external system, for example:

- `package.json` scripts
- `just`
- `go-task`
- `mise`

The task is the semantic unit. The task surface is just one way to expose it.

## 5. Trigger Bindings

Trigger bindings connect tasks to event-driven surfaces such as:

- `pre-commit`
- `pre-push`
- CI
- internal custom triggers like `pre-cr`

This is how Git hooks and CI fit into the model without pretending they are the same thing as user tasks.

## What Is Executed Today

Only managed file reconciliation is executed today.

That means:

- `pyproject.toml` reconciliation is executed
- `.gitattributes` reconciliation is executed
- `.moltark/state.json` updates are executed

The following are currently modeled and resolved, but not executed:

- synthesis hooks
- bootstrap requirements
- user tasks
- task surfaces
- trigger bindings

This distinction is important. The model now has clear categories, but Moltark does not yet contain a command runner or bootstrap engine for them.

## Why Keep Them Anyway

Because the categories matter even before execution exists.

They make the architecture clearer for:

- module authors
- future capability providers
- `show` output
- plan-time diagnostics

That lets the system mature without first baking execution semantics into a muddy task model.
