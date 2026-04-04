# Bazel Gazelle Research

> **Research document** · April 2026  
> Scope: Gazelle as a Bazel-native reconciler, extension system, and source of bootstrap/update patterns for Moltark.

---

## Why Gazelle matters for Moltark

Gazelle is not a generic project template tool. It is a highly opinionated, update-safe reconciler for Bazel metadata. That makes it relevant to Moltark for the same reason OpenTofu is relevant: the strongest lessons are in lifecycle design, extension boundaries, ownership handling, and safe repeated application.

The main takeaway is simple:

- Gazelle keeps the CLI thin.
- Gazelle makes the reconciliation pipeline explicit.
- Gazelle lets language extensions own semantics.
- Gazelle uses structure-aware edits and merge rules instead of whole-file regeneration.

That is close to the shape Moltark wants.

## Sources

- Local clone of [`bazel-contrib/bazel-gazelle`](https://github.com/bazel-contrib/bazel-gazelle), inspected at commit `7c54844`.
- Gazelle repo docs:
  - [`README.md`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/README.md)
  - [`how-gazelle-works.md`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/how-gazelle-works.md)
  - [`gazelle-reference.md`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/gazelle-reference.md)
  - [`extend.md`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/extend.md)
  - [`def.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/def.bzl)
  - [`internal/gazelle_binary.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/internal/gazelle_binary.bzl)
  - [`internal/go_repository.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/internal/go_repository.bzl)
  - [`cmd/gazelle/main.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/cmd/gazelle/main.go)
  - [`cmd/gazelle/update-repos.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/cmd/gazelle/update-repos.go)
  - [`v2/cmd/gazelle/update/update.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/v2/cmd/gazelle/update/update.go)
- Real-world usage examples:
  - [`kythe/kythe/tools/gazelle/BUILD`](https://github.com/kythe/kythe/blob/954bc791a8f66588055ae2741e7f091fa8c5e91b/tools/gazelle/BUILD)
  - [`buildbuddy-io/buildbuddy/BUILD`](https://github.com/buildbuddy-io/buildbuddy/blob/847e5c5d4456a3a026c447002d4e658cfb823159/BUILD)
  - [`GoogleContainerTools/distroless/MODULE.bazel`](https://github.com/GoogleContainerTools/distroless/blob/accc36a5c1af3ce5bf7d7b1d75f293a4798c4c03/MODULE.bazel)
  - [`pingcap/tidb/WORKSPACE`](https://github.com/pingcap/tidb/blob/3cebd11818a26bf563f2ef78de480e545aa7d4b6/WORKSPACE)
  - [`github/codeql/go/extractor/BUILD.bazel`](https://github.com/github/codeql/blob/fb8b5699f28bbf0c3dc2bdee5ea59e6c180336fc/go/extractor/BUILD.bazel)

## Executive summary

Gazelle is best understood as a repository metadata reconciler with an extension API, not as a fancy CLI and not as a broad scaffolding engine.

The most transferable patterns for Moltark are:

- keep the CLI austere and predictable
- separate driver lifecycle from module semantics
- treat update safety as the product, not as a secondary concern
- use structured merge behavior with explicit user escape hatches
- make extension ordering and contribution boundaries explicit
- support both safe update flows and more invasive fix-up flows
- support repo-root configuration plus subtree-local overrides

The least transferable pattern is Gazelle's compiled-in extension model. That fits Bazel well, but Moltark wants shareable remote modules. Moltark should copy Gazelle's explicit extension lifecycle, not its binary assembly constraint.

## How Gazelle's CLI is implemented

### Dependencies and surface area

Gazelle's CLI is deliberately small.

- The top-level command in [`cmd/gazelle/main.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/cmd/gazelle/main.go) uses standard library `flag`, `fmt`, `log`, `os`, and `context`.
- There is no Cobra, no `urfave/cli`, no `bubbletea`, no `lipgloss`, no prompt library, and no terminal styling dependency.
- Help text is hand-written and printed directly to stderr.
- Logging is standard library `log` with a fixed `gazelle:` prefix.

This is notable because Gazelle is successful without any terminal-product theatrics. The complexity lives in the engine and the extension system.

### Command shape

The main command surface is small:

- `update`
- `fix`
- `update-repos`
- `help`

The default command is `update`.

That command vocabulary is instructive:

- `update` is the common safe path
- `fix` is a more invasive migration path
- `update-repos` is a separate dependency/bootstrap path

Moltark should pay attention to that split. It is cleaner than one command that mixes safe reconciliation, migrations, and dependency ingestion together.

### Execution model

Gazelle's `v2` driver in [`v2/cmd/gazelle/update/update.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/v2/cmd/gazelle/update/update.go) is a generic orchestration layer around:

- config loading
- walking the repo
- per-directory generation
- indexing
- dependency resolution
- merge
- output emission

It supports `-mode=fix|print|diff`, recursive or targeted subtree operation, lazy indexing, and profiling flags. Again, the important part is not the flags themselves. The important part is that the lifecycle is explicit and inspectable.

### Bazel-native invocation

Gazelle is primarily designed to be run through Bazel:

- `bazel run //:gazelle`

The Starlark side in [`def.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/def.bzl) wraps the binary in Bazel-friendly runner rules and shell scripts, sets up runfiles, passes environment, and integrates with Bazel toolchains.

This is another useful lesson: the natural invocation surface for a reconciler may be the host ecosystem, not just a standalone binary. Moltark should keep its standalone CLI, but it is worth leaving room for host-native entrypoints later.

## What Gazelle actually bootstraps

Gazelle is narrower than Projen or Copier.

It does not scaffold whole products. It bootstraps and reconciles Bazel metadata.

In practice, it can bootstrap or update:

- root `BUILD` or `BUILD.bazel` wiring so the repo has a runnable `gazelle` target
- package-level `BUILD` files for supported languages
- dependency labels in generated rules
- repository rules in `WORKSPACE` through `update-repos`
- Bzlmod external dependency declarations via the `go_deps` extension
- external repositories fetched via `go_repository`, including generation for fetched Go dependencies

This is the right interpretation for Moltark:

- Gazelle is "bootstrap" in the sense of repository reconciliation bootstrapping
- not "bootstrap" in the sense of app skeleton generation

That distinction matters because Moltark should likely support both:

- initial repo bootstrap
- repeated structural reconciliation

Gazelle shows how much value you can get from the second category alone.

## Nice use of Starlark and Bazel extension patterns

### 1. `gazelle_binary` composes extensions explicitly

[`internal/gazelle_binary.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/internal/gazelle_binary.bzl) is one of the best examples in the repo.

It takes a list of language extension libraries, generates a tiny Go source file with imports and `NewLanguage()` calls, and compiles a custom Gazelle binary around that set.

That gives Gazelle:

- explicit extension inclusion
- deterministic extension ordering
- no ambient plugin discovery
- a clean language boundary between generic driver and language semantics

Moltark should copy the explicitness, not necessarily the exact mechanism. For Moltark, the analogue is:

- explicit module imports in `Moltarkfile`
- explicit capability/fact contribution contracts
- deterministic module resolution and ordering

### 2. Directives give repo-local control without new global config files

Gazelle leans heavily on directives such as:

- `# gazelle:exclude`
- `# gazelle:map_kind`
- `# gazelle:resolve`
- `# gazelle:repository_macro`

These are powerful because they travel with the file graph they affect.

This is not a direct fit for Moltark, because Moltark already has `Moltarkfile` and state. But the broader pattern is transferable:

- local override points are valuable
- the closer the override is to the surface it affects, the more understandable the system is

For Moltark, the likely equivalents are:

- local ownership markers
- local override blocks
- component-local facts/providers
- explicit state about managed surfaces

### 3. `update-repos` treats dependency ingestion as a separate concern

[`cmd/gazelle/update-repos.go`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/cmd/gazelle/update-repos.go) is useful because it does not pretend dependency ingestion is the same as rule generation.

It has its own command, flags, importers, and merge behavior. It can write directly to `WORKSPACE` or into a macro and then wire that macro back into the workspace.

That pattern is strong. Moltark should probably treat some future flows similarly:

- dependency sync or import
- template adoption
- structural migrations

Those should not all be hidden inside one `apply`.

### 4. `go_repository` is bootstrapping plus reconciliation for external repos

[`internal/go_repository.bzl`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/internal/go_repository.bzl) is a powerful example of "bootstrap only what is missing, otherwise reconcile".

It fetches dependencies, optionally generates build files for them, respects multiple acquisition strategies, and separates fetching from BUILD generation.

That is relevant to Moltark's future module story:

- bootstrap source acquisition and template application are separate concerns
- fetched modules may still need local reconciliation rules
- reproducibility and caching matter

## Real-world usage patterns seen on GitHub

### Kythe: custom Gazelle binary

[`kythe/kythe/tools/gazelle/BUILD`](https://github.com/kythe/kythe/blob/954bc791a8f66588055ae2741e7f091fa8c5e91b/tools/gazelle/BUILD) defines a custom `gazelle_binary` with:

- built-in proto support
- a repo-local Go language extension

That validates the custom-binary pattern in real use, not just in Gazelle docs. Large repos do specialize Gazelle instead of accepting only the default bundle.

### BuildBuddy: multi-language custom binary plus deep directives

[`buildbuddy-io/buildbuddy/BUILD`](https://github.com/buildbuddy-io/buildbuddy/blob/847e5c5d4456a3a026c447002d4e658cfb823159/BUILD) is a high-signal example.

It uses:

- `gazelle_binary` with `DEFAULT_LANGUAGES`
- an additional Python Gazelle extension
- a repo-local TypeScript extension
- many repo-root directives:
  - `exclude`
  - `build_file_name`
  - `prefix`
  - `proto disable`
  - `map_kind`
  - multiple `resolve` overrides

This shows Gazelle working as a real reconciler in a polyglot monorepo, not just a small Go project helper.

The key Moltark lesson is that large repos need:

- a core engine
- selective language/module composition
- coarse repo-root defaults
- sharp local overrides

### Distroless: Bzlmod `go_deps` in a broader polyglot repo

[`GoogleContainerTools/distroless/MODULE.bazel`](https://github.com/GoogleContainerTools/distroless/blob/accc36a5c1af3ce5bf7d7b1d75f293a4798c4c03/MODULE.bazel) uses:

- `go_deps = use_extension("@gazelle//:extensions.bzl", "go_deps")`
- module declarations for Go dependencies
- alongside Python, OCI, Node, and other extension ecosystems

This is a strong precedent for Moltark's module direction:

- one root module file
- many ecosystem-specific extensions
- explicit extension usage
- repo-wide composition rather than one archetype template

### PingCAP TiDB: repository macro indirection

[`pingcap/tidb/WORKSPACE`](https://github.com/pingcap/tidb/blob/3cebd11818a26bf563f2ef78de480e545aa7d4b6/WORKSPACE) uses:

- `# gazelle:repository_macro DEPS.bzl%go_deps`

This is a subtle but important pattern. Gazelle does not require all generated dependency state to live inline in the root workspace file. It can point to structured indirection.

Moltark should remember this when choosing where managed state lands:

- some generated or managed material belongs in its own module file or macro-like surface
- root files should orchestrate, not necessarily contain everything

### CodeQL: `map_kind` to custom macros

[`github/codeql/go/extractor/BUILD.bazel`](https://github.com/github/codeql/blob/fb8b5699f28bbf0c3dc2bdee5ea59e6c180336fc/go/extractor/BUILD.bazel) contains:

- `# gazelle:map_kind go_binary codeql_go_binary //go:rules.bzl`

This is one of the clearest examples of adapting a generic generator to repo-local abstractions without forking the whole tool.

Moltark should preserve this kind of escape hatch:

- generic module behavior
- with local remapping or wrapping points
- without forcing hard forks of the whole system

## The best bootstrapping patterns Moltark should learn from Gazelle

### 1. Treat the lifecycle as the product

Gazelle's strongest document is arguably [`how-gazelle-works.md`](https://github.com/bazel-contrib/bazel-gazelle/blob/master/how-gazelle-works.md), because it explains the lifecycle clearly:

1. Load
2. Generate
3. Resolve
4. Write

For Moltark, the analogous lifecycle should stay explicit:

1. Evaluate modules and `Moltarkfile`
2. Build desired project/component model
3. Inspect current repo state
4. Resolve providers, facts, and ownership surfaces
5. Compute plan and classify changes
6. Apply deterministic edits
7. Persist updated state

That lifecycle should be visible in code, in docs, and eventually in diagnostics.

### 2. Safe update and invasive fix-up should be different modes

Gazelle keeps `update` and `fix` separate.

That is a strong precedent for Moltark. Even if Moltark does not expose a literal `fix` command soon, it should preserve the architectural distinction between:

- safe reconciliation
- migrations or transformations that may rename, delete, or normalize user-visible state

That separation improves trust.

### 3. Use structure-aware merge rules, not "render file and overwrite"

Gazelle works at the syntax-tree level and then applies merge rules that distinguish:

- mergeable fields
- non-mergeable fields
- explicit `# keep` user protection

This is directly aligned with Moltark's ownership model. The lesson is not just "use parsers". The lesson is:

- define ownership semantics per surface
- encode which attributes are tool-managed versus user-managed
- preserve comments and local edits where possible
- make user escape hatches explicit

Moltark should continue moving toward that model across TOML, JSON, YAML, and later code-based config surfaces.

### 4. Keep the generic driver separate from ecosystem logic

Gazelle's driver does walking, config, merge orchestration, and emission. Languages own:

- directives
- rule generation
- import resolution
- repo updater logic

This maps well to Moltark's desired split:

- core owns evaluation, planning, ownership, state, and apply orchestration
- modules own ecosystem semantics, providers, facts, and surfaced resources

This is the most important structural lesson from Gazelle.

### 5. Support partial and lazy work in large repos

Gazelle's lazy indexing and directory-targeted operation are important for monorepos.

Moltark should plan for equivalent capabilities:

- subtree-targeted planning
- component-targeted apply
- lazy provider/fact resolution when full-repo scans are too expensive
- diagnostics that make it clear what part of the repo was considered

This matters if Moltark wants to be a Bazel companion in real monorepos.

### 6. Root configuration plus subtree overrides is a winning pattern

Gazelle directives inherit downward through the repo tree. Large repos use that aggressively.

Moltark already has a project tree and parent-relative paths. Gazelle reinforces that this is the right direction:

- repo root sets defaults
- subtree or component scope can override
- ownership and behavior are best understood in a path-aware hierarchy

### 7. Generated state can live behind indirection

Gazelle's `repository_macro` support shows a valuable pattern:

- root files can reference generated or managed sub-files
- not every managed surface should be inlined into the root entrypoint

Moltark should preserve room for:

- root orchestration files
- delegated generated files
- module-specific managed subtrees

That will matter more as modules become shareable and monorepo usage grows.

## Where Gazelle does not map cleanly to Moltark

### 1. Compiled-in extensions are a Bazel fit, not necessarily a Moltark fit

Gazelle cannot dynamically load extensions at runtime, so it assembles a binary containing the needed languages.

That is reasonable in Bazel. It is not the right default for Moltark's intended remote module ecosystem.

Moltark should copy:

- explicit extension contracts
- deterministic module selection
- ordered composition

But Moltark should not copy:

- "rebuild the binary to add module behavior" as the primary extension story

### 2. Comment directives are a good fit for BUILD files, but not a universal answer

Gazelle benefits from BUILD files being both configuration and generated target surfaces.

Moltark manages many different file kinds. It should not overfit to comment directives everywhere. Its stronger primitives are:

- explicit state
- ownership contracts
- module/component declarations
- structured file primitives

Comment directives may still be useful in some future code-managed files, but they should not become the universal control plane.

### 3. Gazelle is narrower than Moltark's ambition

Gazelle does not try to manage:

- repo-wide developer workflows in the Projen sense
- CI, hooks, package managers, editor config, policy bundles as first-class cross-cutting components
- long-term multi-surface template evolution across many config types

Moltark wants all of that. So Gazelle is a source of patterns, not a whole blueprint.

## Concrete recommendations for Moltark

### Near-term

- Keep the CLI thin. Resist adding rich terminal UI until the engine and module model are much stronger.
- Keep the reconciliation lifecycle explicit in code and docs.
- Continue building first-class structured file primitives and ownership-aware merge semantics.
- Preserve the distinction between safe reconciliation and more invasive migration/fix flows.
- Add targeted/subtree plan and apply capabilities early if monorepo scale is a real goal.

### Module system

- Copy Gazelle's explicit extension lifecycle, but keep Moltark modules runtime-loadable.
- Keep provider, fact, and task contracts narrow and typed.
- Preserve explicit ordering where module interactions depend on it.
- Allow repo-local wrapping or remapping rather than forcing full module forks.

### Bootstrap and dependency flows

- Consider a dedicated dependency-import or module-sync flow analogous to `update-repos`.
- Keep "bootstrap project structure" distinct from "sync external dependencies" and "reconcile managed config".
- Leave room for generated indirection files rather than forcing all managed state into one root file.

## Moltark-specific interpretation

The most valuable Gazelle lesson is not "generate files from conventions." It is:

- model the system as repeatable reconciliation
- make extension/module semantics explicit
- define ownership and merge boundaries precisely
- keep the default user experience boring and trustworthy

If Moltark adopts that discipline while keeping a broader module ecosystem than Gazelle, it can end up closer to "Gazelle's update trustworthiness plus Projen/Copier's composability" than either class of tool on its own.
