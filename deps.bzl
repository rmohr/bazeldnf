load(
    "@bazel_tools//tools/build_defs/repo:http.bzl",
    "http_file",
)
load(
    "@bazeldnf//internal:rpm.bzl",
    _rpm = "rpm",
)
load(
    "@bazeldnf//internal:rpmtree.bzl",
    _rpmtree = "rpmtree",
)
load(
    "@bazeldnf//internal:rpmtree.bzl",
    _tar2files = "tar2files",
)
load(
    "@bazeldnf//internal:xattrs.bzl",
    _xattrs = "xattrs",
)

rpm = _rpm
rpmtree = _rpmtree
tar2files = _tar2files
xattrs = _xattrs

def bazeldnf_dependencies():
    http_file(
        name = "bazeldnf-linux-amd64",
        executable = True,
        sha256 = "70b786059514033a6b2f313abe4a51fb11ef8737ac72432a55a3219c3ad74f1c",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-linux-amd64"],
    )
    http_file(
        name = "bazeldnf-linux-arm64",
        executable = True,
        sha256 = "bdfa62ff81426bd5e0ced162b40e288c100745d3aa43204c543fc8e30cf02878",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-linux-arm64"],
    )
    http_file(
        name = "bazeldnf-darwin-amd64",
        executable = True,
        sha256 = "0d972edc1b070302673a136a6979a49ad65c5097a2caa1b5f2c09bc170118b58",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-darwin-amd64"],
    )
    http_file(
        name = "bazeldnf-darwin-arm64",
        executable = True,
        sha256 = "6329fc284361c65c47e81898a1b2c330afaa9de677127c1a924c0e0fa345f7d5",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-darwin-arm64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64",
        executable = True,
        sha256 = "c3cd79563e3cd50467480cb31ffcc31f3f99fc371c9ac7fe78468e2ac609dadd",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-linux-ppc64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64le",
        executable = True,
        sha256 = "897d287519dd0609c498a4b16c477dd10b1737721809dc1f2fdf4801e17f3e35",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.8-rc1/bazeldnf-v0.5.8-rc1-linux-ppc64le"],
    )
