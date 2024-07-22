#!/usr/bin/env bash

set -o errexit -o nounset -o pipefail
set -x

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

${SCRIPT_DIR}/generate_tools_versions.sh

bazel run //tools/release
