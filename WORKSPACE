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
load("//:build_deps.bzl", "bazeldnf_build_dependencies")
load("//:deps.bzl", "rpm")

# gazelle:repository_macro build_deps.bzl%bazeldnf_build_dependencies
bazeldnf_build_dependencies()

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
    name = "bazeldnf-linux-amd64",
    executable = True,
    sha256 = "d658a09108bd4c4975aa6bca5372c3a7f72ddcd4abd937f9dc882b5fada57694",
    urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc0/bazeldnf-v0.5.6-rc0-linux-amd64"],
)

http_file(
    name = "bazeldnf-linux-arm64",
    executable = True,
    sha256 = "6ad9a260655bbf7591f52553aaa20436814c1426ae37b1f26e066a257a72890c",
    urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc0/bazeldnf-v0.5.6-rc0-linux-arm64"],
)

http_file(
    name = "bazeldnf-darwin-amd64",
    executable = True,
    sha256 = "487703e29bccf8536df438b0888b5d5381d1b362c3001390a0d70ff3113e8c73",
    urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc0/bazeldnf-v0.5.6-rc0-darwin-amd64"],
)

http_file(
    name = "bazeldnf-darwin-arm64",
    executable = True,
    sha256 = "261b11758afc7ce03568026691f62a78fe63ba6e87d1fc51771d3715eec44bcd",
    urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc0/bazeldnf-v0.5.6-rc0-darwin-arm64"],
)
