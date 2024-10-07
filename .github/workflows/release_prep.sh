#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Set by GH actions, see
# https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables

# The prefix is chosen to match what GitHub generates for source archives
PREFIX="bazeldnf-${GITHUB_REF_NAME}"
ARCHIVE="bazeldnf-${GITHUB_REF_NAME}.tar.gz"
ARCHIVE_TMP=$(mktemp)

# NB: configuration for 'git archive' is in /.gitattributes
git archive --format=tar --prefix=${PREFIX}/ --worktree-attributes  ${GITHUB_REF_NAME} > $ARCHIVE_TMP

############
# Patch up the archive to have integrity hashes for built binaries that we downloaded in the GHA workflow.
# Now that we've run `git archive` we are free to pollute the working directory.

# Delete the placeholder files
tar --file $ARCHIVE_TMP --delete ${PREFIX}/tools/version.bzl
tar --file $ARCHIVE_TMP --delete ${PREFIX}/tools/integrity.bzl

mkdir -p ${PREFIX}/tools

PREFIX=$PREFIX ${SCRIPT_DIR}/generate_tools_versions.sh

INTEGRITY=$(jq \
  --from-file .github/workflows/integrity.jq \
  --arg PREFIX "bazeldnf-${GITHUB_REF_NAME}-" \
  --slurp \
  --raw-input artifacts/*.sha256 \
)

cat >${PREFIX}/tools/integrity.bzl <<EOF
"Generated during release by release_prep.sh, using integrity.jq"

INTEGRITY = ${INTEGRITY}

EOF

# Append that generated files back into the archive
tar --file $ARCHIVE_TMP --append ${PREFIX}/tools/version.bzl
tar --file $ARCHIVE_TMP --append ${PREFIX}/tools/integrity.bzl

# END patch up the archive
############

gzip < $ARCHIVE_TMP > $ARCHIVE
SHA=$(shasum -a 256 $ARCHIVE | awk '{print $1}')

# Set by GH actions, see
# https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
REPO_URL="${GITHUB_REPOSITORY:-rmohr/bazeldnf}"

cat << EOF
## Using [Bzlmod] with Bazel 6:

Add to your \`MODULE.bazel\` file:

\`\`\`starlark
bazel_dep(name = "bazeldnf", version = "${GITHUB_REF_NAME:1}")
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
    url = "https://github.com/${REPO_URL}/releases/download/${GITHUB_REF_NAME}/${ARCHIVE}",
)

load(
    "@bazeldnf//bazeldnf:repositories.bzl",
    "bazeldnf_dependencies",
)

bazeldnf_dependencies()
\`\`\`
EOF
