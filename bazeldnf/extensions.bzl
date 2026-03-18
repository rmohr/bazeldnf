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
load("@bazeldnf//bazeldnf:alias_macros.bzl", aliases="default")

aliases(
    name = "{name}",
    rpms = {data},
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
    architectures = {architectures},
    visibility = ["//visibility:public"],
)
"""

_UPDATE_LOCK_FILE_TEMPLATE = """\
fail("Lock file hasn't been generated for this repository, please run `bazel run @{repo}//:update-lock-file` first")
"""

def _repository_file(path):
    if path.repo_name == "":
        if path.package:
            return "%s/%s" % (path.package, path.name)
        return path.name
    fail("lock file can only be updated if it belongs to the bazel workspace")

def _alias_repository_impl(repository_ctx):
    """Creates a repository that aliases other repositories."""
    repository_ctx.file("WORKSPACE", "")
    repository_ctx.watch(repository_ctx.attr.lock_file)

    repofile = "invalid-repo.yaml"
    if repository_ctx.attr.repofile:
        repofile = _repository_file(repository_ctx.attr.repofile)

    lock_file_path = _repository_file(repository_ctx.attr.lock_file)

    repository_ctx.file(
        "BUILD.bazel",
        _ALIAS_REPO_TOP_LEVEL_TEMPLATE.format(
            cache_dir = '"{}"'.format(repository_ctx.attr.cache_dir) if repository_ctx.attr.cache_dir else None,
            path = lock_file_path,
            rpms = ", ".join(["'{}'".format(x) for x in repository_ctx.attr.rpms_to_install]),
            excludes = ", ".join(["'{}'".format(x) for x in repository_ctx.attr.excludes]),
            repofile = repofile,
            nobest = "True" if repository_ctx.attr.nobest else "False",
            architectures = repr(repository_ctx.attr.architectures),
        ),
    )

    for name, metadata in repository_ctx.attr.packages_metadata.items():
        repository_ctx.file(
            "%s/BUILD.bazel" % name,
            _ALIAS_TEMPLATE.format(
                name = name,
                data = metadata,
            ),
        )

    if not repository_ctx.attr.packages_metadata:
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
        "packages_metadata": attr.string_dict(),
        "lock_file": attr.label(),
        "rpms_to_install": attr.string_list(),
        "excludes": attr.string_list(),
        "repofile": attr.label(),
        "repository_prefix": attr.string(),
        "nobest": attr.bool(default = False),
        "cache_dir": attr.string(),
        "architectures": attr.string_list(),
    },
)

def _get_architectures(architecture, architectures):
    """ Create effective list of architectures based on user input """
    if architecture and architectures:
        fail("Can't combine `architecture` and `architectures`")
    return architectures or [(architecture or "x86_64")]

def _build_rpm_lookup(rpms_list):
    """Build a dictionary for efficient RPM lookup by name/id.

    Returns a dict mapping RPM name/id to RPM data.
    """
    rpm_lookup = {}
    for rpm in rpms_list:
        # Determine the RPM identifier
        rpm_id = rpm.get("id", rpm.get("name", None))
        if not rpm_id:
            # Fallback to URL-based name
            urls = rpm.get("urls", [])
            if len(urls) > 0:
                rpm_id = urls[0].rsplit("/", 1)[-1]
        if rpm_id:
            rpm_lookup[rpm_id] = rpm
    return rpm_lookup

def _build_transitive_deps(rpm_lookup, target_name):
    """Build transitive dependency closure for a target.

    Returns a dict mapping RPM name to RPM data for all transitive dependencies.
    Uses iterative passes to avoid recursion (not supported in Starlark).
    """
    visited = {}
    to_process = {target_name: True}

    # Iterate up to max depth to resolve all transitive dependencies
    # This is a safety limit to prevent infinite loops in case of circular deps
    for _ in range(1000):
        if len(to_process) == 0:
            break

        # Process all current items
        current_batch = list(to_process.keys())
        to_process = {}

        for current in current_batch:
            # Skip if already processed
            if current in visited:
                continue

            # Find the RPM in the lookup dict
            rpm = rpm_lookup.get(current)
            if not rpm:
                continue

            # Make a copy to avoid mutating the original
            rpm_copy = dict(rpm)
            visited[current] = rpm_copy

            # Add dependencies to next batch
            deps = rpm_copy.get("dependencies", [])
            for dep in deps:
                if dep not in visited:
                    to_process[dep] = True

    return visited

def _handle_lock_file(config, module_ctx, registered_rpms = {}, registered_blobs = {}):
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
        "architectures": _get_architectures(config.architecture, config.architectures),
    }

    module_ctx.watch(config.lock_file)

    if config.cache_dir:
        repository_args["cache_dir"] = config.cache_dir

    # Data for generating alias repository
    # Keyed with a Bazel package name in the root of the alias repository (usually just a RPM package name),
    # Values are a list of resolved RPMs in a form of a dict, containing:
    # - package - just an RPM package name (optional - lock file may be missing it)
    # - id - some unique identifier for the config
    # - repo_name - apparent repo name where the .rpm file is downloaded to
    packages_metadata = {}

    if module_ctx.path(config.lock_file).exists:
        content = module_ctx.read(config.lock_file)
        lock_file_json = json.decode(content)

        # Build lookup dictionary for efficient RPM access
        rpm_lookup = _build_rpm_lookup(lock_file_json.get("rpms", []))

        # Create a blob repository for each available rpm in the lock file
        for rpm in lock_file_json.get("rpms", []):
            _add_blob_rpm_repository(config, rpm, lock_file_json, registered_blobs)

        # Create repositories for each top-level target with suffixed dependencies
        for target in config.rpms:
            if target not in rpm_lookup:
                fail("requested rpm %s is not known to the lock file %s" % (target, config.lock_file))

            # Build transitive dependency closure for this target
            target_deps = _build_transitive_deps(rpm_lookup, target)
            repo_info = _add_rpm_repository(config, rpm_lookup[target], registered_rpms, dependencies = target_deps)

            packages_metadata.setdefault(repo_info.get("package", repo_info["id"]), []).append(repo_info)

        if not config.rpms:
            # if the user didn't ask for a list of RPMs then make all of the RPMs available with no dependencies
            for rpm in rpm_lookup.values():
                blob_name, _, _ = _normalize_repository_name(rpm, config.rpm_repository_prefix, config.lock_file)
                repo_info = _add_rpm_repository(config, rpm, registered_rpms, [blob_name])
                packages_metadata.setdefault(repo_info.get("package", repo_info["id"]), []).append(repo_info)

        for rpm in rpm_lookup.values():
            # for RPMs with no id or name then we will make then available with only a dependency to it's blob
            if rpm.get("id", None) or rpm.get("name"):
                continue
            blob_name, _, _ = _normalize_repository_name(rpm, config.rpm_repository_prefix, config.lock_file)
            repo_info = _add_rpm_repository(config, rpm, registered_rpms, [blob_name])
            packages_metadata.setdefault(repo_info.get("package", repo_info["id"]), []).append(repo_info)

        # if there's targets without matching RPMs we need to create a null target
        # so that consumers have something consistent that they can depend on
        for target in lock_file_json.get("targets", []):
            packages_metadata.setdefault(target, [])
    elif config.ignore_missing_lockfile:
        for target in config.rpms:
            packages_metadata.setdefault(target, [])

    # Encode aliases metadata in a form that could be passed with one of the `attr`-allowed types:
    repository_args["packages_metadata"] = {package: json.encode(metadata) for package, metadata in packages_metadata.items()}

    _alias_repository(
        **repository_args
    )

    return config.name

def _normalize_repository_name(rpm, rpm_repository_prefix, lock_file):
    # Older lockfiles may not have `id` field.
    # Name was the equivalent. We need to pop both.
    package = rpm.get("name", None)
    id = rpm.get("id", package)
    if not id:
        urls = rpm.get("urls", [])
        if len(urls) < 1:
            fail("invalid entry in %s: %s" % (lock_file, rpm))
        id = urls[0].rsplit("/", 1)[-1]

    name = id.replace("+", "plus")
    if rpm_repository_prefix:
        name = "{}{}".format(rpm_repository_prefix, name)

    return name, id, package

def _get_blob_prefix(rpm_repository_prefix):
    if not rpm_repository_prefix:
        return "blob-"
    return "blob-{}-".format(rpm_repository_prefix)

def _add_blob_rpm_repository(config, rpm, lock_file_json, registered_blobs):
    name, _, _ = _normalize_repository_name(rpm, _get_blob_prefix(config.rpm_repository_prefix), config.lock_file)

    # prevent the same blob to be registered more than once, needed for multiple lock files
    if name in registered_blobs:
        return

    registered_blobs[name] = 1

    repository = rpm.get("repository")

    mirrors = lock_file_json.get("repositories", {}).get(repository, None)

    if mirrors == None:
        fail("couldn't resolve %s in %s" % (repository, lock_file_json["repositories"]))

    href = rpm.get("urls")[0]
    urls = ["%s/%s" % (x, href) for x in mirrors]

    rpm_repository(
        name = name,
        urls = urls,
        create_blob = True,
        blob_mode = True,
    )

    return

def _add_rpm_repository(config, rpm, registered_rpms, dependencies = []):
    # fix for cases like c++
    dependencies = [x.replace("+", "plus") for x in dependencies]

    # point to the actual blob
    dependencies = ["@{}{}//blob".format(_get_blob_prefix(config.rpm_repository_prefix), x) for x in dependencies]

    repo_prefix = config.rpm_repository_prefix
    if repo_prefix:
        repo_prefix = "{}-".format(repo_prefix)

    name, id, package = _normalize_repository_name(rpm, repo_prefix, config.lock_file)

    # the same rpm may be in the transitive closure of an already explored rpm, but it may be
    # a requested target, in which case we need to override the previously defined case
    if name in registered_rpms:
        return registered_rpms[name]

    rpm_repository(
        name = name,
        dependencies = dependencies,
        create_blob = False,
        blob_mode = True,
    )

    metadata = {
        "repo_name": name,
        "id": id,
    }

    if package:
        metadata["package"] = package

    registered_rpms[name] = metadata

    return metadata

def _bazeldnf_extension(module_ctx):
    # make sure all our dependencies are registered as those may be needed when those
    # depending in this repo build the toolchain from sources
    repos = []

    # blobs are unique for the entire bazel workspace
    registered_blobs = dict()

    for mod in module_ctx.modules:
        registered_rpms = dict()
        for config in mod.tags.config:
            repos.append(
                _handle_lock_file(
                    config,
                    module_ctx,
                    registered_rpms,
                    registered_blobs,
                ),
            )

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
        "ignore_missing_lockfile": attr.bool(
            doc = """In case lockfile does not exist, create null rpm targets so that clients can still depend on them.

            One won't be prompted with "please run `bazel run @{repo}//:update-lock-file` first".""",
            default = False,
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
            doc = "Single architecture to enable in addition to noarch",
        ),
        "architectures": attr.string_list(
            doc = """Custom list of architectures (can't be used with `architecture`).

                Can use more than one. The list defines architectures priority -
                with the first one having the highest priority.
                `noarch` is implicitly added at the end (if not present on the list).""",
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
