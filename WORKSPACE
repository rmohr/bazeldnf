workspace(name = "bazeldnf")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazel_skylib",
    sha256 = "bc283cdfcd526a52c3201279cda4bc298652efa898b10b4db0837dc51652756f",
    urls = [
        "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.7.1/bazel-skylib-1.7.1.tar.gz",
        "https://github.com/bazelbuild/bazel-skylib/releases/download/1.7.1/bazel-skylib-1.7.1.tar.gz",
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
load("//:build_deps.bzl", "bazeldnf_build_dependencies")
load("//bazeldnf:deps.bzl", "bazeldnf_dependencies", "rpm")

# gazelle:repository_macro build_deps.bzl%bazeldnf_build_dependencies
bazeldnf_build_dependencies()

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

http_archive(
    name = "rules_pkg",
    sha256 = "cad05f864a32799f6f9022891de91ac78f30e0fa07dc68abac92a628121b5b11",
    urls = [
        "https://github.com/bazelbuild/rules_pkg/releases/download/1.0.0/rules_pkg-1.0.0.tar.gz",
    ],
)

load("@rules_pkg//:deps.bzl", "rules_pkg_dependencies")

rules_pkg_dependencies()

http_archive(
    name = "buildifier_prebuilt",
    sha256 = "8ada9d88e51ebf5a1fdff37d75ed41d51f5e677cdbeafb0a22dda54747d6e07e",
    strip_prefix = "buildifier-prebuilt-6.4.0",
    urls = [
        "http://github.com/keith/buildifier-prebuilt/archive/6.4.0.tar.gz",
    ],
)

load("@buildifier_prebuilt//:deps.bzl", "buildifier_prebuilt_deps")

buildifier_prebuilt_deps()

load("@buildifier_prebuilt//:defs.bzl", "buildifier_prebuilt_register_toolchains")

buildifier_prebuilt_register_toolchains()

http_archive(
    name = "aspect_bazel_lib",
    sha256 = "6d758a8f646ecee7a3e294fbe4386daafbe0e5966723009c290d493f227c390b",
    strip_prefix = "bazel-lib-2.7.7",
    url = "https://github.com/aspect-build/bazel-lib/releases/download/v2.7.7/bazel-lib-v2.7.7.tar.gz",
)

load("@aspect_bazel_lib//lib:repositories.bzl", "aspect_bazel_lib_dependencies", "aspect_bazel_lib_register_toolchains")

# Required bazel-lib dependencies

aspect_bazel_lib_dependencies()

# Register bazel-lib toolchains

aspect_bazel_lib_register_toolchains()
