#!/usr/bin/env bash

set -euo pipefail

BAZELDNF_BIN="$(realpath "${BASH_SOURCE[0]}.runfiles/@@BAZELDNF_RLOCATION@@")"
if [ -z "${BUILD_WORKSPACE_DIRECTORY-}" ]; then
  echo "error: BUILD_WORKSPACE_DIRECTORY not set" >&2
  exit 1
fi

cd "${BUILD_WORKSPACE_DIRECTORY}"

exec "$BAZELDNF_BIN" lockfile @@LOCKFILE_ARGS@@ "$@"
