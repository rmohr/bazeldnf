"Wraps bazeldnf tool binary through a toolchain"

def _bazeldnf_toolchain(ctx):
    return [
        platform_common.ToolchainInfo(
            _tool = ctx.file.tool,
        ),
    ]

bazeldnf_toolchain = rule(
    _bazeldnf_toolchain,
    attrs = {
        "tool": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "bazeldnf executable",
        ),
    },
    provides = [platform_common.ToolchainInfo],
)

BAZELDNF_TOOLCHAIN = "@bazeldnf//bazeldnf:toolchain"
