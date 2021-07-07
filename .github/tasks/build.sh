#!/usr/bin/env bash
set -eo pipefail

: "${RELEASE_VERSION:?}"

export BOSH_LOG_LEVEL=debug
export BOSH_LOG_PATH="$PWD/bosh.log"

bosh create-release --name kubo \
    --sha2 \
    --tarball="kubo-release-${RELEASE_VERSION}.tgz"