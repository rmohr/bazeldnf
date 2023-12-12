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
        sha256 = "7e1035d8bd2f25b787b04843f4d6a05e7cdbd3995926497c2a587e52da6262a8",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-linux-amd64"],
    )
    http_file(
        name = "bazeldnf-linux-arm64",
        executable = True,
        sha256 = "d954b785bfd79dbbd66a2f3df02b0d3a51f1fed4508a6d88fda13a9d24c878cc",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-linux-arm64"],
    )
    http_file(
        name = "bazeldnf-darwin-amd64",
        executable = True,
        sha256 = "92afc7f6475981adf9ae32b563b051869d433d8d8c9666e28a1c1c6e840394cd",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-darwin-amd64"],
    )
    http_file(
        name = "bazeldnf-darwin-arm64",
        executable = True,
        sha256 = "c5e99ed72448026ee63259a0a28440f8b43a125414fa312edbd569c2976515a3",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-darwin-arm64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64",
        executable = True,
        sha256 = "9d5337c1afe4bab858742718ac4c230d9ca368299cb97c50078eef14ae180922",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-linux-ppc64"],
    )
    http_file(
        name = "bazeldnf-linux-ppc64le",
        executable = True,
        sha256 = "7ea4db00947914bc1c51e8f042fe220a3167c65815c487eccd0c541ecfa78aa1",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/v0.5.9-rc1/bazeldnf-v0.5.9-rc1-linux-ppc64le"],
    )
