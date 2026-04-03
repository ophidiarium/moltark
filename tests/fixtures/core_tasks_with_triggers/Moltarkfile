python = use("moltark/python")
uv = use("astral/uv")
core = use("moltark/core")

root = python.python_project(
    name = "demo",
    path = ".",
    version = "0.1.0",
    requires_python = ">=3.12",
)

worker = python.python_project(
    name = "demo-worker",
    parent = root,
    path = "packages/worker",
    version = "0.1.0",
    requires_python = ">=3.12",
)

uv.uv_workspace(
    root = root,
    members = [worker],
)

core.synthesis_hook(
    phase = "post_synthesize",
    target = root,
    description = "validate resolved task graph",
)

core.bootstrap_requirement(
    tool = "uv",
    target = worker,
    purpose = "python workspace bootstrap",
    strategies = ["official-installer", "pipx"],
)

core.python_dependency(
    target = worker,
    requirement = "httpx>=0.28",
)

core.task(
    name = "test",
    target = worker,
    command = ["pytest", "-q"],
    runtime = "python",
    tags = ["test", "python"],
)

core.task(
    name = "lint",
    target = worker,
    command = ["ruff", "check", "."],
    runtime = "python",
    tags = ["lint", "python"],
)

core.task_surface(
    kind = "package_scripts",
    name = "uv",
    target = worker,
)

core.trigger_binding(
    trigger = "pre-push",
    target = root,
    match_tags = ["test"],
)

core.trigger_binding(
    trigger = "ci",
    target = root,
    match_names = ["lint"],
)
