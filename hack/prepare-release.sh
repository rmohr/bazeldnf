#!/bin/bash

set -e -o pipefail -u

echo >&2 "Preparing bazeldnf ${VERSION} release"

BASE_DIR="$(
	cd "$(dirname "$BASH_SOURCE[0]")/../"
	pwd
)"

rm -rf dist
mkdir -p dist

function build_arch() {
	os=$1
	arch=$2

	bazel build --platforms=@io_bazel_rules_go//go/toolchain:${os}_${arch} //cmd
	cp -L bazel-bin/cmd/cmd_/cmd dist/bazeldnf-${VERSION}-${os}-${arch}
}

function write_arch() {
	os=$1
	arch=$2

	DIGEST=$(sha256sum dist/bazeldnf-${VERSION}-${os}-${arch} | cut -d " " -f 1)
	cat <<EOT >>./deps.bzl
    http_file(
        name = "bazeldnf-${os}-${arch}",
        executable = True,
        sha256 = "${DIGEST}",
        urls = ["https://github.com/rmohr/bazeldnf/releases/download/${VERSION}/bazeldnf-${VERSION}-${os}-${arch}"],
    )
EOT

}

build_arch linux amd64
build_arch linux arm64
build_arch darwin amd64
build_arch darwin arm64
build_arch linux ppc64
build_arch linux ppc64le

cat <<EOT >./deps.bzl
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
EOT

write_arch linux amd64
write_arch linux arm64
write_arch darwin amd64
write_arch darwin arm64
write_arch linux ppc64
write_arch linux ppc64le

git commit -a -m "Bump prebuilt binary references for ${VERSION}"

git tag ${VERSION}

git archive --format tar.gz HEAD >./dist/bazeldnf-${VERSION}.tar.gz

DIGEST=$(sha256sum dist/bazeldnf-${VERSION}.tar.gz | cut -d " " -f 1)

cat <<EOT >>./dist/releasenote.txt
\`\`\`python
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "bazeldnf",
    sha256 = "${DIGEST}",
    urls = [
        "https://github.com/rmohr/bazeldnf/releases/download/${VERSION}/bazeldnf-${VERSION}.tar.gz",
    ],
)

load("@bazeldnf//:deps.bzl", "bazeldnf_dependencies")

bazeldnf_dependencies()
\`\`\`
EOT

# Only update the README if we don't build a release candidate
if [[ "${VERSION}" != *"-rc"* ]]; then

	lead='^<!-- install_start -->$'
	tail='^<!-- install_end -->$'
	sed -i -e "/$lead/,/$tail/{ /$lead/{p; r dist/releasenote.txt
        }; /$tail/p; d }" README.md

	git commit -a -m "Bump install instructions for readme in ${VERSION}"
fi
