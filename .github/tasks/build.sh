#!/usr/bin/env bash
set -eo pipefail

RELEASE_TIMESTAMP=$(date +%s)

bosh create-release --name kubo \
    --sha2 \
    --tarball="kubo-release-${RELEASE_TIMESTAMP}.tgz"