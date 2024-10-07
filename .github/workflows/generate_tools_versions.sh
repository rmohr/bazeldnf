#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -x

# Set by GH actions, see
# https://docs.github.com/en/actions/learn-github-actions/environment-variables#default-environment-variables
REPO_URL="${GITHUB_REPOSITORY:-rmohr/bazeldnf}"
OUT=${PREFIX:-.}/tools/version.bzl
cat > ${OUT} <<EOF
"Generated during release generate_tools_prebuilts.sh"

VERSION = "${GITHUB_REF_NAME}"

REPO_URL = "${REPO_URL}"
EOF
