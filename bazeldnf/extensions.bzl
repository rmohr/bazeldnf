"""Extensions for bzlmod.

Installs the bazeldnf toolchain.

based on: https://github.com/bazel-contrib/rules-template/blob/0dadcb716f06f672881681155fe6d9ff6fc4a4f4/mylang/extensions.bzl
"""

load("@bazel_features//:features.bzl", "bazel_features")
load("@bazel_skylib//lib:versions.bzl", "versions")
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_jar")
load("//internal:rpm.bzl", null_rpm_repository = "null_rpm", rpm_repository = "rpm")
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
    actual = "@{actual_name}//rpm",
    visibility = ["{visibility}"],
)
"""

_ALIAS_REPO_TOP_LEVEL_TEMPLATE = """\
load("@bazeldnf//bazeldnf/private:lock-file-helpers.bzl", "fetch_dnf_repo", "update_lock_file")

fetch_dnf_repo(
    name = "fetch-repo",
    repofile = "{repofile}",
    cache_dir = {cache_dir},
    visibility = ["//visibility:public"],
)

update_lock_file(
    name = "update-lock-file",
    lock_file = "{path}",
    rpms = [{rpms}],
    excludes = [{excludes}],
    repofile = "{repofile}",
    nobest = {nobest},
    cache_dir = {cache_dir},
    architecture = "{architecture}",
    visibility = ["//visibility:public"],
)
"""

_UPDATE_LOCK_FILE_TEMPLATE = """\
fail("Lock file hasn't been generated for this repository, please run `bazel run @{repo}//:update-lock-file` first")
"""

def _alias_repository_impl(repository_ctx):
    """Creates a repository that aliases other repositories."""
    repository_ctx.file("WORKSPACE", "")
    repository_ctx.watch(repository_ctx.attr.lock_file)

    repofile = "invalid-repo.yaml"
    if repository_ctx.attr.repofile:
        repofile = repository_ctx.path(repository_ctx.attr.repofile)

    lock_file_path = repository_ctx.path(repository_ctx.attr.lock_file)

    repository_ctx.file(
        "BUILD.bazel",
        _ALIAS_REPO_TOP_LEVEL_TEMPLATE.format(
            cache_dir = '"{}"'.format(repository_ctx.attr.cache_dir) if repository_ctx.attr.cache_dir else None,
            path = lock_file_path,
            rpms = ", ".join(["'{}'".format(x) for x in repository_ctx.attr.rpms_to_install]),
            excludes = ", ".join(["'{}'".format(x) for x in repository_ctx.attr.excludes]),
            repofile = repofile,
            nobest = "True" if repository_ctx.attr.nobest else "False",
            architecture = repository_ctx.attr.architecture,
        ),
    )

    requested = dict([[x, 1] for x in repository_ctx.attr.requested])

    for rpm in repository_ctx.attr.rpms:
        actual_name = rpm.repo_name
        name = rpm.repo_name

        if repository_ctx.attr.repository_prefix:
            name = actual_name.split(repository_ctx.attr.repository_prefix, 1)[-1]

        visibility = "//visibility:public"

        if repository_ctx.attr.requested and name not in requested:
            visibility = "//:__subpackages__"

        repository_ctx.file(
            "%s/BUILD.bazel" % name,
            _ALIAS_TEMPLATE.format(
                name = name,
                actual_name = actual_name,
                visibility = visibility,
            ),
        )

    if not repository_ctx.attr.rpms:
        for rpm in repository_ctx.attr.rpms_to_install:
            repository_ctx.file(
                "%s/BUILD.bazel" % rpm,
                _UPDATE_LOCK_FILE_TEMPLATE.format(
                    repo = repository_ctx.name.rsplit("~", 1)[-1],
                ),
            )

_alias_repository = repository_rule(
    implementation = _alias_repository_impl,
    attrs = {
        "architecture": attr.string(values = ["i686", "x86_64", "aarch64", ""]),
        "cache_dir": attr.string(),
        "excludes": attr.string_list(),
        "lock_file": attr.label(),
        "nobest": attr.bool(default = False),
        "requested": attr.string_list(),
        "repofile": attr.label(),
        "repository_prefix": attr.string(),
        "rpms_to_install": attr.string_list(),
        "rpms": attr.label_list(default = []),
    },
)

def _to_rpm_repo_name(prefix, rpm_name):
    name = rpm_name.replace("+", "plus")
    return "{}{}".format(prefix, name)

def _handle_lock_file(config, module_ctx, registered_rpms = {}):
    if not config.lock_file:
        fail("No lock file provided for %s" % config.name)

    repository_args = {
        "name": config.name,
        "lock_file": config.lock_file,
        "rpms_to_install": config.rpms,
        "excludes": config.excludes,
        "repofile": config.repofile,
        "repository_prefix": config.rpm_repository_prefix,
        "nobest": config.nobest,
        "architecture": config.architecture,
    }

    module_ctx.watch(config.lock_file)

    if config.cache_dir:
        repository_args["cache_dir"] = config.cache_dir

    if not module_ctx.path(config.lock_file).exists:
        _alias_repository(
            **repository_args
        )
        return config.name

    content = module_ctx.read(config.lock_file)
    lock_file_json = json.decode(content)

    if not config.ignore_deps:
        if versions.is_at_least("7", versions.get()) and not versions.is_at_least("7.4.0", versions.get()):
            fail("ignore_deps requires Bazel 7.4+ for Bazel 7")
        if versions.is_at_least("8", versions.get()) and not versions.is_at_least("8.1.0", versions.get()):
            fail("ignore_deps requires Bazel 8.1+ for Bazel 8")

    for rpm in lock_file_json.get("rpms", []):
        dependencies = rpm.pop("dependencies", [])
        if config.ignore_deps:
            dependencies = []
        else:
            dependencies = [x.replace("+", "plus") for x in dependencies]
            dependencies = ["@{}{}//rpm:rpm-file".format(config.rpm_repository_prefix, x) for x in dependencies]

        rpm_name = rpm.pop("name", None)
        if not rpm_name:
            urls = rpm.get("urls", [])
            if len(urls) < 1:
                fail("invalid entry in %s for %s" % (config.lock_file, rpm_name))
            rpm_name = urls[0].rsplit("/", 1)[-1]

        name = _to_rpm_repo_name(config.rpm_repository_prefix, rpm_name)
        if name in registered_rpms:
            continue
        registered_rpms[name] = 1
        repository = rpm.pop("repository")
        mirrors = lock_file_json.get("repositories", {}).get(repository, None)
        if mirrors == None:
            fail("couldn't resolve %s in %s" % (repository, lock_file_json["repositories"]))
        href = rpm.pop("urls")[0]
        urls = ["%s/%s" % (x, href) for x in mirrors]
        rpm_repository(
            name = name,
            dependencies = dependencies,
            urls = urls,
            **rpm
        )

    # if there's targets without matching RPMs we need to create a null target
    # so that consumers have something consistent that they can depend on
    for target in lock_file_json.get("targets", []):
        name = _to_rpm_repo_name(config.rpm_repository_prefix, target)
        if name in registered_rpms:
            continue

        null_rpm_repository(name = name)
        registered_rpms[name] = 1

    repository_args["rpms"] = ["@@%s//rpm" % x for x in registered_rpms.keys()]
    repository_args["requested"] = [x.replace("+", "plus") for x in lock_file_json.get("targets", [])]

    _alias_repository(
        **repository_args
    )

    return config.name

def _bazeldnf_extension(module_ctx):
    # make sure all our dependencies are registered as those may be needed when those
    # dependening in this repo build the toolchain from sources
    repos = []
    for mod in module_ctx.modules:
        legacy = True
        name = "bazeldnf_rpms"
        registered_rpms = dict()
        for config in mod.tags.config:
            repos.append(
                _handle_lock_file(
                    config,
                    module_ctx,
                    registered_rpms,
                ),
            )

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
                requested = rpms,
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
        "name": attr.string(
            doc = "Name of the generated proxy repository",
            default = "bazeldnf_rpms",
        ),
        "cache_dir": attr.string(
            doc = "Path of the bazeldnf cache repository",
        ),
        "lock_file": attr.label(
            doc = """\
Label of the JSON file that contains the RPMs to expose, there's no legacy mode \
for RPMs defined by a lock file.

The lock file content is as:
```json
    {
        "name": "optional name for the proxy repository, defaults to the file name",
        "cli-arguments": [
            "cli",
            "arguments",
            "used",
            "for"
            "generation",
        ],
        "repositories": {
            "repo-name": [
                "https://repo-url/path",
            ],
        },
        "rpms": [
            {
                "name": "<name of the rpm>",
                "urls": ["<url0>", ...],
                "sha256": "<sha256 of the file>",
                "integrity": "<integrity of the file>"
            }
        ],
        "targets": [
            "target to install",
        ],
        "ignored": [
            "ignored package",
        ],
    }
```
""",
            allow_single_file = [".json"],
        ),
        "rpm_repository_prefix": attr.string(
            doc = "A prefix to add to all generated rpm repositories",
            default = "",
        ),
        "repofile": attr.label(
            doc = "YAML file that defines the repositories used for this lock file",
            allow_single_file = [".yaml"],
        ),
        "rpms": attr.string_list(
            doc = "name of the RPMs to install",
        ),
        "excludes": attr.string_list(
            doc = "Regex to pass to bazeldnf to exclude from the dependency tree",
        ),
        "nobest": attr.bool(
            doc = "Allow picking versions which are not the newest",
            default = False,
        ),
        "ignore_deps": attr.bool(
            doc = "Don't include dependencies in resulting repositories",
            default = False,
        ),
        "architecture": attr.string(
            doc = "Architectures to enable in addition to noarch",
            values = ["i686", "x86_64", "aarch64", ""],
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

def _protobuf_java_extension(module_ctx):
    http_jar(
        name = "protobuf-java",
        integrity = "sha256-0C+GOpCj/8d9Xu7AMcGOV58wx8uY8/OoFP6LiMQ9O8g=",
        urls = ["https://repo1.maven.org/maven2/com/google/protobuf/protobuf-java/4.27.3/protobuf-java-4.27.3.jar"],
    )

    kwargs = {}
    if bazel_features.external_deps.extension_metadata_has_reproducible:
        kwargs["reproducible"] = True

    if module_ctx.root_module_has_non_dev_dependency:
        kwargs["root_module_direct_deps"] = []
        kwargs["root_module_direct_dev_deps"] = ["protobuf-java"]
    else:
        kwargs["root_module_direct_deps"] = []
        kwargs["root_module_direct_dev_deps"] = ["protobuf-java"]

    return module_ctx.extension_metadata(**kwargs)

protobuf_java = module_extension(
    implementation = _protobuf_java_extension,
)
