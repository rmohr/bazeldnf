#!/usr/bin/env bash

set -euo pipefail

BAZELDNF_SHORT_PATH="${BASH_SOURCE[0]}.runfiles/@@REPO_NAME@@/@@BAZELDNF_SHORT_PATH@@"
JQ_SHORT_PATH="${BASH_SOURCE[0]}.runfiles/@@REPO_NAME@@/@@JQ_SHORT_PATH@@"
if [ -z "${BUILD_WORKSPACE_DIRECTORY-}" ]; then
  echo "error: BUILD_WORKSPACE_DIRECTORY not set" >&2
  exit 1
fi

cd ${BUILD_WORKSPACE_DIRECTORY}

if [ -z "@@BZLMOD_ARGS@@" ]; then
  exec $BAZELDNF_SHORT_PATH bzlmod $(cat @@LOCK_FILE@@ | $JQ_SHORT_PATH -r '."cli-arguments"'[]) "$@"
else
  exec $BAZELDNF_SHORT_PATH bzlmod @@BZLMOD_ARGS@@ "$@"
fi
