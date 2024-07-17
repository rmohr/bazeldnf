"Provides a wrapper to run bazeldnf as a run target"

load("@bazel_skylib//lib:shell.bzl", "shell")
load("//bazeldnf:toolchain.bzl", "BAZELDNF_TOOLCHAIN")

def _bazeldnf_impl(ctx):
    transitive_dependencies = []
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    args = []
    if ctx.attr.command:
        args += [ctx.attr.command]
    if ctx.attr.rulename:
        args += ["--name", ctx.attr.rulename]
    if ctx.attr.rpmtree:
        args += ["--rpmtree", ctx.attr.rpmtree]
    if ctx.file.tar:
        args += ["--input", ctx.file.tar.path]
        transitive_dependencies += [ctx.attr.tar.files]
    for lib in ctx.attr.libs:
        args += [lib]

    toolchain = ctx.toolchains[BAZELDNF_TOOLCHAIN]

    substitutions = {
        "@@BAZELDNF_SHORT_PATH@@": shell.quote(toolchain._tool.short_path),
        "@@ARGS@@": shell.array_literal(args),
    }
    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )
    runfiles = ctx.runfiles(
        files = [toolchain._tool],
        transitive_files = depset([], transitive = transitive_dependencies),
    )
    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

_bazeldnf = rule(
    implementation = _bazeldnf_impl,
    attrs = {
        "command": attr.string(
            values = [
                "",
                "ldd",
                "sandbox",
            ],
            default = "",
        ),
        "rulename": attr.string(),
        "rpmtree": attr.string(),
        "libs": attr.string_list(),
        "tar": attr.label(allow_single_file = True),
        "_runner": attr.label(
            default = "@bazeldnf//internal:runner.bash.template",
            allow_single_file = True,
        ),
    },
    toolchains = [BAZELDNF_TOOLCHAIN],
    executable = True,
)

def bazeldnf(**kwargs):
    if kwargs.get("rpmtree"):
        kwargs["tar"] = kwargs["rpmtree"]
    _bazeldnf(
        **kwargs
    )
