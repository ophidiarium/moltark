core = use("moltark/core")

app = core.project(
    id = "app",
    kind = "go_project",
    path = ".",
    attributes = {
        "go_version": "1.24",
    },
)

core.bootstrap_requirement(
    tool = "golangci-lint",
    target = app,
    purpose = "Go linting",
    strategies = ["brew", "mise"],
)

core.fact(
    name = "moltark.language.go",
    target = app,
    values = {
        "version": "1.24",
    },
)

core.task(
    name = "lint",
    target = app,
    command = ["golangci-lint", "run", "./..."],
    runtime = "go",
    tags = ["lint", "go"],
)

core.trigger_binding(
    trigger = "pre-commit",
    target = app,
    match_names = ["lint"],
)

core.json_file(
    target = app,
    path = ".vscode/settings.json",
    values = {
        "editor.codeActionsOnSave": {
            "source.fixAll": "explicit",
        },
        "go.formatTool": "gofumpt",
        "go.lintOnSave": "package",
        "go.lintTool": "golangci-lint",
    },
)

core.json_file(
    target = app,
    path = ".vscode/extensions.json",
    values = {
        "recommendations": [
            "golang.Go",
        ],
    },
)

core.toml_file(
    target = app,
    path = ".golangci.toml",
    values = {
        "run": {
            "go": core.fact_value(
                target = app,
                name = "moltark.language.go",
                path = "version",
            ),
        },
    },
)
