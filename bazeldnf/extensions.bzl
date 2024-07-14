"""Extensions for bzlmod.

Installs the bazeldnf toolchain.

based on: https://github.com/bazel-contrib/rules-template/blob/0dadcb716f06f672881681155fe6d9ff6fc4a4f4/mylang/extensions.bzl
"""

load("//tools:version.bzl", "VERSION")
load(":repositories.bzl", "bazeldnf_register_toolchains")

_DEFAULT_NAME = "bazeldnf"

bazeldnf_toolchain = tag_class(attrs = {
    "name": attr.string(doc = """\
Base name for generated repositories, allowing more than one bazeldnf toolchain to be registered.
Overriding the default is only permitted in the root module.
""", default = _DEFAULT_NAME),
})

def _toolchain_extension(module_ctx):
    registrations = {}
    for mod in module_ctx.modules:
        for toolchain in mod.tags.toolchain:
            if toolchain.name != _DEFAULT_NAME and not mod.is_root:
                fail("""\
                Only the root module may override the default name for the bazeldnf toolchain.
                This prevents conflicting registrations in the global namespace of external repos.
                """)
            registrations[toolchain.name] = 1

    for name in registrations.keys():
        bazeldnf_register_toolchains(
            name = name,
            register = False,
        )

bazeldnf = module_extension(
    implementation = _toolchain_extension,
    tag_classes = {"toolchain": bazeldnf_toolchain},
)
