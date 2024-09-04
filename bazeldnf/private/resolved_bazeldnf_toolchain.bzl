"Helper to get a binary that can be passed into rules that can call bazeldnf"

load("//bazeldnf:toolchain.bzl", "BAZELDNF_TOOLCHAIN")

def _resolved_bazeldnf_toolchain(ctx):
    toolchain = ctx.toolchains[BAZELDNF_TOOLCHAIN]
    out = ctx.actions.declare_file("bazeldnf")
    ctx.actions.symlink(
        output = out,
        target_file = toolchain._tool,
        is_executable = True,
    )

    return [DefaultInfo(
        files = depset([out]),
        executable = out,
    )]

resolved_bazeldnf_toolchain = rule(
    implementation = _resolved_bazeldnf_toolchain,
    toolchains = [BAZELDNF_TOOLCHAIN],
    executable = True,
    doc = """\
    Creates a pure binary that can be exposed to other bazel targets

    It allows to pass a bazeldnf (prebuilt or not based on the registered toolchain)
    as:
    ```
    some_language_binary(
        name = "foo",
        data = [
            "@bazeldnf//bazeldnf"
        ],
        env = {
            "BAZELDNF": "$(location @bazeldnf//bazeldnf)"
        }
    )
    ```
    """,
)
