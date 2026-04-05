core = use("moltark/core")

repo = core.project(
    id = "repo",
    kind = "workspace",
    path = ".",
)

core.json_file(
    target = repo,
    path = ".vscode/settings.json",
    values = {
        "editor.codeActionsOnSave": {
            "source.fixAll": "explicit",
        },
        "files.trimTrailingWhitespace": True,
    },
)

core.toml_file(
    target = repo,
    path = "typos.toml",
    values = {
        "default": {
            "extend-words": {
                "teh": "teh",
            },
        },
        "files": {
            "extend-exclude": [
                "vendor/**",
            ],
        },
    },
)

core.yaml_file(
    target = repo,
    path = ".github/labeler.yml",
    values = {
        "docs": {
            "changed-files": {
                "any-glob-to-any-file": [
                    "docs/**",
                    "README.md",
                ],
            },
        },
        "go": {
            "changed-files": {
                "any-glob-to-any-file": [
                    "**/*.go",
                ],
            },
        },
    },
)
