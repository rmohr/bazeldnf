def _bazeldnf_toolchain(ctx):
    return [
        platform_common.ToolchainInfo(
            _tool = ctx.executable.tool,
        ),
    ]

bazeldnf_toolchain = rule(
    _bazeldnf_toolchain,
    attrs = {
        "tool": attr.label(
            allow_single_file = True,
            mandatory = True,
            cfg = "exec",
            executable = True,
            doc = "bazeldnf executable",
        ),
    },
    provides = [platform_common.ToolchainInfo],
)

PLATFORMS = [
    "linux-amd64",
    "linux-arm64",
    "linux-ppc64",
    "linux-ppc64le",
    "linux-s390x",
    "darwin-amd64",
    "darwin-arm64",
]

BAZELDNF_TOOLCHAIN = "@bazeldnf//bazeldnf:toolchain"

def declare_toolchain(toolchain_prefix, os, arch):  # buildifier: disable=unnamed-macro
    """Create the custom and native toolchain for a platform

    Args:
        toolchain_prefix: The tool the toolchain is being used for
        os: The OS the toolchain is compatible with
        arch: The arch the toolchain is compatible with
    """

    name = "%s-%s" % (os, arch)
    bazeldnf_toolchain(
        name = name,
        tool = "@%s-%s//file" % (toolchain_prefix, name),
    )

    if os == "darwin":
        os = "osx"
    if arch == "amd64":
        arch = "x86_64"
    if arch == "ppc64":
        arch = "ppc"

    native.toolchain(
        name = name + "-toolchain",
        toolchain_type = "@bazeldnf//bazeldnf:toolchain",
        exec_compatible_with = [
            "@platforms//os:%s" % os,
            "@platforms//cpu:%s" % arch,
        ],
        toolchain = name,
    )

def _bazeldnf_prebuilt_setup_impl(repo_ctx):
    toolchain_prefix = repo_ctx.attr.toolchain_prefix
    build_bzl = 'load("@bazeldnf//bazeldnf:toolchain.bzl", "declare_toolchain")'

    for plat in PLATFORMS:
        os, arch = plat.split("-")
        build_bzl += '\ndeclare_toolchain( toolchain_prefix = "{toolchain_prefix}", os = "{os}", arch = "{arch}")'.format(
            toolchain_prefix = toolchain_prefix,
            os = os,
            arch = arch,
        )
    repo_ctx.file("BUILD.bazel", build_bzl)

_bazeldnf_prebuilt_setup = repository_rule(
    implementation = _bazeldnf_prebuilt_setup_impl,
    attrs = {
        "toolchain_prefix": attr.string(),
    },
)

def bazeldnf_prebuilt_register_toolchains(name, toolchain_prefix = "bazeldnf", register_toolchains = True):
    _bazeldnf_prebuilt_setup(
        name = name,
        toolchain_prefix = toolchain_prefix,
    )

    if register_toolchains:
        toolchains = ["@%s//:%s-toolchain" % (name, x) for x in PLATFORMS]
        native.register_toolchains(*toolchains)
