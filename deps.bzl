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
        sha256 = "d658a09108bd4c4975aa6bca5372c3a7f72ddcd4abd937f9dc882b5fada57694",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc2/bazeldnf-v0.5.6-rc2-linux-amd64"],
    )
    http_file(
        name = "bazeldnf-linux-arm64",
        executable = True,
        sha256 = "6ad9a260655bbf7591f52553aaa20436814c1426ae37b1f26e066a257a72890c",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc2/bazeldnf-v0.5.6-rc2-linux-arm64"],
    )
    http_file(
        name = "bazeldnf-darwin-amd64",
        executable = True,
        sha256 = "487703e29bccf8536df438b0888b5d5381d1b362c3001390a0d70ff3113e8c73",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc2/bazeldnf-v0.5.6-rc2-darwin-amd64"],
    )
    http_file(
        name = "bazeldnf-darwin-arm64",
        executable = True,
        sha256 = "261b11758afc7ce03568026691f62a78fe63ba6e87d1fc51771d3715eec44bcd",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.6-rc2/bazeldnf-v0.5.6-rc2-darwin-arm64"],
    )
