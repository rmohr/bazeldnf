"""bazeldnf public dependency for WORKSPACE"""

load(
    "@bazel_tools//tools/build_defs/repo:http.bzl",
    "http_archive",
    "http_file",
)
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")
load(
    "@bazeldnf//internal:rpm.bzl",
    _rpm = "rpm",
)

rpm = _rpm

def bazeldnf_dependencies():
    """bazeldnf dependencies when consuming the repo externally"""
    maybe(
        http_archive,
        name = "bazel_skylib",
        sha256 = "f24ab666394232f834f74d19e2ff142b0af17466ea0c69a3f4c276ee75f6efce",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/bazel-skylib/releases/download/1.4.0/bazel-skylib-1.4.0.tar.gz",
            "https://github.com/bazelbuild/bazel-skylib/releases/download/1.4.0/bazel-skylib-1.4.0.tar.gz",
        ],
    )
    maybe(
        http_archive,
        name = "platforms",
        urls = [
            "https://mirror.bazel.build/github.com/bazelbuild/platforms/releases/download/0.0.10/platforms-0.0.10.tar.gz",
            "https://github.com/bazelbuild/platforms/releases/download/0.0.10/platforms-0.0.10.tar.gz",
        ],
        sha256 = "218efe8ee736d26a3572663b374a253c012b716d8af0c07e842e82f238a0a7ee",
    )
    http_file(
        name = "bazeldnf-linux-amd64",
        executable = True,
        sha256 = "7e1035d8bd2f25b787b04843f4d6a05e7cdbd3995926497c2a587e52da6262a8",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-linux-amd64"],
    )
    http_file(
        name = "bazeldnf-linux-arm64",
        executable = True,
        sha256 = "d954b785bfd79dbbd66a2f3df02b0d3a51f1fed4508a6d88fda13a9d24c878cc",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-linux-arm64"],
    )
    http_file(
        name = "bazeldnf-darwin-amd64",
        executable = True,
        sha256 = "92afc7f6475981adf9ae32b563b051869d433d8d8c9666e28a1c1c6e840394cd",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-darwin-amd64"],
    )
    http_file(
        name = "bazeldnf-darwin-arm64",
        executable = True,
        sha256 = "c5e99ed72448026ee63259a0a28440f8b43a125414fa312edbd569c2976515a3",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-darwin-arm64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64",
        executable = True,
        sha256 = "9d5337c1afe4bab858742718ac4c230d9ca368299cb97c50078eef14ae180922",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-linux-ppc64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64le",
        executable = True,
        sha256 = "7ea4db00947914bc1c51e8f042fe220a3167c65815c487eccd0c541ecfa78aa1",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-linux-ppc64le"],
    )
    http_file(
        name = "bazeldnf-linux-s390x",
        executable = True,
        sha256 = "09aa4abcb1d85da11642889826b982ef90547eb32099fc8177456c92f66a4cfd",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9/bazeldnf-v0.5.9-linux-s390x"],
    )
