"""Extensions for bzlmod.

Installs the bazeldnf toolchain.

based on: https://github.com/bazel-contrib/rules-template/blob/0dadcb716f06f672881681155fe6d9ff6fc4a4f4/mylang/extensions.bzl
"""

load("@bazel_features//:features.bzl", "bazel_features")
load("//internal:rpm.bzl", rpm_repository = "rpm")
load(":repositories.bzl", "bazeldnf_register_toolchains")

_DEFAULT_NAME = "bazeldnf"

def _bazeldnf_toolchain_extension(module_ctx):
    repos = []
    for mod in module_ctx.modules:
        for toolchain in mod.tags.register:
            if toolchain.name != _DEFAULT_NAME and not mod.is_root:
                fail("""\
                Only the root module may override the default name for the bazeldnf toolchain.
                This prevents conflicting registrations in the global namespace of external repos.
                """)
            if mod.is_root and toolchain.disable:
                break
            bazeldnf_register_toolchains(
                name = toolchain.name,
                register = False,
            )
            if mod.is_root:
                repos.append(toolchain.name + "_toolchains")

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

bazeldnf_toolchain = module_extension(
    implementation = _bazeldnf_toolchain_extension,
    tag_classes = {
        "register": _toolchain_tag,
    },
)

_ALIAS_TEMPLATE = """\
alias(
    name = "{name}",
    actual = "@{name}//rpm",
    visibility = ["//visibility:public"],
)
"""

def _alias_repository_impl(repository_ctx):
    """Creates a repository that aliases other repositories."""
    repository_ctx.file("WORKSPACE", "")
    for rpm in repository_ctx.attr.rpms:
        repo_name = rpm.repo_name
        repository_ctx.file("%s/BUILD.bazel" % repo_name, _ALIAS_TEMPLATE.format(name = repo_name))

_alias_repository = repository_rule(
    implementation = _alias_repository_impl,
    attrs = {
        "rpms": attr.label_list(),
    },
)

def _handle_lock_file(lock_file, module_ctx):
    content = module_ctx.read(lock_file)
    lock_file_json = json.decode(content)
    name = lock_file_json.get("name", lock_file.name.rsplit(".json", 1)[0])

    rpms = []

    for rpm in lock_file_json.get("rpms", []):
        rpm_name = rpm.pop("name", None)
        if not rpm_name:
            urls = rpm.get("urls", [])
            if len(urls) < 1:
                fail("invalid entry in %s for %s" % (lock_file, rpm_name))
            rpm_name = urls[0].rsplit("/", 1)[-1]
        rpm_repository(name = rpm_name, **rpm)
        rpms.append(rpm_name)

    _alias_repository(
        name = name,
        rpms = ["@@%s//rpm" % x for x in rpms],
    )

    return name

def _bazeldnf_extension(module_ctx):
    repos = []

    for mod in module_ctx.modules:
        legacy = True
        name = "bazeldnf_rpms"
        for config in mod.tags.config:
            if not config.legacy_mode:
                legacy = False
                name = config.name or name
            if config.lock_file:
                repos.append(_handle_lock_file(config.lock_file, module_ctx))

        rpms = []

        for rpm in mod.tags.rpm:
            rpm_repository(
                name = rpm.name,
                urls = rpm.urls,
                sha256 = rpm.sha256,
                integrity = rpm.integrity,
            )

            if mod.is_root and legacy:
                repos.append(rpm.name)
            else:
                rpms.append(rpm.name)

        if not legacy and rpms:
            _alias_repository(
                name = name,
                rpms = ["@@%s//rpm" % x for x in rpms],
            )
            repos.append(name)

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

_config_tag = tag_class(
    attrs = {
        "legacy_mode": attr.bool(
            default = True,
            doc = """\
If true, the module is loaded in legacy mode and exposes one Bazel repository \
per rpm entry in this invocation of the bazel extension.
""",
        ),
        "name": attr.string(
            doc = "Name of the generated proxy repository",
            default = "bazeldnf_rpms",
        ),
        "lock_file": attr.label(
            doc = """\
Label of the JSON file that contains the RPMs to expose, there's no legacy mode \
for RPMs defined by a lock file.

The lock file content is as:
```json
    {
        "name": "optional name for the proxy repository, defaults to the file name",
        "rpms": [
            {
                "name": "<name of the rpm>",
                "urls": ["<url0>", ...],
                "sha256": "<sha256 of the file>",
                "integrity": "<integrity of the file>"
            }
        ]
    }
```
""",
            allow_single_file = [".json"],
        ),
    },
)

bazeldnf = module_extension(
    implementation = _bazeldnf_extension,
    tag_classes = {
        "rpm": _rpm_tag,
        "config": _config_tag,
    },
)
