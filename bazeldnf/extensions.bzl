"""Extensions for bzlmod.

Installs the bazeldnf toolchain.

based on: https://github.com/bazel-contrib/rules-template/blob/0dadcb716f06f672881681155fe6d9ff6fc4a4f4/mylang/extensions.bzl
"""

load("@bazel_features//:features.bzl", "bazel_features")
load("//internal:rpm.bzl", rpm_repository = "rpm")
load(":repositories.bzl", "bazeldnf_register_toolchains")

_DEFAULT_NAME = "bazeldnf"

def _toolchain_extension(module_ctx):
    registrations = {}
    repos = []

    for mod in module_ctx.modules:
        for toolchain in mod.tags.toolchain:
            if toolchain.name != _DEFAULT_NAME and not mod.is_root:
                fail("""\
                Only the root module may override the default name for the bazeldnf toolchain.
                This prevents conflicting registrations in the global namespace of external repos.
                """)
            if mod.is_root and toolchain.disable:
                break
            registrations[toolchain.name] = 1
            if mod.is_root:
                repos.append(toolchain.name + "_toolchains")

    for name in registrations.keys():
        bazeldnf_register_toolchains(
            name = name,
            register = False,
        )

    for mod in module_ctx.modules:
        for rpm in mod.tags.rpm:
            rpm_repository(
                name = rpm.name,
                urls = rpm.urls,
                sha256 = rpm.sha256,
                integrity = rpm.integrity,
            )

            if mod.is_root:
                repos.append(rpm.name)

    kwargs = {}
    if bazel_features.external_deps.extension_metadata_has_reproducible:
        kwargs["reproducible"] = True

    if module_ctx.root_module_has_non_dev_dependency:
        kwargs["root_module_direct_deps"] = repos
        kwargs["root_module_direct_dev_deps"] = []
    else:
        kwargs["root_module_direct_deps"] = []
        kwargs["root_module_direct_dev_deps"] = repos

    return module_ctx.extension_metadata(**kwargs)

_toolchain_tag = tag_class(
    attrs = {
        "name": attr.string(
            doc = """\
Base name for generated repositories, allowing more than one bazeldnf toolchain to be registered.
Overriding the default is only permitted in the root module.
""",
            default = _DEFAULT_NAME,
        ),
        "disable": attr.bool(default = False),
    },
    doc = "Allows registering a prebuilt bazeldnf toolchain",
)

_rpm_tag = tag_class(
    attrs = {
        "name": attr.string(doc = "Name of the generated repository"),
        "urls": attr.string_list(doc = "URLs from which to download the RPM file"),
        "sha256": attr.string(doc = """\
The expected SHA-256 of the file downloaded.
This must match the SHA-256 of the file downloaded.
_It is a security risk to omit the SHA-256 as remote files can change._
At best omitting this field will make your build non-hermetic.
It is optional to make development easier but either this attribute or
`integrity` should be set before shipping.
"""),
        "integrity": attr.string(doc = """\
Expected checksum in Subresource Integrity format of the file downloaded.
This must match the checksum of the file downloaded.
_It is a security risk to omit the checksum as remote files can change._
At best omitting this field will make your build non-hermetic.
It is optional to make development easier but either this attribute or
`sha256` should be set before shipping.
"""),
    },
    doc = "Allows registering a Bazel repository wrapping an RPM file",
)

bazeldnf = module_extension(
    implementation = _toolchain_extension,
    tag_classes = {
        "toolchain": _toolchain_tag,
        "rpm": _rpm_tag,
    },
)
