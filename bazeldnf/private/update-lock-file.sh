#!/usr/bin/env bash

set -euo pipefail

BAZELDNF_SHORT_PATH="${BASH_SOURCE[0]}.runfiles/@@REPO_NAME@@/@@BAZELDNF_SHORT_PATH@@"
if [ -z "${BUILD_WORKSPACE_DIRECTORY-}" ]; then
  echo "error: BUILD_WORKSPACE_DIRECTORY not set" >&2
  exit 1
fi

cd ${BUILD_WORKSPACE_DIRECTORY}

exec $BAZELDNF_SHORT_PATH lockfile @@BZLMOD_ARGS@@ "$@"
