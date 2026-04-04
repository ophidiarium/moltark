# moltark

**Moltark** is a modern software project templater, powered by Starlark, with first-class support for updates and template evolution.

It helps teams bootstrap new software projects from reusable templates and keep them aligned as those templates evolve over time.

> **Why “moltark”?**  
> Like a snake molting its skin, a project can adopt a new shape without losing its identity.  
> Moltark is about evolving projects over time, not just generating them once.

## Status

Moltark is at an early stage of development.

This repository is in its initial setup phase while the implementation is being built. The published packages are early placeholder releases and do not yet provide the intended functionality.

## Documentation

Architecture and testing notes live in [`docs/`](./docs/README.md).

Recommended reading order:

- [`docs/README.md`](./docs/README.md)
- [`docs/concepts/01-core-concepts.md`](./docs/concepts/01-core-concepts.md)
- [`docs/concepts/02-modules-and-providers.md`](./docs/concepts/02-modules-and-providers.md)
- [`docs/concepts/03-execution-model.md`](./docs/concepts/03-execution-model.md)
- [`docs/testing.md`](./docs/testing.md)

For the current component-centric direction, start with:

- lightweight project anchors in [`docs/concepts/01-core-concepts.md`](./docs/concepts/01-core-concepts.md)
- module/provider boundaries, first-class `json_file` / `toml_file` / `yaml_file` primitives, the minimal Go + VS Code example, and opaque root projects in [`docs/concepts/02-modules-and-providers.md`](./docs/concepts/02-modules-and-providers.md)

## Development

The repo now includes a Bazel build maintained with Gazelle using modern bzlmod configuration.

Use [`bazelisk`](https://github.com/bazelbuild/bazelisk) so the version pinned in [`.bazelversion`](./.bazelversion) is selected automatically.

Common commands:

- regenerate and normalize Go `BUILD.bazel` files: `bazelisk run //:gazelle`
- build the CLI: `bazelisk build //:moltark`
- run the Bazel suite: `bazelisk test //...`
- run the Go suite directly: `go test ./...`

The Bazel setup is intentionally Gazelle-first:

- `MODULE.bazel` owns external dependency resolution through `rules_go` and Gazelle's `go_deps`
- the root [`BUILD.bazel`](./BUILD.bazel) owns the Gazelle entrypoint and repo-level directives
- package-level `BUILD.bazel` files are generated and then only hand-maintained where Gazelle cannot infer runtime data such as fixtures, snapshots, and `.feature` files

For Gherkin specifically:

- feature files are modeled as first-class Bazel inputs through local `gherkin_library(...)` rules in [`tools/bazel/gherkin_defs.bzl`](./tools/bazel/gherkin_defs.bzl)
- execution still goes through the existing Go `godog` suite
- the local `godog_feature_test(...)` macro wires transitive `.feature` files into Bazel runfiles without pulling in a separate Ruby or C++ runtime

## The problem

Most project generators are good at one thing: creating a starting point.

That is useful, but incomplete.

Once a project exists, template improvements rarely flow back cleanly into existing repositories. Teams usually end up doing one of three things:

- copying changes by hand
- regenerating and manually reconciling diffs
- accepting drift and letting each repository slowly diverge

This becomes especially painful in organizations that create many repositories over time. Every new project tends to repeat the same investment:

- wiring CI/CD
- setting up linters and formatters
- enabling spell checking
- configuring license validation
- adding copy/paste detection
- introducing complexity monitoring
- setting up git hooks
- establishing repository metadata and automation
- applying security and compliance defaults
- correctly wiring open source licensing and release flows

All of that work is necessary. None of it should have to be redone from scratch for every repository.

Moltark is built around a different model:

- **bootstrap** projects from reusable templates
- **compose** project conventions from smaller building blocks
- **update** existing projects as templates evolve
- keep generated output **reviewable**, **maintainable**, and **owned by the repository**

## Vision

Moltark aims to make project templates behave less like one-time scaffolding and more like evolving organizational building blocks.

A template should not just help start a project. It should continue to carry engineering standards, quality controls, automation, and repository conventions forward as the project grows.

In that model, project setup becomes a reusable and updatable asset for a team or organization rather than repeated manual work scattered across repositories.

## Why this matters even more in the AI era

As software delivery becomes increasingly accelerated by AI agents, strong project foundations matter more, not less.

Agent-driven development works best when repositories provide clear structure, strict feedback loops, and explicit quality boundaries. In practice, that means stronger defaults around:

- rigorous linting and formatting
- repository-wide conventions
- copy/paste and duplication detection
- complexity monitoring
- reproducible automation
- machine-readable project structure
- shared skills and guidance for agents

AI systems are fast, but they also amplify inconsistency when the surrounding project lacks guardrails. A weak repository setup does not stay weak for long under agent-driven change; it drifts faster.

That makes updatable project templates even more valuable. The right defaults can be established once, then continuously improved and propagated as tooling, standards, and workflows evolve.

This is especially important because these standards are changing quickly. In the AI era, project conventions are not static. New linters, duplication detectors, complexity watchers, repository policies, and agent workflows are emerging at a much higher rate than in traditional software environments.

Moltark is designed for that reality: not just to generate projects, but to help them keep pace with rapidly evolving engineering standards, including CI/CD-integrated maintainability checks such as linting, duplication detection, and complexity monitoring with tools like [mehen](https://github.com/ophidiarium/mehen).

## Planned capabilities

- Bootstrap new projects from reusable templates
- Compose templates from smaller modules
- Update existing projects to newer template versions
- Keep template-driven changes understandable and reviewable
- Support multiple languages and software ecosystems
- Enable gradual project evolution instead of regeneration-heavy workflows

## Why Starlark

Moltark uses **Starlark** as its template and composition language.

Starlark offers a useful balance of programmability, determinism, and composability. It is expressive enough for real project logic, while still constrained enough to remain understandable and portable.

Starlark matters to Moltark, but it is not the main point. The main point is making software project templates composable, maintainable, and updatable over time.

## Why not just use a template generator?

Traditional template generators are excellent for creating an initial repository, but they usually stop there.

After generation, the relationship between the project and the template is mostly gone. Future improvements to the template do not naturally flow into already-generated repositories.

Moltark is explicitly designed to keep that relationship alive.

## Why not Projen?

Projen is powerful, but it is built around regeneration from code as the primary model.

That works well in some ecosystems, especially when the generated surface is intentionally treated as derived output. But it becomes less satisfying when teams want:

- more explicit control over updates
- clearer reconciliation with manual edits
- reusable template modules decoupled from one monolithic project-definition model
- a solution that feels more like template evolution than full resynthesis

Moltark is not trying to be “Projen in another language.” It is exploring a different model: project templating with first-class updates.

## Why not Copier?

Copier is one of the closest conceptual neighbors to Moltark.

It already recognizes that project templates should support updates, not just initial generation. That is directionally very aligned.

The main difference is that Moltark aims to push harder on:

- **composable template modules**, not just repository templates
- **Starlark as a structured configuration and composition language**
- a model that feels closer to reusable project building blocks than one template repository per project type

## Why not Bazel macros or rules?

Bazel is excellent for build graphs, reproducibility, and structured configuration, but it is not primarily a project templating system.

Moltark borrows some aesthetic and language inspiration from that ecosystem, but it targets a different problem: bootstrapping and evolving repository structure, configuration, and conventions over time.

## Design principles

- **Bootstrap is not enough** — templates should stay useful after day one
- **Updates are first-class** — template evolution should be part of the model
- **Composition over duplication** — project conventions should be reusable and modular
- **Shared standards should be encoded once** — project setup should not be rebuilt from scratch for every repository
- **Governance should be reusable** — quality, compliance, and automation defaults should travel with templates
- **AI needs stronger guardrails, not weaker ones** — agent-driven development increases the value of consistent tooling and feedback
- **Generated output should remain understandable** — users should be able to read, review, and own their repositories
- **Projects should evolve, not be discarded and recreated**

## Early package status

The current npm and PyPI packages are stub releases published during the project's early setup.

They should not yet be considered functional releases of Moltark.

## Roadmap

Initial areas of focus:

- Go CLI
- Starlark module system
- project bootstrap workflow
- update and reconciliation engine

Initial ecosystem support:

- Go
- Python
- Rust

After that, Moltark is expected to expand into:

- Bazel, where Starlark-native and code-driven configuration are a particularly strong fit
- broader support for ecosystems with code-based configuration, starting with TypeScript and Ruby
- extensible config stubs that can evolve from generated defaults into maintainable project-owned code
  
## License

AGPL
