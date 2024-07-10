# bazeldnf

Bazel library which allows dealing with the whole RPM dependency lifecycle
solely with pure go rules and a static go binary.

## Bazel rules

### rpm rule

The `rpm` rule represents a pure RPM dependency. This dependency is not
processed in any way.  They can be added to your `WORKSPACE` file like this:

```python
load("@bazeldnf//bazeldnf:deps.bzl", "rpm")

rpm(
    name = "libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
    sha256 = "2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
    urls = [
        "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/32/Everything/x86_64/os/Packages/l/libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
        "https://storage.googleapis.com/builddeps/2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
    ],
)
```

### rpmtree

`rpmtree` Takes a list of `rpm` dependencies and merges them into a single
`tar` package.  `rpmtree` rules can be added like this to your `BUILD` files:

```python
load("@bazeldnf//bazeldnf:defs.bzl", "rpmtree")

rpmtree(
    name = "rpmarchive",
    rpms = [
        "@libvirt-libs-6.1.0-2.fc32.x86_64.rpm//rpm",
        "@libvirt-devel-6.1.0-2.fc32.x86_64.rpm//rpm",
    ],
)
```

Since `rpmarchive` is just a tar archive, it can be put into a container
immediately:

```python
container_layer(
    name = "gcloud-layer",
    tars = [
        ":rpmarchive",
    ],
)
```

rpmtrees allow injecting relative symlinks (`pkg_tar` can only inject absolute
symlinks) and xattrs `capabilities`.  The following example adds a relative
link and gives one binary the `cap_net_bind_service` capability to connect to
privileged ports:

```yaml
rpmtree(
    name = "rpmarchive",
    rpms = [
        "@libvirt-libs-6.1.0-2.fc32.x86_64.rpm//rpm",
        "@libvirt-devel-6.1.0-2.fc32.x86_64.rpm//rpm",
    ],
    symlinks = {
        "/var/run": "../run",
    },
    capabilities = {
        "/usr/libexec/qemu-kvm": [
            "cap_net_bind_service",
        ],
    },
)
```

## Running bazeldnf with bazel

The bazeldnf repository needs to be added  to your `WORKSPACE`:

<!-- install_start -->
```python
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazeldnf",
    sha256 = "fb24d80ad9edad0f7bd3000e8cffcfbba89cc07e495c47a7d3b1f803bd527a40",
    urls = [
        "https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9.tar.gz",
    ],
)

load("@bazeldnf//bazeldnf:deps.bzl", "bazeldnf_dependencies")

bazeldnf_dependencies()
```
<!-- install_end -->

Define the `bazeldnf` executable rule in your `BUILD.bazel` file:
```python
load("@bazeldnf//bazeldnf:defs.bzl", "bazeldnf")

bazeldnf(name = "bazeldnf")
```

After adding this code, you can run bazeldnf with Bazel:

```bash
bazel run //:bazeldnf -- --help
```

## Libraries and Headers

One important use-case is to expose headers and libraries inside the RPMs to build targets in bazel.
If we would just blindly expose all libraries to build targets, bazel would try to link any one of them to our binary.
This would obviously not work. Therefore we need a mediator
between `cc_library` and `rpmtree`. This mediator is the `tar2files` target. This target allows extracting
a subset of libraries and headers and providing them to `cc_library` targets.

An example:

```python
load("@bazeldnf//bazeldnf:defs.bzl", "rpm", "rpmtree", "tar2files")

tar2files(
    name = "libvirt-libs",
    files = {
        "/usr/include/libvirt": [
            "libvirt-admin.h",
            "libvirt-common.h",
            "libvirt-domain-checkpoint.h",
            "libvirt-domain-snapshot.h",
            "libvirt-domain.h",
            "libvirt-event.h",
        ],
        "/usr/lib64": [
            "libacl.so.1",
            "libacl.so.1.1.2253",
            "libattr.so.1",
        ],
    },
    tar = ":libvirt-devel",
    visibility = ["//visibility:public"],
)
```

`tar` can take any input which is a tar archive. Conveniently this is what `rpmtree` creates as the default target.
So any `rpmtree` can be used here.
The `files` section contains then files per folder which one wants to expose to `cc_library`:

```python
cc_library(
    name = "rpmlibs",
    srcs = [
        ":libvirt-libs/usr/lib64",
    ],
    hdrs = [
        ":libvirt-libs/usr/include/libvirt",
    ],
    strip_include_prefix="/libvirt-libs/",
    prefix= "libvirt",
)
```

At this point source code linking to these libraries can be compiled, but unit tests would only work if we would manually
list any transitive library. This would be tedious and error prone. However bazeldnf can introspect for you shared libraries
and create `tar2files` rules for you, based on a provided set of libraries.

First define a target like this:

```python
load("@bazeldnf//bazeldnf:defs.bzl", "bazeldnf", "rpm", "rpmtree", "tar2files")

bazeldnf(
    name = "ldd",
    command = "ldd",
    libs = [
        "/usr/lib64/libvirt-lxc.so.0",
        "/usr/lib64/libvirt-qemu.so.0",
        "/usr/lib64/libvirt.so.0",
    ],
    rpmtree = ":libvirt-devel",
    rulename = "libvirt-libs",
)
```

`rulename` containes the `tar2files` target name, `rpmtree` references a given `rpmtree` and `libs` contains
libraries which one wants to link. When now executing the target like this:

```bash
bazel run //:ldd
```

the `tar2files` target will be updated with all transitive library dependencies for  the specified libraries.
In addition, all header directories are updated too for convenience.

## Dependency resolution

One key part of managing RPM dependencies and RPM repository updates via bazel
is the ability to resolve RPM dependencies from repos without external tools
like `dnf` or `yum` and write the resolved dependencies to your `WORKSPACE`.

Here an example on how to add libvirt and bash to your WORKSPACE and BUILD
files.

First write the `repo.yaml` file which contains some basic rpm repos to query:

```bash
bazeldnf init --fc 32 # write a repo.yaml file containing the usual release and update repos for fc32
```

Then write a `rpmtree` rule called `libvirttree` to your BUILD file and all
corresponding RPM dependencies into your WORKSPACE for libvirt:
```bash
bazeldnf rpmtree --workspace /my/WORKSPACE --buildfile /my/BUILD.bazel --name libvirttree libvirt
```

Do the same for bash with a `bashrpmtree` target:

```bash
bazeldnf rpmtree --workspace /my/WORKSPACE --buildfile /my/BUILD.bazel --name bashtree bash
```

Finally prune all unreferenced old RPM files:

```bash
bazeldnf prune --workspace /my/WORKSPACE --buildfile /my/BUILD.bazel
```

By default `bazeldnf rpmtree` will try to find a solution which only contains
the newest packages of all involved repositories. The only exception are pinned
versions themselves. If pinned version require other outdated packages,
the `--nobest` option can be supplied. With this option all packages are
considered. Newest packages will have the higest weight but it may not always be
able to choose them and older packages may be pulled in instead.

### Dependency resolution limitations

##### Missing features

 * Resolving `requires` entries which contain boolean logic like `(gcc if something)`

##### Deliberately not supported

The goal is to build minimal containers with RPMs based on scratch containers.
Therefore the following RPM repository hints will be ignored:

 * `recommends`
 * `supplements`
 * `suggests`
 * `enhances`
