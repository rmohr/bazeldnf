#bazeldnf

Bazel library which allows dealing with the whole RPM dependency lifecycle solely with pure go rules and a static go binary.

## Bazel rules

### rpm rule

The `rpm` rule represents a pure RPM dependency. This dependency is not processed in any way.
 They can be added to your `WORKSPACE` file like this:

```python
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

`rpmtree` Takes a list of `rpm` dependencies and merges them into a single `tar` package.
`rpmtree` rules can be added like this to your `BUILD` files:

```python
rpmtree(
    name = "rpmarchive",
    rpms = [
        "@libvirt-libs-6.1.0-2.fc32.x86_64.rpm//rpm",
        "@libvirt-devel-6.1.0-2.fc32.x86_64.rpm//rpm",
    ],
    include_dir = "/usr/include",
    lib_dir = "/usr/lib64",
)
```

Since `rpmarchive` is just a tar archive, it can be put into a container immediately:

```python
container_layer(
    name = "gcloud-layer",
    tars = [
        ":rpmarchive",
    ],
)
```

## Libraries and Headers

`rpmtree` can also be used to satisvy C and C++ dependencies like this:

```python
cc_library(
    name = "rpmlibs",
    srcs = [
        ":rpmarchive/libs.tar",
    ],
    hdrs = [
        ":rpmarchive/hdrs.tar",
    ],
)
```

The `include_dir` attribute for `rpmtree` tells the target to include headers in that directory in `<target>/hdrs.tar`.
The same for the `lib_dir` attribute on `rpmtree` for libarires. It can be accessed via `<target>/libs.tar`.
These two tar files have in addition `include_dir` and `lib_dir` prefixes stripped from the resulting archive,
which should make it unnecessary to use the strip options on `cc_library`.

## Dependency resolution

One key part of managing RPM dependencies and RPM repository updates via bazel is the ability to
resolve RPM dependencies from repos without external tools like `dnf` or `yum` and write the resolved
dependencies to your `WORKSPACE`.

Here an example on how to add libvirt and bash to your WORKSPACE and BUILD files.

First write the `repo.yaml` file which contains some basic rpm repos to query:

```bash
bazeldnf init --fc 32 # write a repo.yaml file containing the usual release and update repos for fc32
```

Then write a `rpmtree` rule called `libvirttree` to your BUILD file and all corresponding RPM dependencies into your WORKSPACE
for libvirt:
```bash
bazeldnf resolve --workspace /my/WORKSPACE --buildfile /my/BUILD.bazel --rpmtree libvirttree libvirt
```

Do the same for bash with a `bashrpmtree` target:

```bash
bazeldnf resolve --workspace /my/WORKSPACE --buildfile /my/BUILD.bazel --rpmtree bashtree bash
```

Finally prune all unreferenced old RPM files:

```bash
bazeldnf prune --workspace /my/WORKSPACE
```