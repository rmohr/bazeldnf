"A series of helpers for our e2e infrastructure"

load("@bazel_skylib//lib:paths.bzl", "paths")
load(
    "@rules_bazel_integration_test//bazel_integration_test:defs.bzl",
    _default_test_runner = "default_test_runner",
)

def glob_workspace_files(workspace_path, extra_excludes = []):
    """Recursively globs the Bazel workspace files at the specified path.

    Improved from integration_test_utils.glob_workspace_files

    Args:
        workspace_path: A `string` representing the path to glob.
        extra_excludes: Other glob patterns to ignore

    Returns:
        A `list` of the files under the specified path ignoring certain Bazel
        artifacts (e.g. `bazel-*`).
    """
    return native.glob(
        [paths.join(workspace_path, "**", "*")],
        exclude = [
            paths.join(workspace_path, "bazel-*", "**"),
            paths.join(workspace_path, "MODULE.bazel.lock"),
        ] + extra_excludes,
    )

def default_test_runner(name, bazel_cmds = [], **kwargs):
    """GitHub bazel-contrib/setup-bazel action compatible runner

    Args:
        name: name of the target
        bazel_cmds: list of commands to execute
        **kwargs: other arguments to pass to upstream default_test_runner
    """

    bazel_cmds = [
        "--nohome_rc {}".format(x)
        for x in bazel_cmds
    ]

    _default_test_runner(name = name, bazel_cmds = bazel_cmds, **kwargs)
