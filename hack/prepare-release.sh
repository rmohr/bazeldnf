#!/bin/bash

set -e -o pipefail -u

echo >&2 "Preparing bazeldnf ${VERSION} release"

BASE_DIR="$(
	cd "$(dirname "$BASH_SOURCE[0]")/../"
	pwd
)"

rm -rf dist
mkdir -p dist

bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_amd64 //cmd
cp -L bazel-bin/cmd/cmd_/cmd dist/bazeldnf-${VERSION}-linux-amd64
bazel build --platforms=@io_bazel_rules_go//go/toolchain:linux_arm64 //cmd
cp -L bazel-bin/cmd/cmd_/cmd dist/bazeldnf-${VERSION}-linux-arm64
bazel build --platforms=@io_bazel_rules_go//go/toolchain:darwin_amd64 //cmd
cp -L bazel-bin/cmd/cmd_/cmd dist/bazeldnf-${VERSION}-darwin-amd64
bazel build --platforms=@io_bazel_rules_go//go/toolchain:darwin_arm64 //cmd
cp -L bazel-bin/cmd/cmd_/cmd dist/bazeldnf-${VERSION}-darwin-arm64

for os in linux darwin; do
	for arch in amd64 arm64; do

		DIGEST=$(sha256sum dist/bazeldnf-${VERSION}-${os}-${arch} | cut -d " " -f 1)

		# buildozer returns a non-zero exit code (3) if the commands were a success but did not change the file.
		# To make the command idempotent, first set the digest to a kind-of-unique number to work around this behaviour.
		# This way we can assume a zero exit code if no errors occur.
		cat <<EOF | bazel run -- @com_github_bazelbuild_buildtools//buildozer -f -
set sha256 "$(date +%s)"|${BASE_DIR}/WORKSPACE:bazeldnf-${os}-${arch}
EOF

		# set the actual values
		cat <<EOF | bazel run -- @com_github_bazelbuild_buildtools//buildozer -f -
set sha256 "${DIGEST}"|${BASE_DIR}/WORKSPACE:bazeldnf-${os}-${arch}
remove urls |${BASE_DIR}/WORKSPACE:bazeldnf-${os}-${arch}
add urls "https://github.com/rmohr/bazeldnf/releases/download/${VERSION}/bazeldnf-${VERSION}-${os}-${arch}"|${BASE_DIR}/WORKSPACE:bazeldnf-${os}-${arch}
EOF

	done
done

git commit -a -m "Bump prebuilt binary references for ${VERSION}"

git archive --format tar.gz HEAD > ./dist/bazeldnf-${VERSION}.tar.gz

DIGEST=$(sha256sum dist/bazeldnf-${VERSION}.tar.gz | cut -d " " -f 1)

cat <<EOT >> ./dist/releasenote.txt
\`\`\`
http_archive(
    name = "bazeldnf",
    sha256 = "${DIGEST}",
    urls = [
        "https://github.com/rmohr/bazeldnf/releases/download/${VERSION}/bazeldnf-${VERSION}.tar.gz",
    ],
)
\`\`\`
EOT
