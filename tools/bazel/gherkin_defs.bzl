load("@rules_go//go:def.bzl", "go_test")

GherkinInfo = provider(
    doc = "Transitive Gherkin feature specifications",
    fields = {
        "feature_specs": "depset of .feature files contributed by this target and its dependencies",
    },
)

def _collect_feature_specs(srcs, deps):
    return depset(
        srcs,
        transitive = [dep[GherkinInfo].feature_specs for dep in deps],
    )

def _gherkin_library_impl(ctx):
    feature_specs = _collect_feature_specs(ctx.files.srcs, ctx.attr.deps)
    return [
        DefaultInfo(
            files = feature_specs,
            runfiles = ctx.runfiles(transitive_files = feature_specs),
        ),
        GherkinInfo(feature_specs = feature_specs),
    ]

gherkin_library = rule(
    implementation = _gherkin_library_impl,
    attrs = {
        "srcs": attr.label_list(
            allow_files = [".feature"],
            doc = "Feature files owned directly by this target",
        ),
        "deps": attr.label_list(
            providers = [GherkinInfo],
            doc = "Other gherkin_library targets to include transitively",
        ),
    },
    provides = [GherkinInfo],
)

def godog_feature_test(name, srcs, features, deps = None, data = None, **kwargs):
    """Wrap a Go godog test with first-class Gherkin feature runfiles.

    Args:
      name: target name.
      srcs: Go test sources.
      features: list of gherkin_library targets.
      deps: Go deps for the test binary.
      data: additional runtime data files.
      **kwargs: forwarded to go_test.
    """
    if deps == None:
        deps = []
    if data == None:
        data = []

    go_test(
        name = name,
        srcs = srcs,
        deps = deps,
        data = data + features,
        **kwargs
    )
