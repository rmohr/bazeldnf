load("@bazel_skylib//lib:shell.bzl", "shell")

def _bazeldnf_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    args = []
    if ctx.attr.command:
        args += [ctx.attr.command]
    if ctx.attr.rulename:
        args += ["--name", ctx.attr.rulename]
    if ctx.attr.rpmtree:
        args += ["--rpmtree", ctx.attr.rpmtree]
    if ctx.file.tar:
        args += ["-i", ctx.file.tar.path]
    for lib in ctx.attr.libs:
        args += [lib]

    substitutions = {
        "@@BAZELDNF_SHORT_PATH@@": shell.quote(ctx.executable._bazeldnf.short_path),
        "@@ARGS@@": shell.array_literal(args),
    }
    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )
    runfiles = ctx.runfiles(files = [ctx.executable._bazeldnf])
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
            ],
            default = "",
        ),
        "rulename": attr.string(),
        "rpmtree": attr.string(),
        "libs": attr.string_list(),
        "tar": attr.label(allow_single_file = True),
        "_bazeldnf": attr.label(
            default = "@bazeldnf//cmd:cmd",
            cfg = "host",
            executable = True,
        ),
        "_runner": attr.label(
            default = "@bazeldnf//internal:runner.bash.template",
            allow_single_file = True,
        ),
    },
    executable = True,
)

def bazeldnf(**kwargs):
    if kwargs.get("rpmtree"):
        kwargs["tar"] = kwargs["rpmtree"]
    _bazeldnf(
        **kwargs
    )
