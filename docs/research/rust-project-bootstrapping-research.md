# Rust Project Bootstrapping Research

> **Research document** · April 2026  
> Scope: Rust project scaffolding, configuration management, release lifecycle tooling, and template drift or update propagation.

---

## Table of Contents

1. [Bootstrapping — Initial Project Creation](#1-bootstrapping--initial-project-creation)
2. [Framework-Specific Scaffolders](#2-framework-specific-scaffolders)
3. [Configuration Management](#3-configuration-management)
4. [Release Lifecycle](#4-release-lifecycle)
5. [Template Update / Drift Management](#5-template-update--drift-management)
6. [Summary Table](#6-summary-table)

---

## 1. Bootstrapping — Initial Project Creation

### cargo-generate ⭐ 2.4k

> **The de facto standard.** Think Cookiecutter-for-Rust.

- **Repo:** <https://github.com/cargo-generate/cargo-generate>
- **Docs:** <https://cargo-generate.github.io/cargo-generate/>
- **Template language:** Shopify Liquid + Rhai hook scripts + regex placeholders
- **Template registry:** 179+ community templates tagged [`cargo-generate`](https://github.com/topics/cargo-generate) on GitHub, covering ESP-IDF embedded, Ratatui TUI, eBPF via Aya, WASM, Bevy, and more
- **IDE integration:** RustRover integrates it natively

```bash
cargo install cargo-generate
cargo generate gh:rust-github/template          # GitHub shorthand
cargo generate gl:user/template                 # GitLab
cargo generate bb:user/template                 # Bitbucket
cargo generate --git https://github.com/user/template.git --name my-project
```

**Key features:**
- Liquid templating with custom filters (title-case, snake-case, etc.)
- Rhai scripting for conditional logic in hooks (pre/post generation)
- `cargo-generate.toml` config: include/exclude globs, prompt definitions, sub-templates
- `--init` flag: generate into the current directory (does **not** overwrite existing files)
- Favorites system for storing frequently used template URLs
- Official GitHub Action for CI-testing your templates

**Limitation:** Does **not** track template version after generation. Re-running into an existing project will error on any conflict. No drift detection.

---

### cargo-scaffold

> **The Handlebars/TOML alternative.** Language-agnostic, no code required to define a template.

- **Repo:** <https://github.com/iomentum/cargo-scaffold>
- **Docs:** <https://docs.rs/cargo-scaffold>
- **Template language:** Handlebars + `.scaffold.toml` config

```bash
cargo install cargo-scaffold
cargo scaffold https://github.com/username/template.git
cargo scaffold your_local_template_dir
cargo scaffold https://github.com/user/template.git -t v1.2.0   # pin to tag/commit
```

**`.scaffold.toml` example:**

```toml
[template]
name = "my-service"
author = "Jane Doe"
exclude = ["./target"]
notes = "Have fun with {{name}}!"

[hooks]
pre  = ["bash -c pre_script.sh"]
post = ["cargo vendor"]

[parameters.feature]
type    = "string"
message = "Feature name?"
required = true

[parameters.api_style]
type   = "select"
message = "Which API style?"
values  = ["REST", "graphql", "grpc"]
```

**Parameter types:** `string`, `integer`, `float`, `boolean`, `select`, `multiselect`

**Limitation:** Same fire-and-forget model as cargo-generate. Nothing persists to enable future syncing.

---

### cargo-quickstart

> **New (2025), opinionated fast bootstrapper.**

- **Repo:** <https://github.com/sm-moshi/cargo-quickstart>
- **Crate:** <https://lib.rs/crates/quickstart-lib>

Billed as "a blazing fast and opinionated cargo subcommand to bootstrap modern Rust projects with confidence and speed." Scaffolds a full project with Git initialized, best practices configured, and documentation templates ready to go. Early-stage — caveat emptor on maturity.

```bash
git clone https://github.com/sm-moshi/cargo-quickstart
cd cargo-quickstart
cargo install --path crates/quickstart-cli
cargo quickstart new my-project
```

---

## 2. Framework-Specific Scaffolders

These tools are closer to `rails new` or `nest new` — they scaffold a project **and** support ongoing code generation as the project grows.

### Loco — Rails/Django-style Rust framework

- **Site:** <https://loco.rs/>
- **Repo:** <https://github.com/loco-rs/loco>

Includes a built-in ORM, generators, templating, and project scaffolding. Emphasizes convention over configuration.

```bash
cargo install loco-cli
loco new my-app                      # interactive: choose SaaS / REST API / etc.

# Ongoing code generation
cargo loco generate model user name:string email:string
cargo loco generate controller posts
cargo loco generate deployment       # generates Dockerfile or Shuttle config
```

**Limitation for template sync:** `cargo loco generate` is additive codegen (new files only), not template drift correction. No concept of a tracked template version.

---

### create-rust-app — Rust + React fullstack

- **Repo:** <https://github.com/Wulf/create-rust-app>
- **Crate:** <https://lib.rs/crates/create-rust-app>

Generates TypeScript types for Rust code via `#[tsync::tsync]`, scaffolds DB models, endpoints, service files, and typed react-query hooks from a single command.

```bash
cargo install create-rust-app
create-rust-app my-todo-app
cd my-todo-app
create-rust-app   # interactive: generate resource, model, hooks, etc.
```

---

### cargo-leptos — Leptos full-stack WASM

- **Repo:** <https://github.com/leptos-rs/cargo-leptos>

```bash
cargo install cargo-leptos --locked
cargo leptos new --git https://github.com/leptos-rs/start-axum
cargo leptos watch    # dev server with HMR
cargo leptos build --release
```

---

## 3. Configuration Management

### cargo-wizard

> **The closest Rust analog to projen's config management layer.**

- **Repo:** <https://github.com/Kobzol/cargo-wizard>
- **Blog post:** <https://kobzol.github.io/rust/cargo/2024/03/10/rust-cargo-wizard.html>

Simplifies configuration of Cargo projects for maximum runtime performance, fastest compilation time, or minimal binary size. Applies changes to `Cargo.toml` and `config.toml` interactively or via pre-built templates. Featured as the **"plugin of the cycle"** by the Cargo team in the Rust 1.92 development blog.

```bash
cargo install cargo-wizard

# Interactive TUI
cargo wizard

# Direct application
cargo wizard apply fast-runtime dist        # apply fast-runtime profile to 'dist'
cargo wizard apply fast-compile dev         # optimize dev builds for compile speed
cargo wizard apply min-size release         # minimize binary size for release

# Nightly extras (LTO, PGO, etc.)
cargo +nightly wizard
```

**Built-in templates:**
| Template | Goal |
|---|---|
| `fast-runtime` | Maximum execution speed |
| `fast-compile` | Shortest compile times |
| `min-size` | Smallest binary output |

---

## 4. Release Lifecycle

### release-plz — Automated release management

- **Site:** <https://release-plz.dev/>
- **Repo:** <https://github.com/MarcoIeni/release-plz>
- **Crate:** <https://crates.io/crates/release-plz>

The gold standard for automated Rust releases. Bumps versions per Semantic Versioning based on Conventional Commits and API breaking changes detected by `cargo-semver-checks`. Opens a "Release PR" with updated `CHANGELOG.md`, `Cargo.toml`, and `Cargo.lock`. When merged: creates the git tag, publishes to crates.io, and creates a GitHub/GitLab/Gitea release.

```bash
cargo install --locked release-plz

# Local workflow
release-plz update           # bump Cargo.toml + CHANGELOG locally
release-plz release-pr       # open a Release PR
release-plz release          # tag + publish (after merging the PR)
```

**Typical CI setup (`.github/workflows/release-plz.yml`):**

```yaml
on:
  push:
    branches: [main]
jobs:
  release-plz:
    runs-on: ubuntu-latest
    steps:
      - uses: MarcoIeni/release-plz-action@v0.5
        env:
          GITHUB_TOKEN: ${{ secrets.RELEASE_PLZ_TOKEN }}
          CARGO_REGISTRY_TOKEN: ${{ secrets.CARGO_REGISTRY_TOKEN }}
```

---

### cargo-dist — Binary distribution + self-generating CI

- **Repo:** <https://github.com/axodotdev/cargo-dist>
- **Docs:** <https://opensource.axo.dev/cargo-dist/>

Covers the full distribution pipeline: planning, building binaries and installers, hosting artifacts, publishing packages, and announcing releases. Its biggest superpower: **it generates its own CI scripts**.

```bash
cargo install cargo-dist

# First-time setup — writes .github/workflows and dist config to Cargo.toml
cargo dist init

# Check what would be built
cargo dist plan

# Build locally
cargo dist build

# Cross-compilation support (v0.26+)
# Linux: via cargo-zigbuild
# Windows: via cargo-xwin
```

**Recent additions (2024–2025):**
- Cross-compilation on Linux/Windows (cargo-zigbuild, cargo-xwin)
- `cargo-auditable` integration (embeds dependency tree in binaries)
- CycloneDX SBOM generation via `cargo-cyclonedx`
- Now also works as standalone `dist` CLI (no cargo prefix required)
- Supports non-Rust projects via `dist.toml`

---

### oranda — Auto-generated project landing pages

- **Repo:** <https://github.com/axodotdev/oranda>
- **Site:** <https://axodotdev.github.io/oranda/>

Opinionated static site generator designed for developer tool projects. Zero-config: run `oranda build` and get a website pulled from your `README.md`, `CHANGELOG.md`, and cargo-dist release artifacts.

```bash
cargo install oranda
oranda build     # generate site
oranda dev       # build + watch + local server
```

---

## 5. Template Update / Drift Management

> **TL;DR: None of the Rust-native tools support this. It is an open gap in the ecosystem.**

### What "template update" means

The feature in question — tracking a template version and propagating upstream changes to already-bootstrapped projects — requires:

1. **A lockfile** recording the template commit hash at generation time
2. **Diff computation:** old template@commit → new template@HEAD
3. **3-way merge:** apply that diff onto your project, respecting local changes
4. **Conflict resolution:** surface merge conflicts the same way `git merge` would

This is what Python's **[cruft](https://cruft.github.io/cruft/)** provides for Cookiecutter templates. It stores a `.cruft.json` tracking the template commit, and `cruft update` computes and applies the diff. `cruft check` can be added to CI to detect drift.

### Per-tool status

| Tool | Template update support | Notes |
|---|---|---|
| `cargo-generate` | ❌ None | Explicitly refuses to overwrite existing files. Issue [#291](https://github.com/cargo-generate/cargo-generate/issues/291) (partial injection) is open but not merged. No lockfile, no diff. |
| `cargo-scaffold` | ❌ None | Fire-and-forget, same model. |
| `cargo-quickstart` | ❌ None | Too new; no such feature planned. |
| `cargo-wizard` | ❌ Different category | Manages Cargo config profiles, not templates. |
| `Loco generate` | ⚠️ Additive only | Can add new files to existing projects, but no drift tracking. |
| `cargo-dist init` | ⚠️ Narrow case | Re-running regenerates its own CI YAML files. Only manages cargo-dist's own output, not your full project template. |
| `release-plz` | ❌ Different category | Manages release lifecycle, not project structure. |

### Workarounds available today

#### Option A: Use cruft with a Cookiecutter template (recommended)

cruft is language-agnostic. Write a Cookiecutter template that generates your Rust project structure (CI YAMLs, `rustfmt.toml`, `deny.toml`, `Cargo.toml` skeleton, `CONTRIBUTING.md`, etc.) and use cruft for the full lifecycle.

```bash
pip install cruft

# Create
cruft create https://github.com/your-org/cookiecutter-rust-service

# Check drift (good for CI)
cruft check

# Apply upstream template changes to your project (3-way merge)
cruft update

# See what changed vs template
cruft diff

# Link an existing project created with cookiecutter directly
cruft link https://github.com/your-org/cookiecutter-rust-service
```

**`.cruft.json`** (auto-generated, commit this):

```json
{
  "template": "https://github.com/your-org/cookiecutter-rust-service",
  "commit": "a3f2c1b...",
  "context": {
    "cookiecutter": {
      "project_name": "my-service",
      "crate_name": "my_service"
    }
  },
  "skip": ["src/", "tests/"]
}
```

**CI drift detection:**

```yaml
- name: Check template drift
  run: |
    pip install cruft
    cruft check
```

#### Option B: `actions-template-sync` GitHub Action

For the CI/config layer specifically. Opens a PR whenever the source template repository diverges from your project. Works at the Git-repo level rather than file-by-file diff.

```yaml
# .github/workflows/template-sync.yml
on:
  schedule:
    - cron: "0 2 * * 1"   # Every Monday at 2am
  workflow_dispatch:

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: AndreasAugustin/actions-template-sync@v2
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          source_gh_slug: your-org/rust-project-template
```

#### Option C: cargo-dist's self-regeneration (narrow case)

If your main concern is keeping CI pipelines up-to-date, re-running `cargo dist init` updates cargo-dist's own generated workflow files. Limited to cargo-dist's output only.

#### Option D: projen-style ownership (manual discipline)

Design your cargo-generate template so that "managed" files (CI YAMLs, `rustfmt.toml`, `clippy.toml`, `deny.toml`, `.editorconfig`) are explicitly tracked and re-runnable via `cargo generate --init --force` into an existing directory. User code is protected via `.genignore`. No automated diff — requires developer discipline to re-run after template changes.

---

## 6. Summary Table

| Tool | Role | Template update | Analog |
|---|---|---|---|
| `cargo-generate` | Scaffolding from Git templates | ❌ | Cookiecutter |
| `cargo-scaffold` | Scaffolding via TOML + Handlebars | ❌ | Cookiecutter (alt) |
| `cargo-quickstart` | Opinionated bootstrapper (2025) | ❌ | create-next-app |
| `loco new` + `loco generate` | Full-stack app + ongoing codegen | ⚠️ additive only | Rails / projen |
| `cargo-wizard` | Cargo config optimization | ❌ (different category) | projen config mgmt |
| `release-plz` | Semver + changelog + crates.io publish | ❌ (different category) | semantic-release |
| `cargo-dist` | Binary distribution + self-generated CI | ⚠️ own files only | GoReleaser + projen |
| `oranda` | Auto project website | ❌ (different category) | — |
| `cruft` (Python) | **Full template lifecycle with drift sync** | ✅ | cruft (works for Rust too) |

---

## Recommended Stack (2025)

For a general-purpose Rust CLI or library targeting crates.io:

```
cargo-generate          →  bootstrap
cargo-wizard            →  configure profiles
release-plz + cargo-dist →  release & distribute
cruft (if drift matters)  →  ongoing template sync
```

For a Rust web service (internal or public):

```
loco new                →  bootstrap + ongoing codegen
cargo-wizard            →  configure profiles
release-plz             →  release management
cruft                   →  keep CI/config layer in sync with org template
```

---

*Sources: GitHub repositories, official documentation, crates.io, and lib.rs. Verified April 2026.*
