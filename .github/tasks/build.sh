#!/usr/bin/env bash
set -eo pipefail

bosh create-release --name kubo \
    --sha2 \
    --tarball="kubo-release-${RELEASE_VERSION}.tgz"