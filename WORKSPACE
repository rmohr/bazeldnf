workspace(name = "bazeldnf")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazel_skylib",
    sha256 = "f24ab666394232f834f74d19e2ff142b0af17466ea0c69a3f4c276ee75f6efce",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.4.0/bazel-skylib-1.4.0.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.0/bazel-skylib-1.4.0.tar.gz",
    ],
)

load("@bazel_skylib//:workspace.bzl", "bazel_skylib_workspace")

bazel_skylib_workspace()

http_archive(
    name = "com_google_protobuf",
    sha256 = "930c2c3b5ecc6c9c12615cf5ad93f1cd6e12d0aba862b572e076259970ac3a53",
    strip_prefix = "protobuf-3.21.12",
    urls = [
        "https://github.com/protocolbuffers/protobuf/archive/refs/tags/v3.21.12.tar.gz",
    ],
)

load("@com_google_protobuf//:protobuf_deps.bzl", "protobuf_deps")

protobuf_deps()

http_archive(
    name = "io_bazel_rules_go",
    sha256 = "099a9fb96a376ccbbb7d291ed4ecbdfd42f6bc822ab77ae6f1b5cb9e914e94fa",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/rules_go/releases/download/v0.35.0/rules_go-v0.35.0.zip",
        "https://github.com/bazelbuild/rules_go/releases/download/v0.35.0/rules_go-v0.35.0.zip",
    ],
)

http_archive(
    name = "bazel_gazelle",
    sha256 = "efbbba6ac1a4fd342d5122cbdfdb82aeb2cf2862e35022c752eaddffada7c3f3",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-gazelle/releases/download/v0.27.0/bazel-gazelle-v0.27.0.tar.gz",
        "https://github.com/bazelbuild/bazel-gazelle/releases/download/v0.27.0/bazel-gazelle-v0.27.0.tar.gz",
    ],
)

load("@io_bazel_rules_go//go:deps.bzl", "go_register_toolchains", "go_rules_dependencies")
load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")
load("//:deps.bzl", "bazeldnf_dependencies", "rpm")

# gazelle:repository_macro deps.bzl%bazeldnf_dependencies
bazeldnf_dependencies()

go_rules_dependencies()

go_register_toolchains(version = "1.19.2")

gazelle_dependencies()

rpm(
    name = "libvirt-libs-6.1.0-2.fc32.x86_64.rpm",
    sha256 = "3a0a3d88c6cb90008fbe49fe05e7025056fb9fa3a887c4a78f79e63f8745c845",
    urls = [
        "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/32/Everything/x86_64/os/Packages/l/libvirt-libs-6.1.0-2.fc32.x86_64.rpm",
        "https://storage.googleapis.com/builddeps/3a0a3d88c6cb90008fbe49fe05e7025056fb9fa3a887c4a78f79e63f8745c845",
    ],
)

rpm(
    name = "libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
    sha256 = "2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
    urls = [
        "https://download-ib01.fedoraproject.org/pub/fedora/linux/releases/32/Everything/x86_64/os/Packages/l/libvirt-devel-6.1.0-2.fc32.x86_64.rpm",
        "https://storage.googleapis.com/builddeps/2ebb715341b57a74759aff415e0ff53df528c49abaa7ba5b794b4047461fa8d6",
    ],
)

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")

http_file(
    name = "bazeldnf-0.5.6-linux-amd64",
    executable = True,
    urls = [],  # TODO
)
