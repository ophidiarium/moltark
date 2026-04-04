python = use("moltark/python")
uv = use("astral/uv")

root = python.python_project(
    name = "tally",
    path = "tally",
    version = "0.1.0",
    requires_python = ">=3.12",
)

vscode = python.python_project(
    name = "tally-vscode",
    parent = root,
    path = "_integrations/vscode-tally",
    version = "0.1.0",
    requires_python = ">=3.12",
)

intellij = python.python_project(
    name = "tally-intellij",
    parent = root,
    path = "_integrations/intellij-tally",
    version = "0.1.0",
    requires_python = ">=3.12",
)

uv.uv_workspace(
    root = root,
    members = [vscode, intellij],
)
