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
	cat <<EOT >>bazeldnf/deps.bzl
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
build_arch linux s390x

cat <<EOT >bazeldnf/deps.bzl
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
EOT

write_arch linux amd64
write_arch linux arm64
write_arch darwin amd64
write_arch darwin arm64
write_arch linux ppc64
write_arch linux ppc64le
write_arch linux s390x

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

load("@bazeldnf//bazeldnf:deps.bzl", "bazeldnf_dependencies")

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
