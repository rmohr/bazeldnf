#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail

# Set by GH actions, see
# https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
TAG=${GITHUB_REF_NAME}
# The prefix is chosen to match what GitHub generates for source archives
PREFIX="bazeldnf-${TAG:1}"
ARCHIVE="bazeldnf-$TAG.tar.gz"
ARCHIVE_TMP=$(mktemp)

REPO_URL="${GITHUB_REPOSITORY:-rmohr/bazeldnf}"

# NB: configuration for 'git archive' is in /.gitattributes
git archive --format=tar --prefix=${PREFIX}/ --worktree-attributes  ${TAG} > $ARCHIVE_TMP

############
# Patch up the archive to have integrity hashes for built binaries that we downloaded in the GHA workflow.
# Now that we've run `git archive` we are free to pollute the working directory.

# Delete the placeholder file
tar --file $ARCHIVE_TMP --delete ${PREFIX}/bazeldnf/private/prebuilts.bzl

# Add trailing newlines to sha256 files. They were built with
# https://github.com/aspect-build/bazel-lib/blob/main/tools/release/hashes.bzl
for sha in $(ls artifacts/*.sha256); do
  echo "" >> $sha
done

mkdir -p ${PREFIX}/tools
cat >${PREFIX}/tools/prebuilts.bzl <<EOF
"Generated during release by release_prep.sh, using integrity.jq"

VERSION = "${TAG:1}"

REPO_URL = "${REPO_URL}"

PREBUILTS = $(jq \
  --from-file .github/workflows/integrity.jq \
  --slurp \
  --raw-input artifacts/*.sha256 \
)
EOF

# Append that generated file back into the archive
tar --file $ARCHIVE_TMP --append ${PREFIX}/bazeldnf/private/prebuilts.bzl

# END patch up the archive
############

gzip < $ARCHIVE_TMP > $ARCHIVE
SHA=$(shasum -a 256 $ARCHIVE | awk '{print $1}')

cat << EOF
## Using [Bzlmod] with Bazel 6:

Add to your \`MODULE.bazel\` file:

\`\`\`starlark
bazel_dep(name = "bazeldnf", version = "${TAG:1}")
\`\`\`

This will register a prebuilt bazeldnf

[Bzlmod]: https://bazel.build/build/bzlmod

## Using WORKSPACE

\`\`\`starlark
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")
http_archive(
    name = "bazeldnf",
    sha256 = "${SHA}",
    strip_prefix = "${PREFIX}",
    url = "https://github.com/${REPO_URL}/releases/download/${TAG}/${ARCHIVE}",
)

load(
    "@bazeldnf//bazeldnf:repositories.bzl",
    "bazeldnf_dependencies",
)

bazeldnf_dependencies()
\`\`\`
EOF
