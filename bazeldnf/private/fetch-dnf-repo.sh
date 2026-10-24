#!/usr/bin/env bash

set -euo pipefail

BAZELDNF_BIN="$(realpath "${BASH_SOURCE[0]}.runfiles/@@BAZELDNF_RLOCATION@@")"
if [ -z "${BUILD_WORKSPACE_DIRECTORY-}" ]; then
  echo "error: BUILD_WORKSPACE_DIRECTORY not set" >&2
  exit 1
fi

set -ex

cd "${BUILD_WORKSPACE_DIRECTORY}"

exec "$BAZELDNF_BIN" fetch @@CACHE_DIR@@ --repofile @@REPO_FILE@@ "$@"
