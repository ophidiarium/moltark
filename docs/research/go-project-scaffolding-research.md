# Go Project Scaffolding Research

> **Research document** · April 2026  
> Scope: Go project scaffolding and maintenance tooling, with emphasis on update-safe and projen-like approaches.

---

## Original brief

The notes below were copied from an external research draft and retained as source material for Moltark.

---

# Go project scaffolding and maintenance tools in 2025

**Go lacks a single dominant scaffolding tool equivalent to Cookiecutter or projen**, but a rich ecosystem of specialized tools has emerged. The most important finding: **no widely adopted projen-style tool exists for Go** that continuously regenerates project configs from a central definition — a notable ecosystem gap. The closest equivalent is SAP's `go-makefile-maker` (13 stars but deeply functional). For pure scaffolding, **go-blueprint** (8.7k stars) leads the pack, while the official `gonew` has stagnated after both of its flagship template repos were archived in mid-2025. For AI-assisted generation, **Sponge** stands out with DeepSeek R1 integration and 2.8k stars.

---

## The official `gonew` experiment has largely stalled

The Go team's `gonew` (part of `golang.org/x/tools`) launched in July 2023 as an intentionally minimal tool: it copies a Go module and rewrites its module path. That's it — no variable substitution, no interactive prompts, no conditional logic, no ongoing maintenance.

**Repository:** github.com/golang/tools (`cmd/gonew` subdirectory) · **~7,900 stars** (entire tools repo) · **Status:** still marked "highly experimental"

The tool never graduated to a `go new` subcommand in Go 1.24, 1.25, or 1.26. More concerning, **both flagship template repos were archived in July 2025**: GoogleCloudPlatform/go-templates (115 stars) and ServiceWeaver/template (29 stars). The GitHub discussion (#61669, opened by Russ Cox) has **197 upvotes and 59 participants** but shows minimal Go team engagement since launch. Community members have noted the disconnect — one September 2025 commenter observed that every example template from the official blog post is now archived, with no response from maintainers.

`gonew` handles **initial scaffolding only** — one-time copy with module path renaming. It cannot template variables, run post-generation hooks, or update projects when templates change. Several community forks add features (notably betterde's fork with interactive `template.yaml` support), but the core tool remains deliberately bare. For teams wanting more than module-path substitution, gonew is a starting point at best.

---

## go-blueprint dominates community scaffolding with 8.7k stars

**go-blueprint** (github.com/Melkeydev/go-blueprint) is the most popular purpose-built Go scaffolding tool, growing from 2.3k stars in May 2024 to **8.7k by April 2026**. It generates complete, opinionated project structures through a beautiful Charm-powered TUI or non-interactive CLI flags.

The tool supports **6 HTTP frameworks** (Chi, Gin, Fiber, Echo, HttpRouter, Gorilla/mux), **6 database drivers** (PostgreSQL, MySQL, SQLite, MongoDB, Redis, ScyllaDB), and advanced features including HTMX with Templ, React+Vite frontends, WebSocket endpoints, Tailwind CSS v4, Docker/docker-compose, and GitHub Actions CI/CD. Database selections automatically include **Testcontainers-based integration tests** — a standout feature. Generated projects include a Makefile, Air live-reload config, GoReleaser config, and `.env` file.

| Aspect | Details |
|---|---|
| **GitHub** | github.com/Melkeydev/go-blueprint |
| **Stars** | ~8,700 |
| **Latest release** | v0.10.11 (July 2025) |
| **Install** | `go install`, Homebrew, or npm |
| **Ongoing maintenance** | ❌ Initial scaffold only |
| **Limitations** | No ORM integration, no gRPC, no auth scaffolding, React-only frontend (besides HTMX), pre-v1.0 |

go-blueprint is **scaffolding-only** — it generates and exits. The ~9-month gap since the last release (July 2025) suggests the project may be entering a maintenance phase, though issues are still being filed. A companion web UI at go-blueprint.dev lets users configure projects visually before generating.

---

## The projen gap: `go-makefile-maker` is the closest equivalent

The Go ecosystem's most significant tooling gap is the absence of a projen-style tool that continuously regenerates project configuration from a central definition. Only two tools genuinely fill this role.

**sapcc/go-makefile-maker** (github.com/sapcc/go-makefile-maker) is **the closest Go equivalent to projen**. Despite only **13 stars**, it's backed by SAP's Cloud Infrastructure team with **1,357 commits** and active development. From a single `Makefile.maker.yaml`, it regenerates:

- Makefiles (build, test, lint, vendor targets)
- GitHub Actions CI workflows (lint, build, test, coverage, security scanning)
- Dockerfile and .dockerignore
- golangci-lint configuration (`.golangci.yaml`)
- GoReleaser configuration
- Renovate config for automated dependency PRs
- Nix shell.nix and .envrc for dev environments
- License headers and REUSE compliance

The workflow mirrors projen exactly: **edit the YAML, re-run the tool, generated files are overwritten**. It even supports `--autoupdate-deps` for CI-driven automation. The low star count reflects its origin as an internal SAP tool rather than any quality issue.

**Bazel Gazelle** (github.com/bazel-contrib/bazel-gazelle, **~1,400 stars**, v0.47.0 Nov 2025) fills the projen role exclusively for Bazel users. It continuously regenerates BUILD.bazel files from Go source conventions and manages external dependency declarations from `go.mod`. Run `bazel run //:gazelle` after any code change, and build files sync automatically. The complexity tax of adopting Bazel limits its applicability, but within that ecosystem, Gazelle is indispensable.

**Projen itself** has Go bindings (github.com/projen/projen-go, 5 stars) that let you write `.projenrc.go`, but it **does not include a Go project type** — it only lets you use Go to configure TypeScript/Python/Java projects. A custom Go project type would need to be built from scratch.

---

## Template tools range from Cookiecutter ports to in-project generators

Several tools occupy the template-based scaffolding space, each with different strengths.

**hay-kot/scaffold** (github.com/hay-kot/scaffold, **~134 stars**, v0.5.0 Oct 2024) is the most innovative entry. Built as a Cookiecutter alternative in Go, its killer feature is **in-project scaffolding**: a `.scaffolds` directory within existing projects generates components, controllers, and boilerplate on demand — and can **inject code into existing files**. This bridges the gap between one-time scaffolding and ongoing maintenance. It supports Go templates with Sprout helper functions, feature flags, custom delimiters, and has a polished Charm-powered TUI.

**golang-standards/project-layout** (github.com/golang-standards/project-layout, **~49k stars**) is not a tool but a controversial reference layout. Russ Cox publicly criticized it in Issue #117, stating *"this is not a standard Go project layout."* The entire `golang-standards` org was created by a single person for SEO purposes. The README now disclaims official status, and Go's own documentation recommends starting simple with just `main.go` + `go.mod`. Despite the controversy, many scaffolding tools still generate this structure.

Other notable template tools include **lacion/cookiecutter-golang** (~660 stars, Spotify maintains a fork), **SchwarzIT/go-template** (~198 stars, enterprise-oriented with Nix flake support), and **go-scaffold/go-scaffold** (~197 stars, Helm-inspired with Sprig functions). **Autostrada** (autostrada.dev) offers a web-based generator by Alex Edwards (author of "Let's Go") that produces MIT-licensed, framework-free application code.

---

## Sponge leads AI-assisted Go scaffolding with DeepSeek integration

**Sponge** (github.com/go-dev-frame/sponge, **~2,800 stars**, v1.16.1 Dec 2025) is the most significant AI-assisted Go development tool. It combines three code generation engines: built-in templates, custom templates, and **AI-assisted generation using DeepSeek R1**. The AI integration generates business logic implementations from API definitions, with contextual understanding enhanced in v1.15.0.

Sponge generates complete production services — RESTful APIs (Gin), gRPC, HTTP+gRPC hybrids, and gRPC Gateway services — from SQL, Protobuf, or JSON definitions. It integrates **30+ components** including GORM, Redis, MongoDB, Kafka, RabbitMQ, service discovery, circuit breaking, distributed tracing, and monitoring. Unlike most scaffolding tools, Sponge **handles ongoing maintenance** through incremental generation, a code merge tool, and a web UI at localhost:24631 for iterative development.

Beyond Sponge, the AI-assisted landscape for Go is thin. **goboot** (github.com/it-timo/goboot, created May 2025) is a promising but nascent deterministic scaffolding tool. General-purpose AI assistants (Cursor, Copilot, Claude) can scaffold Go projects interactively, but **no dedicated LLM-powered Go project generator** exists that understands requirements and generates complete structures from natural language. This remains an open opportunity.

---

## Ongoing maintenance requires assembling multiple tools

No single tool handles the full lifecycle of Go project maintenance. Teams must compose solutions from specialized tools:

| Tool | What it maintains | Stars | Approach |
|---|---|---|---|
| **sapcc/go-makefile-maker** | Makefile, CI, Dockerfile, lint, Renovate, Nix configs | 13 | Regenerate from YAML |
| **Bazel Gazelle** | BUILD files, dependency rules | 1,400 | Scan source, regenerate |
| **Renovate** | go.mod/go.sum, Dockerfile bases, GH Actions versions | 18k+ | Automated PRs |
| **GoReleaser** | Cross-platform builds, Docker images, changelogs | 15,700 | Execute from YAML config |
| **Task** (go-task/task) | Build/test/lint automation | 12,800+ | YAML task runner |
| **Mage** | Build automation in pure Go | 4,100 | Go functions as targets |
| **Sponge** | Service code, business logic, API definitions | 2,800 | Incremental gen + AI |
| **hay-kot/scaffold** | In-project boilerplate, component generation | 134 | Template injection |

**Renovate** (github.com/renovatebot/renovate, **18k+ stars**) deserves special mention for Go. It maintains `go.mod` dependency versions, handles major version import path migration, supports `GOPROXY`/`GOPRIVATE` for private modules, and auto-tidies vendor directories — all via automated PRs with configurable auto-merge and grouping. GitHub's built-in Dependabot offers simpler Go module updates with zero configuration.

For build automation, **Task** (taskfile.dev, **12.8k+ stars**) and **Mage** (magefile.org, **~4.1k stars**) serve complementary roles. Task uses YAML (`Taskfile.yml`) as a simpler Make alternative; Mage lets you write build targets as plain Go functions with no DSL. Neither generates configuration — they execute it. **bufbuild/makego** offers a middle ground: shared Makefile fragments that can be synced from an upstream template repo, providing a lightweight form of ongoing maintenance used across Buf's open-source projects.

---

## Conclusion

The Go scaffolding ecosystem is **fragmented but functional**, with clear specialization by tool. go-blueprint owns initial project generation with the best developer experience. `gonew` has official backing but appears to be fading. The projen-style gap remains the ecosystem's biggest unmet need — `go-makefile-maker` fills it elegantly but lacks visibility. Sponge represents the AI frontier with real production capabilities, while hay-kot/scaffold uniquely bridges bootstrap and ongoing generation.

For teams choosing today: **use go-blueprint for initial scaffolding**, **go-makefile-maker for ongoing config maintenance**, **Renovate for dependency management**, and **Task or Mage for build automation**. If you're building microservices from API definitions and want AI assistance, Sponge is worth serious evaluation. If you're on Bazel, Gazelle is non-negotiable. And if you're waiting for `gonew` to mature — the evidence suggests you should not hold your breath.
