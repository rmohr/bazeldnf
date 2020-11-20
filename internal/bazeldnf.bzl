load("@bazel_skylib//lib:shell.bzl", "shell")

def _bazeldnf_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")
    substitutions = {
        "@@BAZELDNF_SHORT_PATH@@": shell.quote(ctx.executable._bazeldnf.short_path),
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
    _bazeldnf(
        **kwargs
    )
