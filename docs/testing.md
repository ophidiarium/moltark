# Testing

Moltark uses a fixture-driven testing approach with two main styles:

- snapshot-heavy integration tests
- Gherkin feature scenarios

## Current Test Stack

The current test dependencies are:

- [`go-snaps`](https://github.com/gkampitakis/go-snaps) for snapshot assertions
- [`godog`](https://github.com/cucumber/godog) for Gherkin scenarios

## Directory Layout

- `tests/fixtures/`
  real repository fixtures copied to temp directories during tests
- `tests/integration/`
  end-to-end CLI tests with snapshot assertions
- `tests/integration/__snapshots__/integration_test.snap`
  snapshot output for integration tests
- `tests/features/`
  Gherkin feature files and Godog step bindings

## How Integration Tests Work

The integration tests:

1. copy a fixture repository into a temp dir
2. run `moltark` commands against that temp dir
3. snapshot CLI output and resulting repo state

Helpers live in:

- `internal/testutil/testutil.go`

Snapshot cleanup is handled in:

- `tests/integration/main_test.go`

## How Feature Tests Work

The feature tests use:

- `tests/features/uv_dependencies.feature`
- `tests/features/uv_dependencies_test.go`

Under Bazel, feature files are modeled explicitly through local rules in [`tools/bazel/gherkin_defs.bzl`](../tools/bazel/gherkin_defs.bzl):

- `gherkin_library(...)` makes `.feature` files first-class Bazel inputs with transitive propagation
- `godog_feature_test(...)` wraps `go_test` and adds those features to runfiles automatically
- `godog` remains the actual executor, so feature behavior stays identical between `go test` and `bazelisk test`

Use Gherkin when the scenario reads more clearly as repository behavior than as a raw CLI transcript.

## Commands

Run the full suite:

```bash
go test -count=1 ./...
```

Run the Bazel suite:

```bash
bazelisk test //...
```

Regenerate `BUILD.bazel` files with Gazelle:

```bash
bazelisk run //:gazelle
```

Run only integration tests:

```bash
go test -count=1 ./tests/integration/...
```

Run only feature tests:

```bash
go test -count=1 ./tests/features/...
```

Refresh integration snapshots:

```bash
UPDATE_SNAPS=true go test -count=1 ./tests/integration/...
```

`-count=1` is recommended when working on snapshot-heavy behavior because Go test caching can otherwise make failures look stale after snapshot refreshes.

For Bazel-driven development:

- fixture, snapshot, and `.feature` files are intentionally wired through explicit rule inputs or `data` attributes, so keep those manual entries intact when rerunning Gazelle
- use `bazelisk`, not an arbitrary global `bazel`, so the workspace version pinned in [`.bazelversion`](../.bazelversion) stays authoritative

## What To Test

When behavior changes, prefer adding or updating repository-level scenarios that cover:

- bootstrap
- re-apply with no changes
- template upgrades
- drift detection
- conflict surfacing
- provider and capability resolution
- preservation of user-managed surfaces
- user-visible diagnostics

If a change is architectural but not yet executable, `moltark show` snapshots are often the right first test surface.
