core = use("moltark/core")
python = use("moltark/python")

repo = core.project(
    id = "repo",
    kind = "workspace",
    path = ".",
)

python.python_project(
    id = "backend",
    parent = repo,
    name = "backend",
    path = "backend",
    version = "0.1.0",
    requires_python = ">=3.12",
)
