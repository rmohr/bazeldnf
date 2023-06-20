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
        sha256 = "5f786c1c21edfc5c8db1613eb832269e7c11d2e6007858ef06e11367a1860f85",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-linux-amd64"],
    )
    http_file(
        name = "bazeldnf-linux-arm64",
        executable = True,
        sha256 = "798c9dc01ede8d647b0f501c15b0a79f3c201bd035f7ad9e08ca28180db60ad9",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-linux-arm64"],
    )
    http_file(
        name = "bazeldnf-darwin-amd64",
        executable = True,
        sha256 = "46f90ac9cdc397cf9040f2051dafafb25a0d838c019f563200c95c2889f15b2a",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-darwin-amd64"],
    )
    http_file(
        name = "bazeldnf-darwin-arm64",
        executable = True,
        sha256 = "d5e1eafab40e1ddc4857520db9dc99ed2410c1aa0efb2a115107ae882c74801c",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-darwin-arm64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64",
        executable = True,
        sha256 = "cb1e46c2254dfc4bccf7d8158f94a7de0f802fb5942eb682c103b128503de2dc",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-linux-ppc64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64le",
        executable = True,
        sha256 = "5b619a6333c4a96467ba38d1acbac6937aaf65f1338b7049654650670c04fab6",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.7-rc1/bazeldnf-v0.5.7-rc1-linux-ppc64le"],
    )
