load("@bazel_skylib//lib:shell.bzl", "shell")
load("//bazeldnf:toolchain.bzl", "BAZELDNF_TOOLCHAIN")

def _collect_lockfile_args(ctx):
    lockfile_args = []

    for exclude in ctx.attr.excludes:
        lockfile_args.extend(["--force-ignore-with-dependencies", shell.quote(exclude)])

    if ctx.attr.rpms:
        lockfile_args.extend(ctx.attr.rpms)

    if lockfile_args:
        lockfile_args = ["-r", ctx.attr.repofile, "--lockfile", ctx.attr.lock_file] + lockfile_args

    if ctx.attr.nobest:
        lockfile_args.append("--nobest")

    lockfile_args.append("--ignore-missing")

    return lockfile_args

def _update_lock_file_impl(ctx):
    out_file = ctx.actions.declare_file(ctx.label.name + ".bash")

    bazeldnf = ctx.toolchains[BAZELDNF_TOOLCHAIN]

    substitutions = {
        "@@BAZELDNF_SHORT_PATH@@": bazeldnf._tool.short_path,
        "@@LOCK_FILE@@": shell.quote(ctx.attr.lock_file),
        "@@REPO_NAME@@": ctx.workspace_name,
        "@@LOCKFILE_ARGS@@": " ".join(_collect_lockfile_args(ctx)),
    }

    ctx.actions.expand_template(
        template = ctx.file._runner,
        output = out_file,
        substitutions = substitutions,
        is_executable = True,
    )

    runfiles = ctx.runfiles(
        files = [bazeldnf._tool],
    )

    return [DefaultInfo(
        files = depset([out_file]),
        runfiles = runfiles,
        executable = out_file,
    )]

update_lock_file = rule(
    implementation = _update_lock_file_impl,
    attrs = {
        "lock_file": attr.string(),
        "rpms": attr.string_list(),
        "excludes": attr.string_list(),
        "repofile": attr.string(),
        "nobest": attr.bool(default = False),
        "_runner": attr.label(allow_single_file = True, default = Label("//bazeldnf/private:update-lock-file.sh")),
    },
    toolchains = [
        BAZELDNF_TOOLCHAIN,
    ],
    executable = True,
)
