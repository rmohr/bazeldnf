"""Extensions for bzlmod.

Installs the bazeldnf toolchain.

based on: https://github.com/bazel-contrib/rules-template/blob/0dadcb716f06f672881681155fe6d9ff6fc4a4f4/mylang/extensions.bzl
"""

load("@bazel_features//:features.bzl", "bazel_features")
load("//bazeldnf/private:toolchains_repo.bzl", "toolchains_repo")
load("//internal:rpm.bzl", rpm_repository = "rpm")
load(":repositories.bzl", "bazeldnf_register_toolchains")

_ALIAS_TEMPLATE = """\
alias(
    name = "{name}",
    actual = "@{actual}//rpm",
    visibility = ["//visibility:public"],
)
"""

def _proxy_repo_impl(repository_ctx):
    aliases = [
        _ALIAS_TEMPLATE.format(
            name = v,
            actual = k.workspace_name,
        )
        for k, v in repository_ctx.attr.rpms.items()
    ]
    repository_ctx.file("WORKSPACE", "workspace(name = \"{name}\")".format(name = repository_ctx.name))
    repository_ctx.file("BUILD.bazel", "\n".join(aliases))

_proxy_repo = repository_rule(
    implementation = _proxy_repo_impl,
    attrs = {
        "rpms": attr.label_keyed_string_dict(),
    },
)

_DEFAULT_NAME = "bazeldnf"

def _handle_rpms(alias, mod, module_ctx):
    if not mod.tags.rpm:
        return {}

    rpms = []

    for rpm in mod.tags.rpm:
        name = rpm.name

        if not rpm.urls:
            fail("urls must be specified for %s" % name)

        if not name:
            url = rpm.urls[0]

            # expect the url to be something like host/name-{version}-{release}.{arch}.rpm
            name, _, _ = url.rsplit("-", 2)
            name = name.rsplit("/", 1)[-1]

        rpm_repository(
            name = name,
            urls = rpm.urls,
            sha256 = rpm.sha256,
            integrity = rpm.integrity,
        )

        rpms.append(name)

    _proxy_repo(
        name = alias,
        rpms = dict([["@@%s" % x, x] for x in rpms]),
    )

    return rpms

def _toolchain_extension(module_ctx):
    repos = {}
    dev_repos = {}

    for mod in module_ctx.modules:
        for toolchain in mod.tags.toolchain:
            if toolchain.name != _DEFAULT_NAME and not mod.is_root:
                fail("""\
                Only the root module may override the default name for the bazeldnf toolchain.
                This prevents conflicting registrations in the global namespace of external repos.
                """)

            bazeldnf_register_toolchains(
                name = toolchain.name,
                register = False,
            )
            if mod.is_root:
                repos["%s_toolchains" % toolchain.name] = 1

        alias = "%s-rpms" % _DEFAULT_NAME
        dev_dependency = False
        for _alias in mod.tags.alias:
            alias = _alias.name
            dev_dependency = _alias.dev_dependency

        rpms = _handle_rpms(alias, mod, module_ctx)
        if rpms:
            if not dev_dependency:
                repos[alias] = 1
            else:
                dev_repos[alias] = 1

    kwargs = {
        "root_module_direct_deps": repos.keys(),
        "root_module_direct_dev_deps": dev_repos.keys(),
    }

    # once we move to bazel 7 this becomes True by default
    if bazel_features.external_deps.extension_metadata_has_reproducible:
        kwargs["reproducible"] = True

    return module_ctx.extension_metadata(**kwargs)

bazeldnf_toolchain = tag_class(attrs = {
    "name": attr.string(doc = """\
Base name for generated repositories, allowing more than one bazeldnf toolchain to be registered.
Overriding the default is only permitted in the root module.
""", default = _DEFAULT_NAME),
})

rpm_tag = tag_class(attrs = {
    "integrity": attr.string(doc = "Subresource Integrity (SRI) hash of the RPM file"),
    "name": attr.string(doc = "Exposed name for the RPM Bazel repository, defaults to extract from the first url"),
    "sha256": attr.string(doc = "SHA256 hash of the RPM file"),
    "urls": attr.string_list(doc = "URLs of the RPM file to download from"),
})

bazeldnf = module_extension(
    implementation = _toolchain_extension,
    tag_classes = {
        "toolchain": bazeldnf_toolchain,
        "rpm": rpm_tag,
        "alias": tag_class(attrs = {
            "name": attr.string(doc = "The name of the alias repository"),
            "dev_dependency": attr.bool(doc = "Whether the alias repository is a dev dependency", default = False),
        }),
    },
)
